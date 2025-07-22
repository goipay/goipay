package processor

import (
	"context"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/internal/util"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type ethProcessor struct {
	baseCryptoProcessor[listener.ETHTx, listener.ETHBlock]
}

func generateNextETHAddressHandler(ctx context.Context, q *db.Queries, data *generateNextAddressHandlerData) (db.CryptoAddress, error) {
	var addr db.CryptoAddress

	cd, err := q.FindCryptoDataByUserId(ctx, data.userId)
	if err != nil {
		return addr, err
	}

	keysAndIndices, err := q.FindKeysAndIncrementedIndicesETHCryptoDataById(ctx, cd.EthID)
	if err != nil {
		return addr, err
	}

	pubKey, err := deriveNextETHBasedECPubKeyHelper(indices{major: uint32(keysAndIndices.LastMajorIndex), minor: uint32(keysAndIndices.LastMinorIndex)}, keysAndIndices.MasterPubKey)
	if err != nil {
		return addr, err
	}

	addr, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: crypto.PubkeyToAddress(*pubKey.ToECDSA()).Hex(), Coin: db.CoinTypeETH, IsOccupied: true, UserID: data.userId})
	if err != nil {
		return addr, err
	}

	return addr, nil
}

func newEthProcessor(log *zerolog.Logger, dbConnPool *pgxpool.Pool, invoiceCn chan<- db.Invoice, c *dto.DaemonsConfig) (*ethProcessor, error) {
	client, err := ethclient.Dial(c.Eth.Url)
	if err != nil {
		return nil, err
	}

	base, err := newBaseCryptoProcessor(
		log,
		dbConnPool,
		invoiceCn,
		listener.NewSharedETHDaemonRpcClient(client),
		verifyETHBasedTxHandler,
		generateNextETHAddressHandler,
		util.GetMapKeys(tokenDataETHCompatible[db.CoinTypeETH]),
	)
	if err != nil {
		return nil, err
	}

	return &ethProcessor{baseCryptoProcessor: *base}, nil
}
