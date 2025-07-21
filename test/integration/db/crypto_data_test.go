package db_test

import (
	"context"
	"log"
	"math"
	"testing"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/test"
	test_db "github.com/chekist32/goipay/test/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func createRandomXMRCryptoData(ctx context.Context, q *db.Queries) (db.XmrCryptoDatum, error) {
	return q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: uuid.NewString(), PubSpendKey: uuid.NewString()})
}

func TestCreateXMRCryptoData(t *testing.T) {
	t.Run("Should Return Valid XMR Crypto Data", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			_, err := createRandomXMRCryptoData(ctx, q)
			assert.NoError(t, err)
		})
	})

	t.Run("Should Return SQL Error (non unique public spend key)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr1, err := createRandomXMRCryptoData(ctx, q)
			assert.NoError(t, err)

			_, err = q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: uuid.NewString(), PubSpendKey: xmr1.PubSpendKey})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23505", pgErr.Code)
		})
	})

	t.Run("Should Return SQL Error (non unique private view key)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr1, err := createRandomXMRCryptoData(ctx, q)
			assert.NoError(t, err)

			_, err = q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: xmr1.PrivViewKey, PubSpendKey: uuid.NewString()})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23505", pgErr.Code)
		})
	})

}

func TestCreateCryptoData(t *testing.T) {
	t.Run("Should Return SQL Error (invalid userId)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			var userId pgtype.UUID
			if err := userId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

	t.Run("Should Return SQL Error (invalid xmrId)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmrId, UserID: userId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

	t.Run("Should Return Valid Crypto Data", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			assert.NoError(t, err)
		})
	})
}

func TestFindCryptoDataByUserId(t *testing.T) {
	t.Run("Should Return Valid Crypto Data", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			expectedCryptoData, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			cryptoData, err := q.FindCryptoDataByUserId(ctx, userId)
			assert.NoError(t, err)
			assert.Equal(t, expectedCryptoData, cryptoData)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			var userId1 pgtype.UUID
			if err := userId1.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindCryptoDataByUserId(ctx, userId1)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})
}

func TestFindKeysXMRCryptoDataById(t *testing.T) {
	t.Run("Should Return Proper XMR Crypto Data", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			expectedCryptoData, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			keys, err := q.FindKeysXMRCryptoDataById(ctx, expectedCryptoData.XmrID)
			assert.NoError(t, err)
			assert.Equal(t, xmr.PrivViewKey, keys.PrivViewKey)
			assert.Equal(t, xmr.PubSpendKey, keys.PubSpendKey)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindKeysXMRCryptoDataById(ctx, xmrId)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindKeysXMRCryptoDataById(ctx, xmrId)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

}

func TestFindKeysAndIncrementedIndicesXMRCryptoDataById(t *testing.T) {
	data := []struct {
		prevMajorIndex     int32
		prevMinorIndex     int32
		expectedMajorIndex int32
		expectedMinorIndex int32
	}{
		{
			prevMajorIndex:     0,
			prevMinorIndex:     0,
			expectedMajorIndex: 0,
			expectedMinorIndex: 1,
		},
		{
			prevMajorIndex:     0,
			prevMinorIndex:     124,
			expectedMajorIndex: 0,
			expectedMinorIndex: 125,
		},
		{
			prevMajorIndex:     0,
			prevMinorIndex:     math.MaxInt32,
			expectedMajorIndex: 1,
			expectedMinorIndex: 0,
		},
		{
			prevMajorIndex:     1,
			prevMinorIndex:     2,
			expectedMajorIndex: 1,
			expectedMinorIndex: 3,
		},
	}

	for _, v := range data {
		t.Run("Should Return Properly Incremented Indices", func(t *testing.T) {
			test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
				ctx := context.Background()
				q := db.New(tx)
				qT := test_db.New(tx)

				xmr, err := createRandomXMRCryptoData(ctx, q)
				if err != nil {
					log.Fatal(err)
				}

				_, err = qT.UpdateIndicesXMRCryptoDataById(ctx, test_db.UpdateIndicesXMRCryptoDataByIdParams{ID: xmr.ID, LastMajorIndex: v.prevMajorIndex, LastMinorIndex: v.prevMinorIndex})
				if err != nil {
					log.Fatal(err)
				}

				keysAndIndices, err := q.FindKeysAndIncrementedIndicesXMRCryptoDataById(ctx, xmr.ID)
				assert.NoError(t, err)
				assert.Equal(t, v.expectedMajorIndex, keysAndIndices.LastMajorIndex)
				assert.Equal(t, v.expectedMinorIndex, keysAndIndices.LastMinorIndex)
			})
		})
	}

}

func TestUpdateKeysXMRCryptoDataById(t *testing.T) {
	test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		xmr, err := createRandomXMRCryptoData(ctx, q)
		if err != nil {
			log.Fatal(err)
		}

		_, err = q.UpdateKeysXMRCryptoDataById(ctx, db.UpdateKeysXMRCryptoDataByIdParams{ID: xmr.ID, PrivViewKey: uuid.NewString(), PubSpendKey: uuid.NewString()})
		assert.NoError(t, err)
	})
}

func TestSetXMRCryptoDataByUserId(t *testing.T) {
	t.Run("Should Return Valid Crypto Data", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			xmr1, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.SetXMRCryptoDataByUserId(ctx, db.SetXMRCryptoDataByUserIdParams{UserID: userId, XmrID: xmr1.ID})
			assert.NoError(t, err)
		})
	})

	t.Run("Should Return SQL Error (invalid xmr_id)", func(t *testing.T) {
		test.RunInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.SetXMRCryptoDataByUserId(ctx, db.SetXMRCryptoDataByUserIdParams{UserID: userId, XmrID: xmrId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

}
