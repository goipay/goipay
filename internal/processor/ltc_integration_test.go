package processor

import (
	"context"
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/test"
	test_db "github.com/chekist32/goipay/test/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	ltcrpc "github.com/ltcsuite/ltcd/rpcclient"
	"github.com/stretchr/testify/assert"
)

func createUserWithLtcData(ctx context.Context, q *db.Queries) (pgtype.UUID, db.CryptoDatum, db.LtcCryptoDatum) {
	userId, err := q.CreateUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	ltcData, err := q.CreateLTCCryptoData(ctx, "zpub6o5L7tQbC4zavTL1Lzq1eg5qev4WQMXNWMGoruoHSX8YRss8V4U1k4UUae8abXpVxNh9eBHTLBGBjvuCSRtfVtAmf4LRBtsNxQX4gpj56Dc")
	if err != nil {
		log.Fatal(err)
	}
	cd, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{LtcID: ltcData.ID, UserID: userId})
	if err != nil {
		log.Fatal(err)
	}

	return userId, cd, ltcData
}

func createNewTestLtcDaemon1() *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         "api.chainup.net/litecoin/mainnet/0b1abdf17ecc4b20b110ee73e17e7493",
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true,
		DisableTLS:   false,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func createNewTestLtcDaemon2() *ltcrpc.Client {
	connCfg := &ltcrpc.ConnConfig{
		Host:         "api.chainup.net/litecoin/mainnet/0b1abdf17ecc4b20b110ee73e17e7493",
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true,
		DisableTLS:   false,
	}
	client, err := ltcrpc.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func TestGenerateNextLtcAddressHandler(t *testing.T) {
	t.Parallel()

	data := []struct {
		prevMajorIndex int32
		prevMinorIndex int32
		expectedAddr   string
	}{
		{
			prevMajorIndex: 0,
			prevMinorIndex: 0,
			expectedAddr:   "ltc1quc3y9flfnhd9zeg5dm5h4ckzz040ykk495rstl",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: 124,
			expectedAddr:   "ltc1qu0a864puuay8hsz9sg5sjl92d5qsxlp9w6740v",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: math.MaxInt32,
			expectedAddr:   "ltc1quhh8zju3ah6hcjq624nvudfd0sevn9g3u80y2r",
		},
		{
			prevMajorIndex: 1,
			prevMinorIndex: 2,
			expectedAddr:   "ltc1qqnla6k30dly8vj370jqmx0ytynwrc92uxvxd45",
		},
	}

	ctx := context.Background()

	dbConn, _, close := getPostgresWithDbConn()
	defer close(ctx)

	for _, d := range data {
		t.Run(fmt.Sprintf("Should Return Valid Address Ma %v Mi %v", d.prevMajorIndex, d.prevMinorIndex), func(t *testing.T) {
			test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
				// Given
				q := db.New(dbConn).WithTx(tx)
				qT := test_db.New(dbConn).WithTx(tx)
				userId, cd, _ := createUserWithLtcData(ctx, q)

				_, err := qT.UpdateIndicesLTCCryptoDataById(ctx, test_db.UpdateIndicesLTCCryptoDataByIdParams{ID: cd.LtcID, LastMajorIndex: d.prevMajorIndex, LastMinorIndex: d.prevMinorIndex})
				if err != nil {
					log.Fatal(err)
				}

				// When
				addr, err := generateNextLTCAddressHandler(ctx, q, &generateNextAddressHandlerData{userId: userId, network: listener.MainnetLTC})

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, d.expectedAddr, addr.Address)
			})
		})
	}

}

func TestVerifyLTCTxHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConn, _, close := getPostgresWithDbConn()
	defer close(ctx)

	daemon := listener.NewSharedLTCDaemonRpcClient(createNewTestLtcDaemon1(), createNewTestLtcDaemon2())

	t.Run("Should Return Right Amount (Valid Tx)", func(t *testing.T) {
		test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
			// Given
			q := db.New(dbConn).WithTx(tx)
			userId, _, _ := createUserWithXmrData(ctx, q)

			expectedTxId := "4c699b97e516791e9189211af52f0f18fb24d71e86fb73530dfc9fc1fb00fc33"
			expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
				UserID:                userId,
				Coin:                  db.CoinTypeLTC,
				CryptoAddress:         "ltc1qa9fetyxs65t03w32vfyen4w2nph9uq9wr7pmg4",
				RequiredAmount:        3.05398200,
				ExpiresAt:             pgtype.Timestamptz{Time: time.Now().Add(time.Minute), Valid: true},
				ConfirmationsRequired: 0,
			})
			if err != nil {
				log.Fatal(err)
			}

			txs, err := daemon.GetTransactions([]string{expectedTxId})
			if err != nil {
				log.Fatal(err)
			} else if len(txs) < 1 {
				log.Fatalf("Invalid Tx Id: %v", expectedTxId)
			}

			// When
			amount, err := verifyLTCTxHandler(ctx, q, &verifyTxHandlerData[listener.LTCTx]{invoice: expectedInvoice, tx: txs[0]})

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedInvoice.RequiredAmount, amount)
		})
	})

	t.Run("Should Return 0 Amount (Invalid Tx)", func(t *testing.T) {
		test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
			// Given
			q := db.New(dbConn).WithTx(tx)
			userId, _, _ := createUserWithLtcData(ctx, q)

			expectedTxId := "16132b08000b8af8e4ff5e63d1b10d2cf730ace428c778b4da44080b4ccc58a8"
			expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
				UserID:                userId,
				Coin:                  db.CoinTypeLTC,
				CryptoAddress:         "ltc1qa9fetyxs65t03w32vfyen4w2nph9uq9wr7pmg4",
				RequiredAmount:        3.05398200,
				ExpiresAt:             pgtype.Timestamptz{Time: time.Now().Add(time.Minute), Valid: true},
				ConfirmationsRequired: 0,
			})
			if err != nil {
				log.Fatal(err)
			}

			txs, err := daemon.GetTransactions([]string{expectedTxId})
			if err != nil {
				log.Fatal(err)
			} else if len(txs) < 1 {
				log.Fatalf("Invalid Tx Id: %v", expectedTxId)
			}

			// When
			amount, err := verifyLTCTxHandler(ctx, q, &verifyTxHandlerData[listener.LTCTx]{invoice: expectedInvoice, tx: txs[0]})

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, float64(0), amount)
		})
	})
}
