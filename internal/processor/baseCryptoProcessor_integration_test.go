package processor

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/internal/util"
	"github.com/chekist32/goipay/test"
	db_test "github.com/chekist32/goipay/test/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func createNewTestBaseCryptoProcessor() (chan db.Invoice, *baseCryptoProcessor, testcontainers.Container, func(ctx context.Context)) {
	invoiceCn := make(chan db.Invoice)
	pendingInvoices := new(util.SyncMapTypeSafe[string, pendingInvoice])
	dbConn, postgres, close := test.SpinUpPostgresContainerAndGetPgxpool(fmt.Sprintf("%v/../../sql/migrations", os.Getenv("PWD")))

	return invoiceCn,
		&baseCryptoProcessor{
			dbConnPool:      dbConn,
			invoiceCn:       invoiceCn,
			pendingInvoices: pendingInvoices,
			log:             &zerolog.Logger{},
		},
		postgres,
		close
}

func TestBroadcastInvoice(t *testing.T) {
	t.Parallel()

	t.Run("Should Receive The Invoice", func(t *testing.T) {
		ctx := context.Background()

		invoiceCn, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)
		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{
			Address:    uuid.NewString(),
			Coin:       db.CoinTypeXMR,
			IsOccupied: true,
			UserID:     userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		var expiresAt pgtype.Timestamptz
		if err := expiresAt.Scan(time.Now().UTC()); err != nil {
			log.Fatal(err)
		}
		invoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
			CryptoAddress:         expectedAddr.Address,
			Coin:                  expectedAddr.Coin,
			RequiredAmount:        1.0,
			ConfirmationsRequired: 0,
			ExpiresAt:             expiresAt,
			UserID:                userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		p.broadcastUpdatedInvoice(ctx, &invoice)
		receivedInvoice := test.GetValueFromCnOrLogFatalWithTimeout(invoiceCn, listener.MIN_SYNC_TIMEOUT, "Timeout expired")

		assert.Equal(t, invoice, receivedInvoice)
	})
}

func TestReleaseAddressHelper(t *testing.T) {
	t.Parallel()

	t.Run("Should Properly Release The Address", func(t *testing.T) {
		ctx := context.Background()

		_, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)
		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{
			Address:    uuid.NewString(),
			Coin:       db.CoinTypeXMR,
			IsOccupied: true,
			UserID:     userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		var expiresAt pgtype.Timestamptz
		if err := expiresAt.Scan(time.Now().UTC()); err != nil {
			log.Fatal(err)
		}
		invoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
			CryptoAddress:         expectedAddr.Address,
			Coin:                  expectedAddr.Coin,
			RequiredAmount:        1.0,
			ConfirmationsRequired: 0,
			ExpiresAt:             expiresAt,
			UserID:                userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: invoice.UserID, Coin: invoice.Coin})
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		p.releaseAddressHelper(ctx, &invoice)

		addr, err := q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: invoice.UserID, Coin: invoice.Coin})
		assert.NoError(t, err)
		assert.Equal(t, expectedAddr, addr)
	})
}

func TestExpireInvoice(t *testing.T) {
	t.Parallel()

	t.Run("Should Properly Expire The Invoice And Release The Address", func(t *testing.T) {
		ctx := context.Background()

		invoiceCn, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)
		qTest := db_test.New(p.dbConnPool)
		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{
			Address:    uuid.NewString(),
			Coin:       db.CoinTypeXMR,
			IsOccupied: true,
			UserID:     userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		var expiresAt pgtype.Timestamptz
		if err := expiresAt.Scan(time.Now().UTC()); err != nil {
			log.Fatal(err)
		}
		expectedPendingInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
			CryptoAddress:         expectedAddr.Address,
			Coin:                  expectedAddr.Coin,
			RequiredAmount:        1.0,
			ConfirmationsRequired: 0,
			ExpiresAt:             expiresAt,
			UserID:                userId,
		})
		if err != nil {
			log.Fatal(err)
		}
		invoicePtr := &atomic.Pointer[db.Invoice]{}
		invoicePtr.Store(&expectedPendingInvoice)
		_, cancel := context.WithCancel(ctx)
		p.pendingInvoices.Store(expectedPendingInvoice.CryptoAddress, pendingInvoice{invoice: invoicePtr, cancelTimeoutFunc: cancel})

		_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: expectedPendingInvoice.UserID, Coin: expectedPendingInvoice.Coin})
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		go p.expireInvoice(ctx, &expectedPendingInvoice)

		expiredInvoice := test.GetValueFromCnOrLogFatalWithTimeout(invoiceCn, listener.MIN_SYNC_TIMEOUT, "Timeout expired")
		assert.Equal(t, expectedPendingInvoice.ID, expiredInvoice.ID)
		assert.Equal(t, db.InvoiceStatusTypeEXPIRED, expiredInvoice.Status)

		// wait for releaseAddressHelper
		<-time.After(1 * time.Second)

		addr, err := q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: expectedPendingInvoice.UserID, Coin: expectedPendingInvoice.Coin})
		assert.NoError(t, err)
		assert.Equal(t, expectedAddr, addr)

		expiredInvoice = getInvoiceOrFatal(ctx, qTest, &expectedPendingInvoice.ID)
		assert.Equal(t, expectedPendingInvoice.ID, expiredInvoice.ID)
		assert.Equal(t, db.InvoiceStatusTypeEXPIRED, expiredInvoice.Status)
		assert.False(t, expiredInvoice.ConfirmedAt.Valid)
	})
}

func TestConfirmInvoice(t *testing.T) {
	t.Parallel()

	t.Run("Should Properly Confirm The Invoice And Release The Address", func(t *testing.T) {
		ctx := context.Background()

		_, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)
		qTest := db_test.New(p.dbConnPool)
		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{
			Address:    uuid.NewString(),
			Coin:       db.CoinTypeXMR,
			IsOccupied: true,
			UserID:     userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		var expiresAt pgtype.Timestamptz
		if err := expiresAt.Scan(time.Now().UTC()); err != nil {
			log.Fatal(err)
		}
		expectedPendingInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
			CryptoAddress:         expectedAddr.Address,
			Coin:                  expectedAddr.Coin,
			RequiredAmount:        1.0,
			ConfirmationsRequired: 0,
			ExpiresAt:             expiresAt,
			UserID:                userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: expectedPendingInvoice.UserID, Coin: expectedPendingInvoice.Coin})
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		p.confirmInvoice(ctx, &expectedPendingInvoice)

		confirmedInvoice := getInvoiceOrFatal(ctx, qTest, &expectedPendingInvoice.ID)
		assert.Equal(t, expectedPendingInvoice.ID, confirmedInvoice.ID)
		assert.Equal(t, db.InvoiceStatusTypeCONFIRMED, confirmedInvoice.Status)
		assert.NotNil(t, confirmedInvoice.ConfirmedAt)
	})
}

func TestPersistCryptoCacheHelper(t *testing.T) {
	t.Parallel()

	t.Run("Should Properly Persist Crypto Cache", func(t *testing.T) {
		ctx := context.Background()

		_, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)

		cache, err := q.FindCryptoCacheByCoin(ctx, db.CoinTypeXMR)
		if err != nil {
			log.Fatal(err)
		}
		assert.False(t, cache.LastSyncedBlockHeight.Valid)

		lastHeight := int64(123)
		p.persistCryptoCacheHelper(ctx, db.CoinTypeXMR, lastHeight)

		cache, err = q.FindCryptoCacheByCoin(ctx, db.CoinTypeXMR)
		assert.NoError(t, err)
		assert.True(t, cache.LastSyncedBlockHeight.Valid)
		assert.Equal(t, lastHeight, cache.LastSyncedBlockHeight.Int64)
	})
}

func TestHandleInvoice(t *testing.T) {
	t.Parallel()

	t.Run("Should Delete Expired Invoice From The Pending Invoices", func(t *testing.T) {
		ctx := context.Background()

		invoiceCn, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)
		qTest := db_test.New(p.dbConnPool)
		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{
			Address:    uuid.NewString(),
			Coin:       db.CoinTypeXMR,
			IsOccupied: true,
			UserID:     userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		var expiresAt pgtype.Timestamptz
		if err := expiresAt.Scan(time.Now().UTC()); err != nil {
			log.Fatal(err)
		}
		expectedPendingInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
			CryptoAddress:         expectedAddr.Address,
			Coin:                  expectedAddr.Coin,
			RequiredAmount:        1.0,
			ConfirmationsRequired: 0,
			ExpiresAt:             expiresAt,
			UserID:                userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		p.handleInvoice(ctx, expectedPendingInvoice)

		_ = test.GetValueFromCnOrLogFatalWithTimeout(invoiceCn, listener.MIN_SYNC_TIMEOUT, "Timeout expired")
		<-time.After(1 * time.Second)

		_, ok := p.pendingInvoices.Load(expectedPendingInvoice.CryptoAddress)
		assert.False(t, ok)

		addr, err := q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: expectedPendingInvoice.UserID, Coin: expectedPendingInvoice.Coin})
		assert.NoError(t, err)
		assert.Equal(t, expectedAddr, addr)

		expiredInvoice := getInvoiceOrFatal(ctx, qTest, &expectedPendingInvoice.ID)
		assert.Equal(t, expectedPendingInvoice.ID, expiredInvoice.ID)
		assert.Equal(t, db.InvoiceStatusTypeEXPIRED, expiredInvoice.Status)
		assert.False(t, expiredInvoice.ConfirmedAt.Valid)
	})

	t.Run("Should Return From handleInvoiceHelper", func(t *testing.T) {
		ctx := context.Background()

		_, p, _, close := createNewTestBaseCryptoProcessor()
		defer close(ctx)

		q := db.New(p.dbConnPool)
		qTest := db_test.New(p.dbConnPool)
		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{
			Address:    uuid.NewString(),
			Coin:       db.CoinTypeXMR,
			IsOccupied: true,
			UserID:     userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		var expiresAt pgtype.Timestamptz
		if err := expiresAt.Scan(time.Now().UTC().Add(1 * time.Hour)); err != nil {
			log.Fatal(err)
		}
		expectedPendingInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
			CryptoAddress:         expectedAddr.Address,
			Coin:                  expectedAddr.Coin,
			RequiredAmount:        1.0,
			ConfirmationsRequired: 0,
			ExpiresAt:             expiresAt,
			UserID:                userId,
		})
		if err != nil {
			log.Fatal(err)
		}

		cancelCtx, cancel := context.WithCancel(ctx)

		p.handleInvoice(cancelCtx, expectedPendingInvoice)
		cancel()

		_, ok := p.pendingInvoices.Load(expectedPendingInvoice.CryptoAddress)
		assert.True(t, ok)

		_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{UserID: expectedPendingInvoice.UserID, Coin: expectedPendingInvoice.Coin})
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		pendingInvoice := getInvoiceOrFatal(ctx, qTest, &expectedPendingInvoice.ID)
		assert.Equal(t, expectedPendingInvoice, pendingInvoice)
	})
}
