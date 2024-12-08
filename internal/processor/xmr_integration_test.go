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
	"github.com/chekist32/go-monero/utils"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/internal/util"
	"github.com/chekist32/goipay/test"
	db_test "github.com/chekist32/goipay/test/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	priv_view_key   = "8aa763d1c8d9da4ca75cb6ca22a021b5cca376c1367be8d62bcc9cdf4b926009"
	pub_spend_key   = "38e9908d33d034de0ba1281aa7afe3907b795cea14852b3d8fe276e8931cb130"
	xmr_daemon_addr = "http://node.monerodevs.org:38089"
)

func getInvoiceOrFatal(ctx context.Context, q *db_test.Queries, id *pgtype.UUID) db.Invoice {
	invoice, err := q.FindInvoiceById(ctx, *id)
	if err != nil {
		log.Fatal(err)
	}

	return db.Invoice{
		ID:                    invoice.ID,
		CryptoAddress:         invoice.CryptoAddress,
		Coin:                  db.CoinType(invoice.Coin),
		RequiredAmount:        invoice.RequiredAmount,
		ActualAmount:          invoice.ActualAmount,
		ConfirmationsRequired: invoice.ConfirmationsRequired,
		CreatedAt:             invoice.CreatedAt,
		ConfirmedAt:           invoice.ConfirmedAt,
		Status:                db.InvoiceStatusType(invoice.Status),
		ExpiresAt:             invoice.ExpiresAt,
		TxID:                  invoice.TxID,
		UserID:                invoice.UserID,
	}
}

func createUserWithXmrData(ctx context.Context, q *db.Queries) (pgtype.UUID, db.CryptoDatum, db.XmrCryptoDatum) {
	userId, err := q.CreateUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	xmrData, err := q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: priv_view_key, PubSpendKey: pub_spend_key})
	if err != nil {
		log.Fatal(err)
	}
	cd, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmrData.ID, UserID: userId})
	if err != nil {
		log.Fatal(err)
	}

	return userId, cd, xmrData
}

func createNewTestDaemon(urlStr string, user string, pass string) (daemon.IDaemonRpcClient, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return daemon.NewDaemonRpcClient(daemon.NewRpcConnection(u, user, pass)), nil
}

func createNewTestXmrProcessor() (chan db.Invoice, *xmrProcessor, testcontainers.Container, func(ctx context.Context)) {
	invoiceCn := make(chan db.Invoice)
	pendingInvoices := new(util.SyncMapTypeSafe[string, pendingInvoice])
	dbConn, postgres, close := test.SpinUpPostgresContainerAndGetPgxpool(fmt.Sprintf("%v/../../sql/migrations", os.Getenv("PWD")))

	d, err := createNewTestDaemon(xmr_daemon_addr, "", "")
	if err != nil {
		log.Fatal(err)
	}

	xmr := &xmrProcessor{
		daemon:   d,
		daemonEx: listener.NewDaemonRpcClientExecutor(d, &zerolog.Logger{}),
		network:  utils.Stagenet,
		baseCryptoProcessor: baseCryptoProcessor{
			log:             &zerolog.Logger{},
			dbConnPool:      dbConn,
			invoiceCn:       invoiceCn,
			pendingInvoices: pendingInvoices,
		},
	}

	return invoiceCn,
		xmr,
		postgres,
		close
}

func TestGenerateNextXmrAddressHelper(t *testing.T) {
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

	for _, d := range data {
		t.Run(fmt.Sprintf("Should Return Valid Address Ma %v Mi %v", d.prevMajorIndex, d.prevMinorIndex), func(t *testing.T) {
			// Given
			ctx := context.Background()

			_, p, _, close := createNewTestXmrProcessor()
			defer close(ctx)

			q := db.New(p.dbConnPool)
			userId, cd, _ := createUserWithXmrData(ctx, q)

			_, err := q.UpdateIndicesXMRCryptoDataById(ctx, db.UpdateIndicesXMRCryptoDataByIdParams{ID: cd.XmrID, LastMajorIndex: d.prevMajorIndex, LastMinorIndex: d.prevMinorIndex})
			if err != nil {
				log.Fatal(err)
			}

			// When
			addr, err := p.generateNextXmrAddressHelper(ctx, q, userId)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, d.expectedAddr, addr.Address)
		})
	}

}

func TestCreateInvoice(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	_, p, _, close := createNewTestXmrProcessor()
	defer close(ctx)

	q := db.New(p.dbConnPool)
	userId, _, _ := createUserWithXmrData(ctx, q)

	// When
	createdInvoice, err := p.createInvoice(ctx, &dto.NewInvoiceRequest{
		UserId:        util.PgUUIDToString(userId),
		Coin:          db.CoinTypeXMR,
		Amount:        123,
		Timeout:       600,
		Confirmations: 0,
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, db.CoinTypeXMR, createdInvoice.Coin)
	assert.EqualValues(t, 123, createdInvoice.RequiredAmount)
	assert.EqualValues(t, 0, createdInvoice.ConfirmationsRequired)
	assert.Equal(t, "74xhb5sXRsnDZv8RKFEv7LAMfUq5AmGEEB77SVvsUJf8bLvFMSEfc8YYyJHF6xNNnjAZQmgqZp76AjT8bD6qKkLZLeR42oi", createdInvoice.CryptoAddress)

}

func TestPersistCryptoCache(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	_, p, _, close := createNewTestXmrProcessor()
	defer close(ctx)

	q := db.New(p.dbConnPool)

	// When
	p.persistCryptoCache(ctx)

	// Assert
	cache, err := q.FindCryptoCacheByCoin(ctx, db.CoinTypeXMR)

	assert.NoError(t, err)
	assert.True(t, cache.LastSyncedBlockHeight.Valid)
}

func TestNewXmrProcessor(t *testing.T) {
	t.Parallel()

	// Given
	config := &dto.DaemonsConfig{Xmr: dto.DaemonConfig{Url: xmr_daemon_addr, User: "", Pass: ""}}

	// When
	xmr, err := newXmrProcessor(nil, make(chan db.Invoice), config, &zerolog.Logger{})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, utils.Stagenet, xmr.network)
	assert.NotNil(t, xmr.pendingInvoices)
	assert.NotNil(t, xmr.daemon)
	assert.NotNil(t, xmr.daemonEx)
}

func TestHandleInvoicePbReq(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	invCn, p, _, close := createNewTestXmrProcessor()
	defer close(ctx)

	q := db.New(p.dbConnPool)
	qTest := db_test.New(p.dbConnPool)
	userId, _, _ := createUserWithXmrData(ctx, q)

	req := &dto.NewInvoiceRequest{
		UserId:        util.PgUUIDToString(userId),
		Coin:          db.CoinTypeXMR,
		Amount:        123,
		Timeout:       600,
		Confirmations: 0,
	}

	// When
	invoice, err := p.handleInvoicePbReq(ctx, req)

	// Assert
	assert.NoError(t, err)

	expectedInvoice := getInvoiceOrFatal(ctx, qTest, &invoice.ID)

	assert.Equal(t, expectedInvoice, *invoice)

	pendingInvoice, ok := p.pendingInvoices.Load(expectedInvoice.CryptoAddress)
	assert.True(t, ok)
	assert.Equal(t, expectedInvoice, *pendingInvoice.invoice.Load())

	invoiceFromCn := test.GetValueFromCnOrLogFatalWithTimeout(invCn, util.SEND_TIMEOUT, "Timeout expired")
	assert.Equal(t, expectedInvoice, invoiceFromCn)
}

func TestLoad(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	_, p, _, close := createNewTestXmrProcessor()
	defer close(ctx)

	q := db.New(p.dbConnPool)

	// When
	err := p.load(ctx)

	// Assert
	assert.NoError(t, err)

	<-time.After(500 * time.Millisecond)
	cache, err := q.FindCryptoCacheByCoin(ctx, db.CoinTypeXMR)
	if err != nil {
		log.Fatal(err)
	}

	assert.True(t, cache.LastSyncedBlockHeight.Valid)
}

func TestVerifyMoneroTxOnTxMempool(t *testing.T) {
	t.Parallel()

	// Given
	ctx := context.Background()

	invCn, p, _, close := createNewTestXmrProcessor()
	defer close(ctx)

	q := db.New(p.dbConnPool)
	qTest := db_test.New(p.dbConnPool)
	userId, _, _ := createUserWithXmrData(ctx, q)

	expectedTxId := "eae833d591cf3333c1002c10ac4e8e74e65328a93933b404d6e40437911bf1cc"
	expectedReq := &dto.NewInvoiceRequest{
		UserId:        util.PgUUIDToString(userId),
		Coin:          db.CoinTypeXMR,
		Amount:        0.00101,
		Timeout:       600,
		Confirmations: 0,
	}
	if _, err := p.handleInvoicePbReq(ctx, expectedReq); err != nil {
		log.Fatal(err)
	}

	expectedPendingInvoice := test.GetValueFromCnOrLogFatalWithTimeout(invCn, util.SEND_TIMEOUT, "Timeout expired")

	txs, err := p.daemon.GetTransactions([]string{expectedTxId}, true, false, false)
	if err != nil {
		log.Fatal(err)
	} else if len(txs.Txs) < 1 {
		log.Fatalf("Invalid Tx Id: %v", expectedTxId)
	}

	// When
	p.verifyMoneroTxOnTxMempool(ctx, incomingMoneroTxGetTx(txs.Txs[0]))

	// Assert
	pendingMempoolInvoice := test.GetValueFromCnOrLogFatalWithTimeout(invCn, util.SEND_TIMEOUT, "Timeout expired")
	assert.Equal(t, util.PgUUIDToString(expectedPendingInvoice.ID), util.PgUUIDToString(pendingMempoolInvoice.ID))
	assert.Equal(t, expectedReq.Amount, pendingMempoolInvoice.ActualAmount.Float64)
	assert.Equal(t, db.InvoiceStatusTypePENDINGMEMPOOL, pendingMempoolInvoice.Status)
	assert.Equal(t, expectedTxId, pendingMempoolInvoice.TxID.String)

	confirmedInvoice := test.GetValueFromCnOrLogFatalWithTimeout(invCn, util.SEND_TIMEOUT, "Timeout expired")
	assert.Equal(t, util.PgUUIDToString(expectedPendingInvoice.ID), util.PgUUIDToString(confirmedInvoice.ID))
	assert.Equal(t, db.InvoiceStatusTypeCONFIRMED, confirmedInvoice.Status)
	assert.True(t, confirmedInvoice.ConfirmedAt.Valid)

	confirmedInvoiceDb := getInvoiceOrFatal(ctx, qTest, &confirmedInvoice.ID)
	assert.Equal(t, confirmedInvoice, confirmedInvoiceDb)

	_, ok := p.pendingInvoices.Load(confirmedInvoice.CryptoAddress)
	assert.False(t, ok)

	<-time.After(100 * time.Millisecond)

	addr, err := qTest.FindCryptoAddressByAddress(ctx, pendingMempoolInvoice.CryptoAddress)
	assert.NoError(t, err)
	assert.False(t, addr.IsOccupied)
}
