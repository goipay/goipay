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
	"github.com/stretchr/testify/assert"
)

func createUserWithBtcData(ctx context.Context, q *db.Queries) (pgtype.UUID, db.CryptoDatum, db.BtcCryptoDatum) {
	userId, err := q.CreateUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	btcData, err := q.CreateBTCCryptoData(ctx, "tpubDCUURn3yPT4P3SkrUq9rG1RyJK6BGhmrovvSAF61LHLCZhNUMRw7FANPmhGuDWXo3GMkc6C4ZFGBuPMrovjdnXhtJfQE3uK3s6QzFuiQaz9")
	if err != nil {
		log.Fatal(err)
	}
	cd, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{BtcID: btcData.ID, UserID: userId})
	if err != nil {
		log.Fatal(err)
	}

	return userId, cd, btcData
}

func createNewTestBtcDaemon() *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         "nd-457-466-409.p2pify.com/a11fc6c0ed68edbea7096b4bd950db15",
		User:         "angry-kowalevski",
		Pass:         "galley-zone-nectar-swerve-unread-sprang",
		HTTPPostMode: true,
		DisableTLS:   false,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func TestGenerateNextBtcAddressHandler(t *testing.T) {
	t.Parallel()

	data := []struct {
		prevMajorIndex int32
		prevMinorIndex int32
		expectedAddr   string
	}{
		{
			prevMajorIndex: 0,
			prevMinorIndex: 0,
			expectedAddr:   "tb1qqdcfs9s5gjsnmazcsqfe2h6gwzwdu2eufesk8h",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: 124,
			expectedAddr:   "tb1qg3hmjm7waaj3ejem5qt9kzcm40qppmq49lgpnk",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: math.MaxInt32,
			expectedAddr:   "tb1q4lnztm5gjh3jqeahl00gk85aprqcm9vdl3gzr8",
		},
		{
			prevMajorIndex: 1,
			prevMinorIndex: 2,
			expectedAddr:   "tb1qwfu6lpvaxekpz7qjdlfvwrs44yaq02gj509pxn",
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
				userId, cd, _ := createUserWithBtcData(ctx, q)

				_, err := qT.UpdateIndicesBTCCryptoDataById(ctx, test_db.UpdateIndicesBTCCryptoDataByIdParams{ID: cd.BtcID, LastMajorIndex: d.prevMajorIndex, LastMinorIndex: d.prevMinorIndex})
				if err != nil {
					log.Fatal(err)
				}

				// When
				addr, err := generateNextBTCAddressHandler(ctx, q, &generateNextAddressHandlerData{userId: userId, network: listener.SignetBTC})

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, d.expectedAddr, addr.Address)
			})
		})
	}

}

func TestVerifyBTCTxHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConn, _, close := getPostgresWithDbConn()
	defer close(ctx)

	daemon := listener.NewSharedBTCDaemonRpcClient(createNewTestBtcDaemon())

	t.Run("Should Return Right Amount (Valid Tx)", func(t *testing.T) {
		test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
			// Given
			q := db.New(dbConn).WithTx(tx)
			userId, _, _ := createUserWithXmrData(ctx, q)

			expectedTxId := "ca2329777b4c886750347b5bf6e53d5dabc0b4f0cfe5fb3694c569d845eb1abd"
			expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
				UserID:                userId,
				Coin:                  db.CoinTypeBTC,
				CryptoAddress:         "bc1q8e8qkxqtgfypwwnh6zf5msx82yw2p4l9sy26ey",
				RequiredAmount:        0.00480740,
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
			amount, err := verifyBTCTxHandler(ctx, q, &verifyTxHandlerData[listener.BTCTx]{invoice: expectedInvoice, tx: txs[0]})

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedInvoice.RequiredAmount, amount)
		})
	})

	t.Run("Should Return 0 Amount (Invalid Tx)", func(t *testing.T) {
		test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
			// Given
			q := db.New(dbConn).WithTx(tx)
			userId, _, _ := createUserWithBtcData(ctx, q)

			expectedTxId := "b1c496d9e3bd4eeff0b33d0ce6b5c2541244cc71f1fc4051aff83b77c8dabf80"
			expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
				UserID:                userId,
				Coin:                  db.CoinTypeBTC,
				CryptoAddress:         "bc1q8e8qkxqtgfypwwnh6zf5msx82yw2p4l9sy26ey",
				RequiredAmount:        0.00480740,
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
			amount, err := verifyBTCTxHandler(ctx, q, &verifyTxHandlerData[listener.BTCTx]{invoice: expectedInvoice, tx: txs[0]})

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, float64(0), amount)
		})
	})
}
