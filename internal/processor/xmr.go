package processor

import (
	"context"
	"errors"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/chekist32/go-monero/daemon"
	"github.com/chekist32/go-monero/utils"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type pendingInvoice struct {
	invoice           *atomic.Pointer[db.Invoice]
	cancelTimeoutFunc context.CancelFunc
}

type incomingMoneroTx interface {
	txInfo() daemon.MoneroTxInfo
	confirmations() uint64
	doubleSpendSeen() bool
	txId() string
}

type incomingMoneroTxTxPool daemon.MoneroTx

func (i incomingMoneroTxTxPool) txInfo() daemon.MoneroTxInfo {
	return i.TxInfo
}
func (i incomingMoneroTxTxPool) confirmations() uint64 {
	return 0
}
func (i incomingMoneroTxTxPool) doubleSpendSeen() bool {
	return i.DoubleSpendSeen
}
func (i incomingMoneroTxTxPool) txId() string {
	return i.IdHash
}

type incomingMoneroTxGetTx daemon.MoneroTx1

func (i incomingMoneroTxGetTx) txInfo() daemon.MoneroTxInfo {
	return i.TxInfo
}
func (i incomingMoneroTxGetTx) confirmations() uint64 {
	return 0
}
func (i incomingMoneroTxGetTx) doubleSpendSeen() bool {
	return i.DoubleSpendSeen
}
func (i incomingMoneroTxGetTx) txId() string {
	return i.TxHash
}

type xmrProcessor struct {
	daemon   daemon.IDaemonRpcClient
	daemonEx *listener.DaemonRpcClientExecutor
	network  utils.NetworkType

	baseCryptoProcessor
}

func (p *xmrProcessor) verifyMoneroTxOnTxMempool(ctx context.Context, xmrTx incomingMoneroTx) {
	p.pendingInvoices.Range(func(key string, value pendingInvoice) bool {
		go func() {
			txInfo := xmrTx.txInfo()

			q, tx, err := util.InitDbQueriesWithTx(ctx, p.dbConnPool)
			if err != nil {
				p.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
				return
			}
			defer tx.Rollback(ctx)

			invoice := value.invoice.Load()

			cryptoData, err := q.FindCryptoDataByUserId(ctx, invoice.UserID)
			if err != nil {
				p.log.Err(err).Str("queryName", "FindCryptoDataByUserId").Msg(util.DefaultFailedSqlQueryMsg)
				return
			}

			xmrKeys, err := q.FindKeysAndLockXMRCryptoDataById(ctx, cryptoData.XmrID)
			if err != nil {
				p.log.Err(err).Str("queryName", "FindKeysAndLockXMRCryptoDataById").Msg(util.DefaultFailedSqlQueryMsg)
				return
			}

			privView, err := utils.NewPrivateKey(xmrKeys.PrivViewKey)
			if err != nil {
				p.log.Err(err).Msg("An error occurred while creating the XMR private view key.")
				return
			}

			addr, err := utils.NewAddress(invoice.CryptoAddress)
			if err != nil {
				p.log.Err(err).Msg("An error occurred while generating a new XMR subaddress.")
				return
			}
			pubSpend := addr.PublicSpendKey()

			txPub, err := utils.GetTxPublicKeyFromExtra(txInfo.Extra)
			if err != nil {
				p.log.Err(err).Msg("An error occurred while extracting the tx public key from the extra field.")
				return
			}

			for i := 0; i < len(txInfo.Vout); i++ {
				select {
				case <-ctx.Done():
					return
				default:
					out := &txInfo.Vout[i]
					ecdh := &txInfo.RctSignatures.EcdhInfo[i]

					outKey, err := utils.NewPublicKey(out.Target.TaggedKey.Key)
					if err != nil {
						p.log.Err(err).Msg("An error occurred while creating the XMR tx outKey.")
						return
					}

					res, am, err := utils.DecryptOutputPublicSpendKey(pubSpend, uint32(i), outKey, ecdh.Amount, txPub, privView)
					if err != nil {
						p.log.Err(err).Msg("An error occurred while decrypting the XMR tx output.")
						return
					}
					if !res ||
						xmrTx.doubleSpendSeen() ||
						value.invoice.Load().RequiredAmount > utils.XMRToFloat64(am) {
						continue
					}

					var txId pgtype.Text
					if err := txId.Scan(xmrTx.txId()); err != nil {
						p.log.Err(err).Str("fieldName", "txId").Msg(util.DefaultFailedScanningToPostgresqlDataTypeMsg)
						return
					}

					var amount pgtype.Float8
					if err := amount.Scan(utils.XMRToFloat64(am)); err != nil {
						p.log.Err(err).Str("fieldName", "amount").Msg(util.DefaultFailedScanningToPostgresqlDataTypeMsg)
						return
					}

					invoice, err := q.ConfirmInvoiceStatusMempoolById(ctx, db.ConfirmInvoiceStatusMempoolByIdParams{ID: value.invoice.Load().ID, ActualAmount: amount, TxID: txId})
					if err != nil {
						p.log.Err(err).Str("queryName", "ConfirmInvoiceStatusMempoolById").Msg(util.DefaultFailedSqlQueryMsg)
						return
					}

					value.invoice.Store(&invoice)
				}
			}

			tx.Commit(ctx)

			invoice = value.invoice.Load()
			if invoice.Status != db.InvoiceStatusTypePENDING {
				p.broadcastUpdatedInvoice(ctx, invoice)
				p.confirmInvoiceHelper(ctx, value)
			}
		}()

		return true
	})
}

func (p *xmrProcessor) confirmInvoiceHelper(ctx context.Context, value pendingInvoice) {
	invoice := value.invoice.Load()
	if !invoice.TxID.Valid {
		return
	}

	xmrTx, err := p.daemon.GetTransactions([]string{invoice.TxID.String}, true, false, false)
	if err != nil {
		p.log.Err(err).Str("method", "get_transactions").Msg(util.DefaultFailedFetchingXMRDaemonMsg)
		return
	}
	if len(xmrTx.MissedTx) > 0 {
		p.log.Info().Msgf("Tx %v was rejected by blockchain", invoice.TxID.String)
		p.expireInvoice(ctx, invoice)
		return
	}

	if uint64(invoice.ConfirmationsRequired) > xmrTx.Txs[0].Confirmations {
		return
	}
	if _, loaded := p.pendingInvoices.LoadAndDelete(invoice.CryptoAddress); !loaded {
		return
	}
	value.cancelTimeoutFunc()

	confirmedInvoice, err := p.confirmInvoice(ctx, invoice)
	if err != nil {
		return
	}

	go p.releaseAddressHelper(ctx, invoice)
	p.broadcastUpdatedInvoice(ctx, confirmedInvoice)
}

func (p *xmrProcessor) verifyMoneroTxOnNewBlock(ctx context.Context) {
	p.pendingInvoices.Range(func(key string, value pendingInvoice) bool {
		go p.confirmInvoiceHelper(ctx, value)
		return true
	})
}

func (p *xmrProcessor) persistCryptoCache(ctx context.Context) {
	lastHeight := int64(p.daemonEx.LastSyncedBlockHeight())
	p.persistCryptoCacheHelper(ctx, db.CoinTypeXMR, lastHeight)
}

func (p *xmrProcessor) load(ctx context.Context) error {
	q, tx, err := util.InitDbQueriesWithTx(ctx, p.dbConnPool)
	if err != nil {
		p.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return err
	}
	defer tx.Rollback(ctx)

	cache, err := q.FindCryptoCacheByCoin(ctx, db.CoinTypeXMR)
	if err != nil {
		p.log.Err(err).Str("queryName", "FindCryptoCacheByCoin").Msg(util.DefaultFailedSqlQueryMsg)
		return err
	}

	res, err := p.daemon.GetLastBlockHeader(false)
	if err != nil {
		p.log.Err(err).Str("method", "get_last_block_header").Msg(util.DefaultFailedFetchingXMRDaemonMsg)
		return err
	}

	height := int64(res.Result.BlockHeader.Height)
	if cache.LastSyncedBlockHeight.Valid {
		height = cache.LastSyncedBlockHeight.Int64
	}

	tx.Commit(ctx)

	go func() {
		blockCn := p.daemonEx.NewBlockChan()

		for {
			select {
			case res := <-blockCn:
				go func() {
					txsRes, err := p.daemon.GetTransactions(res.BlockDetails.TxHashes, true, false, false)
					if err != nil {
						p.log.Err(err).Str("method", "get_transactions").Msg(util.DefaultFailedFetchingXMRDaemonMsg)
						return
					}

					for i := 0; i < len(txsRes.Txs); i++ {
						go p.verifyMoneroTxOnTxMempool(ctx, incomingMoneroTxGetTx(txsRes.Txs[i]))
					}

				}()

				go p.verifyMoneroTxOnNewBlock(ctx)

			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		txPoolCn := p.daemonEx.NewTxPoolChan()

		for {
			select {
			case res := <-txPoolCn:
				go p.verifyMoneroTxOnTxMempool(ctx, incomingMoneroTxTxPool(res))
			case <-ctx.Done():
				return
			}
		}
	}()

	p.daemonEx.Start(uint64(height))

	go func() {
		p.persistCryptoCache(ctx)

		for {
			select {
			case <-time.After(persist_cache_timeout):
				go p.persistCryptoCache(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (p *xmrProcessor) generateNextXmrAddressHelper(ctx context.Context, q *db.Queries, userId pgtype.UUID) (db.CryptoAddress, error) {
	var addr db.CryptoAddress

	cd, err := q.FindCryptoDataByUserId(ctx, userId)
	if err != nil {
		return addr, err
	}

	indices, err := q.FindIndicesAndLockXMRCryptoDataById(ctx, cd.XmrID)
	if err != nil {
		return addr, err
	}

	keys, err := q.FindKeysAndLockXMRCryptoDataById(ctx, cd.XmrID)
	if err != nil {
		return addr, err
	}

	viewKey, err := utils.NewPrivateKey(keys.PrivViewKey)
	if err != nil {
		return addr, err
	}

	spendKey, err := utils.NewPublicKey(keys.PubSpendKey)
	if err != nil {
		return addr, err
	}

	indices.LastMinorIndex++
	if indices.LastMinorIndex <= 0 {
		indices.LastMinorIndex = 0
		indices.LastMajorIndex++
	}

	subAddr, err := utils.GenerateSubaddress(viewKey, spendKey, uint32(indices.LastMajorIndex), uint32(indices.LastMinorIndex), p.network)
	if err != nil {
		return addr, err
	}

	addr, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: subAddr.Address(), Coin: db.CoinTypeXMR, IsOccupied: true, UserID: userId})
	if err != nil {
		return addr, err
	}

	if _, err := q.UpdateIndicesXMRCryptoDataById(ctx, db.UpdateIndicesXMRCryptoDataByIdParams{ID: cd.XmrID, LastMajorIndex: indices.LastMajorIndex, LastMinorIndex: indices.LastMinorIndex}); err != nil {
		return addr, err
	}

	return addr, nil
}

func (p *xmrProcessor) createInvoice(ctx context.Context, req *dto.NewInvoiceRequest) (*db.Invoice, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, p.dbConnPool)
	if err != nil {
		p.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, err
	}
	defer tx.Rollback(ctx)

	var userId pgtype.UUID
	if err := userId.Scan(req.UserId); err != nil {
		return nil, err
	}

	coin := req.Coin

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout < listener.MIN_SYNC_TIMEOUT {
		timeout = listener.MIN_SYNC_TIMEOUT
	}

	var expiresAt pgtype.Timestamptz
	if err := expiresAt.Scan(time.Now().UTC().Add(timeout)); err != nil {
		return nil, err
	}

	addr, err := q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: userId, Coin: coin})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		addr, err = p.generateNextXmrAddressHelper(ctx, q, userId)
		if err != nil {
			return nil, err
		}
	}

	invoice, err := q.CreateInvoice(
		ctx,
		db.CreateInvoiceParams{
			CryptoAddress:         addr.Address,
			Coin:                  coin,
			RequiredAmount:        req.Amount,
			ConfirmationsRequired: int16(req.Confirmations),
			ExpiresAt:             expiresAt,
			UserID:                userId,
		},
	)
	if err != nil {
		p.log.Err(err).Str("queryName", "CreateInvoice").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, err
	}

	tx.Commit(ctx)

	return &invoice, nil
}

func (p *xmrProcessor) handleInvoicePbReq(ctx context.Context, req *dto.NewInvoiceRequest) (*db.Invoice, error) {
	invoice, err := p.createInvoice(ctx, req)
	if err != nil {
		return nil, err
	}

	p.handleInvoice(ctx, *invoice)
	p.broadcastUpdatedInvoice(ctx, invoice)

	return invoice, nil
}

func newXmrProcessor(dbConnPool *pgxpool.Pool, invoiceCn chan<- db.Invoice, c *dto.DaemonsConfig, log *zerolog.Logger) (*xmrProcessor, error) {
	u, err := url.Parse(c.Xmr.Url)
	if err != nil {
		return nil, err
	}

	d := daemon.NewDaemonRpcClient(daemon.NewRpcConnection(u, c.Xmr.User, c.Xmr.Pass))

	res, err := d.GetInfo()
	if err != nil {
		return nil, err
	}

	net := utils.Mainnet
	if res.Result.Stagenet {
		net = utils.Stagenet
	} else if res.Result.Testnet {
		net = utils.Testnet
	}

	return &xmrProcessor{
			baseCryptoProcessor: baseCryptoProcessor{
				log:             log,
				dbConnPool:      dbConnPool,
				invoiceCn:       invoiceCn,
				pendingInvoices: new(util.SyncMapTypeSafe[string, pendingInvoice]),
			},
			daemon:   d,
			daemonEx: listener.NewDaemonRpcClientExecutor(d, log),
			network:  net,
		},
		nil
}
