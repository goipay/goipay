package processor

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/chekist32/go-monero/daemon"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/test"
	test_db "github.com/chekist32/goipay/test/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func createUserWithXmrData(ctx context.Context, q *db.Queries) (pgtype.UUID, db.CryptoDatum, db.XmrCryptoDatum) {
	userId, err := q.CreateUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	xmrData, err := q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: "8aa763d1c8d9da4ca75cb6ca22a021b5cca376c1367be8d62bcc9cdf4b926009", PubSpendKey: "38e9908d33d034de0ba1281aa7afe3907b795cea14852b3d8fe276e8931cb130"})
	if err != nil {
		log.Fatal(err)
	}
	cd, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmrData.ID, UserID: userId})
	if err != nil {
		log.Fatal(err)
	}

	return userId, cd, xmrData
}

func createNewTestXMRDaemon() daemon.IDaemonRpcClient {
	u, err := url.Parse("http://node.monerodevs.org:38089")
	if err != nil {
		log.Fatal(err)
	}

	return daemon.NewDaemonRpcClient(daemon.NewRpcConnection(u, "", ""))
}

func getPostgresWithDbConn() (*pgxpool.Pool, testcontainers.Container, func(ctx context.Context)) {
	return test.SpinUpPostgresContainerAndGetPgxpool(fmt.Sprintf("%v/../../sql/migrations", os.Getenv("PWD")))
}

func TestGenerateNextXmrAddressHandler(t *testing.T) {
	t.Parallel()

	data := []struct {
		prevMajorIndex int32
		prevMinorIndex int32
		expectedAddr   string
	}{
		{
			prevMajorIndex: 0,
			prevMinorIndex: 0,
			expectedAddr:   "74xhb5sXRsnDZv8RKFEv7LAMfUq5AmGEEB77SVvsUJf8bLvFMSEfc8YYyJHF6xNNnjAZQmgqZp76AjT8bD6qKkLZLeR42oi",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: 124,
			expectedAddr:   "72KK86oj9H4AMSwaeisZLjA5FLEucwSXaQs7ncvkwRLV3wNCoLa81cQWo2tSwHp68hhLP2oPSipLGNtCPx1ojdqA4HyKgbC",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: math.MaxInt32,
			expectedAddr:   "72c2F4L6XMu28Wf4e5yiVfKJcb4uDzvM9DxSAydF9o766RUiVqXawkhUcz7y59EBRrDafZB8DezLbLSrtb5xPL7s6PZ2zoj",
		},
		{
			prevMajorIndex: 1,
			prevMinorIndex: 2,
			expectedAddr:   "77QbnheAWq19ZU5fYV3ERJF1zCtd1wbV2W5DUJnaXt91cxTwZosvfqBi4Uknd53EABG9TYhWPZqxhN5HoBbamZjBC2EpWrb",
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
				userId, cd, _ := createUserWithXmrData(ctx, q)

				_, err := qT.UpdateIndicesXMRCryptoDataById(ctx, test_db.UpdateIndicesXMRCryptoDataByIdParams{ID: cd.XmrID, LastMajorIndex: d.prevMajorIndex, LastMinorIndex: d.prevMinorIndex})
				if err != nil {
					log.Fatal(err)
				}

				// When
				addr, err := generateNextXMRAddressHandler(ctx, q, &generateNextAddressHandlerData{userId: userId, network: listener.StagenetXMR})

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, d.expectedAddr, addr.Address)
			})
		})
	}

}

func TestVerifyXMRTxHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConn, _, close := getPostgresWithDbConn()
	defer close(ctx)

	daemon := listener.NewSharedXMRDaemonRpcClient(createNewTestXMRDaemon())

	t.Run("Should Return Right Amount (Valid Tx)", func(t *testing.T) {
		test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
			// Given
			q := db.New(dbConn).WithTx(tx)
			userId, _, _ := createUserWithXmrData(ctx, q)

			expectedTxId := "eae833d591cf3333c1002c10ac4e8e74e65328a93933b404d6e40437911bf1cc"
			expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
				UserID:                userId,
				Coin:                  db.CoinTypeXMR,
				CryptoAddress:         "74xhb5sXRsnDZv8RKFEv7LAMfUq5AmGEEB77SVvsUJf8bLvFMSEfc8YYyJHF6xNNnjAZQmgqZp76AjT8bD6qKkLZLeR42oi",
				RequiredAmount:        0.00101,
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
			amount, err := verifyXMRTxHandler(ctx, q, &verifyTxHandlerData[listener.XMRTx]{invoice: expectedInvoice, tx: txs[0]})

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedInvoice.RequiredAmount, amount)
		})
	})

	t.Run("Should Return 0 Amount (Invalid Tx)", func(t *testing.T) {
		test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
			// Given
			q := db.New(dbConn).WithTx(tx)
			userId, _, _ := createUserWithXmrData(ctx, q)

			expectedTxId := "7c9b8bc6278b0a5b957b1cf099f92a471be55ce1e9a8b25e3b364eb4f90f9b6f"
			expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
				UserID:                userId,
				Coin:                  db.CoinTypeXMR,
				CryptoAddress:         "74xhb5sXRsnDZv8RKFEv7LAMfUq5AmGEEB77SVvsUJf8bLvFMSEfc8YYyJHF6xNNnjAZQmgqZp76AjT8bD6qKkLZLeR42oi",
				RequiredAmount:        0.00101,
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
			amount, err := verifyXMRTxHandler(ctx, q, &verifyTxHandlerData[listener.XMRTx]{invoice: expectedInvoice, tx: txs[0]})

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, float64(0), amount)
		})
	})
}
