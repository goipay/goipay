package processor

import (
	"context"
	"errors"
	"net/url"

	"github.com/chekist32/go-monero/daemon"
	"github.com/chekist32/go-monero/utils"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type xmrProcessor struct {
	baseCryptoProcessor[listener.XMRTx, listener.XMRBlock]
}

func verifyXMRTxHandler(ctx context.Context, q *db.Queries, data *verifyTxHandlerData[listener.XMRTx]) (float64, error) {
	cryptoData, err := q.FindCryptoDataByUserId(ctx, data.invoice.UserID)
	if err != nil {
		return 0, err
	}

	xmrKeys, err := q.FindKeysXMRCryptoDataById(ctx, cryptoData.XmrID)
	if err != nil {
		return 0, err
	}

	privView, err := utils.NewPrivateKey(xmrKeys.PrivViewKey)
	if err != nil {
		return 0, errors.New("error occurred while creating the XMR private view key")
	}

	addr, err := utils.NewAddress(data.invoice.CryptoAddress)
	if err != nil {
		return 0, errors.New("error occurred while generating a new XMR subaddress")
	}
	pubSpend := addr.PublicSpendKey()

	txPub, err := utils.GetTxPublicKeyFromExtra(data.tx.TxInfo.Extra)
	if err != nil {
		return 0, errors.New("error occurred while extracting the tx public key from the extra field")
	}

	var amount uint64 = 0
	for i := 0; i < len(data.tx.TxInfo.Vout); i++ {
		out := &data.tx.TxInfo.Vout[i]
		ecdh := &data.tx.TxInfo.RctSignatures.EcdhInfo[i]

		outKey, err := utils.NewPublicKey(out.Target.TaggedKey.Key)
		if err != nil {
			return 0, errors.New("error occurred while creating the XMR tx outKey")
		}

		res, am, err := utils.DecryptOutputPublicSpendKey(pubSpend, uint32(i), outKey, ecdh.Amount, txPub, privView)
		if err != nil {
			return 0, errors.New("error occurred while decrypting the XMR tx output")
		}
		if res {
			amount += am
		}
	}

	return utils.XMRToFloat64(amount), nil
}

func generateNextXMRAddressHandler(ctx context.Context, q *db.Queries, data *generateNextAddressHandlerData) (db.CryptoAddress, error) {
	var addr db.CryptoAddress

	net, err := func() (utils.NetworkType, error) {
		switch data.network {
		case listener.MainnetXMR:
			return utils.Mainnet, nil
		case listener.StagenetXMR:
			return utils.Stagenet, nil
		case listener.TestnetXMR:
			return utils.Testnet, nil
		default:
			return 255, errors.New("invalid XMR network type")
		}
	}()
	if err != nil {
		return addr, err
	}

	cd, err := q.FindCryptoDataByUserId(ctx, data.userId)
	if err != nil {
		return addr, err
	}

	keysIndicesData, err := q.FindKeysAndIncrementedIndicesXMRCryptoDataById(ctx, cd.XmrID)
	if err != nil {
		return addr, err
	}

	viewKey, err := utils.NewPrivateKey(keysIndicesData.PrivViewKey)
	if err != nil {
		return addr, err
	}

	spendKey, err := utils.NewPublicKey(keysIndicesData.PubSpendKey)
	if err != nil {
		return addr, err
	}

	subAddr, err := utils.GenerateSubaddress(viewKey, spendKey, uint32(keysIndicesData.LastMajorIndex), uint32(keysIndicesData.LastMinorIndex), net)
	if err != nil {
		return addr, err
	}

	addr, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: subAddr.Address(), Coin: db.CoinTypeXMR, IsOccupied: true, UserID: data.userId})
	if err != nil {
		return addr, err
	}

	return addr, nil
}

func newXmrProcessor(log *zerolog.Logger, dbConnPool *pgxpool.Pool, invoiceCn chan<- db.Invoice, c *dto.DaemonsConfig) (*xmrProcessor, error) {
	u, err := url.Parse(c.Xmr.Url)
	if err != nil {
		return nil, err
	}

	base, err := newBaseCryptoProcessor(
		log,
		dbConnPool,
		invoiceCn,
		listener.NewSharedXMRDaemonRpcClient(daemon.NewDaemonRpcClient(daemon.NewRpcConnection(u, c.Xmr.User, c.Xmr.Pass))),
		verifyXMRTxHandler,
		generateNextXMRAddressHandler,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &xmrProcessor{baseCryptoProcessor: *base}, nil
}
