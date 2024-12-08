package processor

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type cryptoProcessor interface {
	persistCryptoCache(ctx context.Context)
	load(ctx context.Context) error
	createInvoice(ctx context.Context, req *dto.NewInvoiceRequest)
	handleInvoicePbReq(ctx context.Context, req *dto.NewInvoiceRequest) (*db.Invoice, error)
	expireInvoice(ctx context.Context, invoice *db.Invoice)
}

type baseCryptoProcessor struct {
	log *zerolog.Logger

	dbConnPool *pgxpool.Pool

	invoiceCn chan<- db.Invoice

	pendingInvoices *util.SyncMapTypeSafe[string, pendingInvoice]
}

func (b *baseCryptoProcessor) releaseAddressHelper(ctx context.Context, invoice *db.Invoice) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, b.dbConnPool)
	if err != nil {
		b.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return
	}
	defer tx.Rollback(ctx)

	if _, err := q.UpdateIsOccupiedByCryptoAddress(ctx, db.UpdateIsOccupiedByCryptoAddressParams{IsOccupied: false, Address: invoice.CryptoAddress}); err != nil {
		b.log.Err(err).Str("queryName", "UpdateIsOccupiedByCryptoAddress").Msg(util.DefaultFailedSqlQueryMsg)
		return
	}

	tx.Commit(ctx)
}

func (b *baseCryptoProcessor) broadcastUpdatedInvoice(ctx context.Context, invoice *db.Invoice) {
	go func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, util.SEND_TIMEOUT)
		defer cancel()

		select {
		case b.invoiceCn <- *invoice:
			b.log.Debug().Str("invoiceId", util.PgUUIDToString(invoice.ID)).Msg("Invoice broadcasted")
			return
		case <-timeoutCtx.Done():
			b.log.Debug().Str("invoiceId", util.PgUUIDToString(invoice.ID)).Msg("Timeout expired")
			return
		}
	}()
}

func (b *baseCryptoProcessor) expireInvoice(ctx context.Context, invoice *db.Invoice) {
	if _, loaded := b.pendingInvoices.LoadAndDelete(invoice.CryptoAddress); !loaded {
		return
	}

	q, tx, err := util.InitDbQueriesWithTx(ctx, b.dbConnPool)
	if err != nil {
		b.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return
	}
	defer tx.Rollback(ctx)

	expiredInvoice, err := q.ExpireInvoiceById(ctx, invoice.ID)
	if err != nil {
		b.log.Err(err).Str("queryName", "ExpireInvoiceById").Msg(util.DefaultFailedSqlQueryMsg)
		return
	}

	go b.releaseAddressHelper(ctx, invoice)
	b.broadcastUpdatedInvoice(ctx, &expiredInvoice)

	tx.Commit(ctx)
}

func (b *baseCryptoProcessor) persistCryptoCacheHelper(ctx context.Context, coin db.CoinType, lastBlockHeight int64) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, b.dbConnPool)
	if err != nil {
		b.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return
	}
	defer tx.Rollback(ctx)

	var height pgtype.Int8
	if err := height.Scan(lastBlockHeight); err != nil {
		b.log.Err(err).Str("fieldName", "height").Msg(util.DefaultFailedScanningToPostgresqlDataTypeMsg)
		return
	}

	if _, err := q.UpdateCryptoCacheByCoin(ctx, db.UpdateCryptoCacheByCoinParams{Coin: coin, LastSyncedBlockHeight: height}); err != nil {
		b.log.Err(err).Str("queryName", "UpdateCryptoCacheByCoin").Msg(util.DefaultFailedSqlQueryMsg)
		return
	}

	tx.Commit(ctx)
}

func (b *baseCryptoProcessor) handleInvoiceHelper(confirmedInvoiceCtx context.Context, invoice *db.Invoice) {
	select {
	case <-time.After(invoice.ExpiresAt.Time.Sub(time.Now().UTC())):
		b.expireInvoice(confirmedInvoiceCtx, invoice)
		return
	case <-confirmedInvoiceCtx.Done():
		return
	}
}
func (b *baseCryptoProcessor) handleInvoice(ctx context.Context, invoice db.Invoice) {
	if _, ok := b.pendingInvoices.Load(invoice.CryptoAddress); ok {
		return
	}

	confirmedInvoiceCtx, cancel := context.WithCancel(ctx)

	invoicePtr := &atomic.Pointer[db.Invoice]{}
	invoicePtr.Store(&invoice)
	b.pendingInvoices.Store(invoice.CryptoAddress, pendingInvoice{invoice: invoicePtr, cancelTimeoutFunc: cancel})

	go b.handleInvoiceHelper(confirmedInvoiceCtx, &invoice)
}

func (b *baseCryptoProcessor) confirmInvoice(ctx context.Context, invoice *db.Invoice) (*db.Invoice, error) {
	q, tx, err := util.InitDbQueriesWithTx(ctx, b.dbConnPool)
	if err != nil {
		b.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return nil, err
	}
	defer tx.Rollback(ctx)

	confirmedInvoice, err := q.ConfirmInvoiceById(ctx, invoice.ID)
	if err != nil {
		b.log.Err(err).Str("queryName", "ConfirmInvoiceById").Msg(util.DefaultFailedSqlQueryMsg)
		return nil, err
	}

	tx.Commit(ctx)

	return &confirmedInvoice, nil
}
