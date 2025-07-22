package processor

import (
	"context"
	"net/url"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type btcProcessor struct {
	baseCryptoProcessor[listener.BTCTx, listener.BTCBlock]
}

func verifyBTCTxHandler(ctx context.Context, q *db.Queries, data *verifyTxHandlerData[listener.BTCTx]) (float64, error) {
	var amount float64 = 0
	for i := 0; i < len(data.tx.Vout); i++ {
		txOut := &data.tx.Vout[i]

		if txOut.ScriptPubKey.Address == data.invoice.CryptoAddress {
			amount += data.tx.Vout[i].Value
		}
	}

	return amount, nil
}

func generateNextBTCAddressHandler(ctx context.Context, q *db.Queries, data *generateNextAddressHandlerData) (db.CryptoAddress, error) {
	var addr db.CryptoAddress

	net, err := func() (*chaincfg.Params, error) {
		switch data.network {
		case listener.MainnetBTC:
			return &chaincfg.MainNetParams, nil
		case listener.TestnetBTC:
			return &chaincfg.TestNet3Params, nil
		case listener.SignetBTC:
			return &chaincfg.SigNetParams, nil
		case listener.RegtestBTC:
			return &chaincfg.RegressionNetParams, nil
		default:
			return nil, util.InvalidNetworkTypeErr
		}
	}()
	if err != nil {
		return addr, err
	}

	cd, err := q.FindCryptoDataByUserId(ctx, data.userId)
	if err != nil {
		return addr, err
	}

	keysAndIndices, err := q.FindKeysAndIncrementedIndicesBTCCryptoDataById(ctx, cd.BtcID)
	if err != nil {
		return addr, err
	}

	mPub, err := hdkeychain.NewKeyFromString(keysAndIndices.MasterPubKey)
	if err != nil {
		return addr, err
	}

	majMPub, err := mPub.Derive(uint32(keysAndIndices.LastMajorIndex))
	if err != nil {
		return addr, err
	}
	minMPub, err := majMPub.Derive(uint32(keysAndIndices.LastMinorIndex))
	if err != nil {
		return addr, err
	}

	pubKey, err := minMPub.ECPubKey()
	if err != nil {
		return addr, err
	}

	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	newAddr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, net)
	if err != nil {
		return addr, err
	}

	addr, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: newAddr.EncodeAddress(), Coin: db.CoinTypeBTC, IsOccupied: true, UserID: data.userId})
	if err != nil {
		return addr, err
	}

	return addr, nil
}

func newBtcProcessor(log *zerolog.Logger, dbConnPool *pgxpool.Pool, invoiceCn chan<- db.Invoice, c *dto.DaemonsConfig) (*btcProcessor, error) {
	u, err := url.Parse(c.Btc.Url)
	if err != nil {
		return nil, err
	}

	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         u.Host + u.RequestURI(),
		User:         c.Btc.User,
		Pass:         c.Btc.Pass,
		DisableTLS:   u.Scheme != "https",
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		return nil, err
	}

	base, err := newBaseCryptoProcessor(
		log,
		dbConnPool,
		invoiceCn,
		listener.NewSharedBTCDaemonRpcClient(client),
		verifyBTCTxHandler,
		generateNextBTCAddressHandler,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &btcProcessor{baseCryptoProcessor: *base}, nil
}
