package processor

import (
	"context"
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/chekist32/goipay/test"
	test_db "github.com/chekist32/goipay/test/db"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func createUserWithBnbData(ctx context.Context, q *db.Queries) (pgtype.UUID, db.CryptoDatum, db.BnbCryptoDatum) {
	userId, err := q.CreateUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	bnbData, err := q.CreateBNBCryptoData(ctx, "xpub6CUf84eg4Ba1jJ3ePzLSSoeQ1ENzP33zCN4982Xoi1TZ1kfYreZe5ECqLm4RVWQHpuB5gixi3gK1PykXzcwWxW7w6d7GWxpsNY7wxNVBHip")
	if err != nil {
		log.Fatal(err)
	}
	cd, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{BnbID: bnbData.ID, UserID: userId})
	if err != nil {
		log.Fatal(err)
	}

	return userId, cd, bnbData
}

func createNewTestBnbDaemon() *ethclient.Client {
	client, err := ethclient.Dial("https://bsc-dataseed.binance.org")
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func TestGenerateNextBnbAddressHandler(t *testing.T) {
	t.Parallel()

	data := []struct {
		prevMajorIndex int32
		prevMinorIndex int32
		expectedAddr   string
	}{
		{prevMajorIndex: 0, prevMinorIndex: 0, expectedAddr: "0x52bDE05866773a211aB01BbaEa9C474B9f24754D"},
		{prevMajorIndex: 0, prevMinorIndex: 124, expectedAddr: "0x947d04a4e66Ac7e26B4207D212b4F0903B09F89F"},
		{prevMajorIndex: 0, prevMinorIndex: math.MaxInt32, expectedAddr: "0x2457A40c2C0D095Dcce88F5919F3afDc60e15CF9"},
		{prevMajorIndex: 1, prevMinorIndex: 2, expectedAddr: "0x4d31b4584A64aD34C6AadB3388620682D2fd9D01"},
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
				userId, cd, _ := createUserWithBnbData(ctx, q)

				_, err := qT.UpdateIndicesBNBCryptoDataById(ctx, test_db.UpdateIndicesBNBCryptoDataByIdParams{ID: cd.BnbID, LastMajorIndex: d.prevMajorIndex, LastMinorIndex: d.prevMinorIndex})
				if err != nil {
					log.Fatal(err)
				}

				// When
				addr, err := generateNextBNBAddressHandler(ctx, q, &generateNextAddressHandlerData{userId: userId, network: listener.MainnetBNB})

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, d.expectedAddr, addr.Address)
			})
		})
	}

}

func TestVerifyBNBTxHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConn, _, close := getPostgresWithDbConn()
	defer close(ctx)

	daemon := listener.NewSharedBNBDaemonRpcClient(createNewTestBnbDaemon())

	t.Run("Should Return Right Amount (Valid Tx)", func(t *testing.T) {
		data := []struct {
			txId    string
			coin    db.CoinType
			amount  float64
			address string
		}{
			{txId: "0xcf25f3e87a652dd45b04414149de2671436f6bcafbe147385b549694461f446d", coin: db.CoinTypeBNB, amount: 0.006, address: "0x36A4d5A2CCB2C73A98C996003bb18A604387e9A1"},
			{txId: "0xb17e7091d7cff4a84e285b741d0cb178a2578ac89172058faaa05634d8eda9a0", coin: db.CoinTypeBSCUSDBEP20, amount: 79.04, address: "0x75c66FF6d9beA32C03740Fe6Fed1E8857d57Fc3d"},
			{txId: "0x82c201066e8293e20b229b929f825df21a6aed3d56a7dd8d9311641c317c11d1", coin: db.CoinTypeUSDCBEP20, amount: 535.72436856, address: "0xc224c5398e4131fb30bb396a9C2377aAF3585B8a"},
			{txId: "0x9e16caec0e6e644e464090590b2f95d0887aa80668d5c8ce89b8521ac51ed847", coin: db.CoinTypeDAIBEP20, amount: 396.704, address: "0x79cD50f440271e48F36994c2c6567B5294e981C8"},
			{txId: "0xa10bc157dfe3c71c3c594687c06bd596c2b6a7143615070d45257d9aae05f933", coin: db.CoinTypeWBTCBEP20, amount: 0.0001167, address: "0x8880aF1800D817499FB2e2D5A4d05De025f0Bdb2"},
			{txId: "0x31ecbccead967fc531b99ecbe8a760ab5b63aa1d28fb0fa683a311ad771987ca", coin: db.CoinTypeUNIBEP20, amount: 25.26665, address: "0x8D802a6212E2F2A59B44a5cFCBdFc40368E2699f"},
			{txId: "0x01099d130028abcaf9e2f4fbd332d42326582824444434db4f03a334d57f4351", coin: db.CoinTypeLINKBEP20, amount: 30.67827324, address: "0x8894E0a0c962CB723c1976a4421c95949bE2D4E3"},
			{txId: "0xa21038f3c734caa3b3caf63ce805cc5acdce432fea42eb594425eab8780b7a69", coin: db.CoinTypeAAVEBEP20, amount: 0.10057147, address: "0xA4649A1942dAB1022e0D301BC61EA004d7D0C1C7"},
			{txId: "0x4996d2707167eec7b55c5f11cb6a39447580494be1035ae5607e8826b1efbac9", coin: db.CoinTypeMATICBEP20, amount: 53.135446188281994656, address: "0x8894E0a0c962CB723c1976a4421c95949bE2D4E3"},
			{txId: "0xecc0c5fee7f0f1258927f6d5bbd2438c4f3ef2dc916cf8270acf9a266e0305bf", coin: db.CoinTypeSHIBBEP20, amount: 445540.848667137766512794, address: "0xA6759f23Fe155a1AF3206b5B8C81738413E86E61"},
			{txId: "0x3cb3d5fa5c4fe034a14090087bd48e6a02e50efdc8c293a325281dbee455e3b3", coin: db.CoinTypeBUSDBEP20, amount: 3.56412, address: "0x808bA92DB0d3D1eeEf92edb076BB3F3379d0ddED"},
			{txId: "0xa358385f9944af6f0758d9f9565cf9fdb2922059b1fc496cd6808d28896e4a75", coin: db.CoinTypeATOMBEP20, amount: 3.53065681, address: "0xb7Bf10D3b0e6D1269C32360dB6bD8E13da74A375"},
			{txId: "0xf8cef6c0b50e67f97b215511ca4b42354f1034c7b64c165a973a168cbe96a2ee", coin: db.CoinTypeARBBEP20, amount: 10, address: "0x39Ba9e663e72d0d5C4153152E8CAFd40BA62F3AB"},
			{txId: "0x5320f78ff26329edcfe9cc57ecb8bb8746868282e6b1891718afc7a246d7d6af", coin: db.CoinTypeETHBEP20, amount: 0.00205115, address: "0x3a129A9Db9970f0Bfa20d5cD753Abf972672E106"},
			{txId: "0x9c893d3c5de4457a44be6c87a8d8881e9a9f8f462412f52853fa94a4e0b22a17", coin: db.CoinTypeXRPBEP20, amount: 1756.781967, address: "0xbeC9c6ec58A532Cd8ACa0Af9cE28BF814651b917"},
			{txId: "0x724a3f942f02f639612d4377e483f42d7f539e20d91e38019b69f021192c4272", coin: db.CoinTypeADABEP20, amount: 118.760980641384787593, address: "0x265EA336b5F722B1400422b73b829Ae9b116cCc4"},
			{txId: "0x19dca4e9ebd5169fbf0ad7eafafc516a69725fc31154c4d166ab5013231e3802", coin: db.CoinTypeTRXBEP20, amount: 24.466208, address: "0x28fD4BA3a1D37C88D4d49dcd988225c8B15c7792"},
			{txId: "0xbc7e311108a6f8cae53db085112def45de30532946293d54f4bc6727bba7b744", coin: db.CoinTypeDOGEBEP20, amount: 127.01012415, address: "0xbc6E76C7349aCd0CD1f9E358DA6B29A7324E309E"},
			{txId: "0x2028b340333093f99ba9f4ca093dbcfd89727727b6266d767840edd0fba7b5d2", coin: db.CoinTypeLTCBEP20, amount: 0.106423242, address: "0xB8b7c7940422C6aefB25eB0e73B7409e78986F2a"},
			{txId: "0xd103564b733e6f10a66ea7f867da3025611285af2d8c7b7a1e26348ad33f6ca3", coin: db.CoinTypeBCHBEP20, amount: 0.68702755, address: "0xf55e06Becc605A68c69075f61ED49DBEE25889B8"},
			{txId: "0xd8e7de4d0939d13cf824a456b37c04d1b512b5cce5a693453ab70b9260a3357d", coin: db.CoinTypeTWTBEP20, amount: 5704.79, address: "0xB26c83CA2d596671589992F08155C2BA3CBF89c1"},
			{txId: "0x3a89f475f3f54bfe2f8e8492591a93a2bd458fa4d81dd8ac03961f5aba1347af", coin: db.CoinTypeAVAXBEP20, amount: 49.999149, address: "0x3457E41A9D5B3B0C92e8647dA56AE189DDf0f409"},
			{txId: "0x0ef66362eb18ef3f8b9a7bfb25235d4de10e79a522458b28e4a1f46b0f52d269", coin: db.CoinTypeCAKEBEP20, amount: 11.62, address: "0xebBB2558dEB063a514BEf5878F87B09C119bFA74"},
		}

		for _, v := range data {
			t.Run(string(v.coin), func(t *testing.T) {
				test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
					// Given
					q := db.New(dbConn).WithTx(tx)
					userId, _, _ := createUserWithXmrData(ctx, q)

					expectedTxId := v.txId
					expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
						UserID:                userId,
						Coin:                  v.coin,
						CryptoAddress:         v.address,
						RequiredAmount:        v.amount,
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
					amount, err := verifyBNBTxHandler(ctx, q, &verifyTxHandlerData[listener.BNBTx]{invoice: expectedInvoice, tx: txs[0]})

					// Assert
					assert.NoError(t, err)
					assert.Equal(t, expectedInvoice.RequiredAmount, amount)
				})
			})
		}
	})

	t.Run("Should Return 0 Amount (Invalid Tx)", func(t *testing.T) {
		data := []struct {
			txId    string
			coin    db.CoinType
			amount  float64
			address string
		}{
			{txId: "0x783457c3cb776fd957ca996259c8339e47e436a93d5e3325466a7bf5c7f7d073", coin: db.CoinTypeBNB, amount: 0.006, address: "0x36A4d5A2CCB2C73A98C996003bb18A604387e9A1"},
			{txId: "0x8cbc4aaed8ea7e913ec1121ab61c60448804da08cb62ba72355758e266feff28", coin: db.CoinTypeBSCUSDBEP20, amount: 79.04, address: "0x75c66FF6d9beA32C03740Fe6Fed1E8857d57Fc3d"},
			{txId: "0xa196ad78728e367c489a3dae7031053bc5c22fab9efefd7bd602a93522c6a913", coin: db.CoinTypeUSDCBEP20, amount: 535.72436856, address: "0xc224c5398e4131fb30bb396a9C2377aAF3585B8a"},
			{txId: "0xe7d62f057fd0d57a047c3c6a866a6f5af3cac41f4ff231c692fc2a27ad0dca17", coin: db.CoinTypeDAIBEP20, amount: 396.704, address: "0x79cD50f440271e48F36994c2c6567B5294e981C8"},
			{txId: "0x3f37ad8c51bf55e2a12d6a1e4f8e0f9619495b9bb850c76c16f4be08db21ebad", coin: db.CoinTypeWBTCBEP20, amount: 0.0001167, address: "0x8880aF1800D817499FB2e2D5A4d05De025f0Bdb2"},
			{txId: "0x7385165a630c55d04797245bdfcc3d431430bf87b696f9f9b9f72ec2fcd1d509", coin: db.CoinTypeUNIBEP20, amount: 25.26665, address: "0x8D802a6212E2F2A59B44a5cFCBdFc40368E2699f"},
			{txId: "0x72b69d8c2df3cde7a70b31ac77370dd686b25ac669e3149679097472bd7e38d2", coin: db.CoinTypeLINKBEP20, amount: 30.67827324, address: "0x8894E0a0c962CB723c1976a4421c95949bE2D4E3"},
			{txId: "0xaf800c118672c393b946b5cf2e777c0b328489ae55a1cb5f2bc44d8d3ccdc5bf", coin: db.CoinTypeAAVEBEP20, amount: 0.10057147, address: "0xA4649A1942dAB1022e0D301BC61EA004d7D0C1C7"},
			{txId: "0x83622a5386e0ca2e99a6511e3d32bfedc12bc79a504086766cab5f865bf05b4d", coin: db.CoinTypeMATICBEP20, amount: 53.135446188281994656, address: "0x8894E0a0c962CB723c1976a4421c95949bE2D4E3"},
			{txId: "0x13ab05511e42516af8f93e50dddf7b1712e24846b312af403916a477270f672f", coin: db.CoinTypeSHIBBEP20, amount: 445540.848667137766512794, address: "0xA6759f23Fe155a1AF3206b5B8C81738413E86E61"},
			{txId: "0x61cb3b44211b9517f5a26c0e8a17f2971568f3d249d9c447da5b1bb13993158e", coin: db.CoinTypeBUSDBEP20, amount: 3.56412, address: "0x808bA92DB0d3D1eeEf92edb076BB3F3379d0ddED"},
			{txId: "0x91ab72575c62dba4376d1b51f1eb4303c44d68ecc9d5cf17eae09a29e568fd0d", coin: db.CoinTypeATOMBEP20, amount: 3.53065681, address: "0xb7Bf10D3b0e6D1269C32360dB6bD8E13da74A375"},
			{txId: "0x644b321ea6b31b227a1523aacab4fbb8b5ac8944ea53ccb5da1153e3402907c2", coin: db.CoinTypeARBBEP20, amount: 10, address: "0x39Ba9e663e72d0d5C4153152E8CAFd40BA62F3AB"},
			{txId: "0x22d370ce715ea342ff37dabb1aa7fcfb9bf94bf2c654a91dcea3773709b750ac", coin: db.CoinTypeETHBEP20, amount: 0.00205115, address: "0x3a129A9Db9970f0Bfa20d5cD753Abf972672E106"},
			{txId: "0xd2b26352fcd5e14a6bcbe6f3e68038df36038fef4eaf43636e0e7463369dd918", coin: db.CoinTypeXRPBEP20, amount: 1756.781967, address: "0xbeC9c6ec58A532Cd8ACa0Af9cE28BF814651b917"},
			{txId: "0xe9adea900f3916ad600389a81f6bec8bebc69f50e4a2fe290df66cf0077d5a17", coin: db.CoinTypeADABEP20, amount: 118.760980641384787593, address: "0x265EA336b5F722B1400422b73b829Ae9b116cCc4"},
			{txId: "0x4f9a5902c6fc64f703d39b3da87636f6e1b1c4f0f3f0357393fdd7d8fbc64dde", coin: db.CoinTypeTRXBEP20, amount: 24.466208, address: "0x28fD4BA3a1D37C88D4d49dcd988225c8B15c7792"},
			{txId: "0x947e44b3ae12acf2de37c9af916b5ae4f1a7157839f35f7370318e145bf3fbdb", coin: db.CoinTypeDOGEBEP20, amount: 127.01012415, address: "0xbc6E76C7349aCd0CD1f9E358DA6B29A7324E309E"},
			{txId: "0x766c0e0fc164e0034e3ae2051b09a526de8225955892e3fc0d0c394f270dc5ed", coin: db.CoinTypeLTCBEP20, amount: 0.106423242, address: "0xB8b7c7940422C6aefB25eB0e73B7409e78986F2a"},
			{txId: "0x1e9895667a7e6fee11ea1902ddb6321d26c4ef4e4b942e43e01e47fc0b64afe2", coin: db.CoinTypeBCHBEP20, amount: 0.68702755, address: "0xf55e06Becc605A68c69075f61ED49DBEE25889B8"},
			{txId: "0x600b12c5503bd86a24773cf9f4320b1d1a3f5d947a75d9e0d6838d4c52dc3ea2", coin: db.CoinTypeTWTBEP20, amount: 5704.79, address: "0xB26c83CA2d596671589992F08155C2BA3CBF89c1"},
			{txId: "0xab8323faedef8136ea5b4e6a11060750b3719812a802a534cf08ac829dd16e35", coin: db.CoinTypeAVAXBEP20, amount: 49.999149, address: "0x3457E41A9D5B3B0C92e8647dA56AE189DDf0f409"},
			{txId: "0xab437c20cc0690dad2f270905bcb30480113c657717d069dd17e5c211e7e0ea7", coin: db.CoinTypeCAKEBEP20, amount: 11.62, address: "0xebBB2558dEB063a514BEf5878F87B09C119bFA74"},
		}

		for _, v := range data {
			t.Run(string(v.coin), func(t *testing.T) {
				test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
					// Given
					q := db.New(dbConn).WithTx(tx)
					userId, _, _ := createUserWithBnbData(ctx, q)

					expectedTxId := v.txId
					expectedInvoice, err := q.CreateInvoice(ctx, db.CreateInvoiceParams{
						UserID:                userId,
						Coin:                  v.coin,
						CryptoAddress:         v.address,
						RequiredAmount:        v.amount,
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
					amount, err := verifyBNBTxHandler(ctx, q, &verifyTxHandlerData[listener.BNBTx]{invoice: expectedInvoice, tx: txs[0]})

					// Assert
					assert.NoError(t, err)
					assert.Equal(t, float64(0), amount)
				})
			})
		}
	})
}
