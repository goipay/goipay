package processor

import (
	"context"
	"unsafe"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/internal/util"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func verifyBNBTxHandler(ctx context.Context, q *db.Queries, data *verifyTxHandlerData[listener.BNBTx]) (float64, error) {
	return verifyETHBasedTxHandler(ctx, q, (*verifyTxHandlerData[listener.ETHTx])(unsafe.Pointer(data)))
}

func generateNextBNBAddressHandler(ctx context.Context, q *db.Queries, data *generateNextAddressHandlerData) (db.CryptoAddress, error) {
	var addr db.CryptoAddress

	cd, err := q.FindCryptoDataByUserId(ctx, data.userId)
	if err != nil {
		return addr, err
	}

	keysAndIndices, err := q.FindKeysAndIncrementedIndicesBNBCryptoDataById(ctx, cd.BnbID)
	if err != nil {
		return addr, err
	}

	pubKey, err := deriveNextETHBasedECPubKeyHelper(indices{major: uint32(keysAndIndices.LastMajorIndex), minor: uint32(keysAndIndices.LastMinorIndex)}, keysAndIndices.MasterPubKey)
	if err != nil {
		return addr, err
	}

	addr, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: crypto.PubkeyToAddress(*pubKey.ToECDSA()).Hex(), Coin: db.CoinTypeBNB, IsOccupied: true, UserID: data.userId})
	if err != nil {
		return addr, err
	}

	return addr, nil
}

type bnbProcessor struct {
	baseCryptoProcessor[listener.BNBTx, listener.BNBBlock]
}

func newBnbProcessor(log *zerolog.Logger, dbConnPool *pgxpool.Pool, invoiceCn chan<- db.Invoice, c *dto.DaemonsConfig) (*bnbProcessor, error) {
	client, err := ethclient.Dial(c.Bnb.Url)
	if err != nil {
		return nil, err
	}

	base, err := newBaseCryptoProcessor(
		log,
		dbConnPool,
		invoiceCn,
		listener.NewSharedBNBDaemonRpcClient(client),
		verifyBNBTxHandler,
		generateNextBNBAddressHandler,
		util.GetMapKeys(tokenDataETHCompatible[db.CoinTypeBNB]),
	)
	if err != nil {
		return nil, err
	}

	return &bnbProcessor{baseCryptoProcessor: *base}, nil
}
