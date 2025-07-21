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

func createUserWithEthData(ctx context.Context, q *db.Queries) (pgtype.UUID, db.CryptoDatum, db.EthCryptoDatum) {
	userId, err := q.CreateUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	ethData, err := q.CreateETHCryptoData(ctx, "xpub6CUf84eg4Ba1jJ3ePzLSSoeQ1ENzP33zCN4982Xoi1TZ1kfYreZe5ECqLm4RVWQHpuB5gixi3gK1PykXzcwWxW7w6d7GWxpsNY7wxNVBHip")
	if err != nil {
		log.Fatal(err)
	}
	cd, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{EthID: ethData.ID, UserID: userId})
	if err != nil {
		log.Fatal(err)
	}

	return userId, cd, ethData
}

func createNewTestEthDaemon() *ethclient.Client {
	client, err := ethclient.Dial("https://ethereum.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func TestGenerateNextEthAddressHandler(t *testing.T) {
	t.Parallel()

	data := []struct {
		prevMajorIndex int32
		prevMinorIndex int32
		expectedAddr   string
	}{
		{
			prevMajorIndex: 0,
			prevMinorIndex: 0,
			expectedAddr:   "0x52bDE05866773a211aB01BbaEa9C474B9f24754D",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: 124,
			expectedAddr:   "0x947d04a4e66Ac7e26B4207D212b4F0903B09F89F",
		},
		{
			prevMajorIndex: 0,
			prevMinorIndex: math.MaxInt32,
			expectedAddr:   "0x2457A40c2C0D095Dcce88F5919F3afDc60e15CF9",
		},
		{
			prevMajorIndex: 1,
			prevMinorIndex: 2,
			expectedAddr:   "0x4d31b4584A64aD34C6AadB3388620682D2fd9D01",
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
				userId, cd, _ := createUserWithEthData(ctx, q)

				_, err := qT.UpdateIndicesETHCryptoDataById(ctx, test_db.UpdateIndicesETHCryptoDataByIdParams{ID: cd.EthID, LastMajorIndex: d.prevMajorIndex, LastMinorIndex: d.prevMinorIndex})
				if err != nil {
					log.Fatal(err)
				}

				// When
				addr, err := generateNextETHAddressHandler(ctx, q, &generateNextAddressHandlerData{userId: userId, network: listener.MainnetETH})

				// Assert
				assert.NoError(t, err)
				assert.Equal(t, d.expectedAddr, addr.Address)
			})
		})
	}

}

func TestVerifyETHTxHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConn, _, close := getPostgresWithDbConn()
	defer close(ctx)

	daemon := listener.NewSharedETHDaemonRpcClient(createNewTestEthDaemon())

	t.Run("Should Return Right Amount (Valid Tx)", func(t *testing.T) {
		data := []struct {
			txId    string
			coin    db.CoinType
			amount  float64
			address string
		}{
			{
				txId:    "0x4caafe1347589f252d2dedd009a5750f8dcb48c86840360a4adc066e691edcbd",
				coin:    db.CoinTypeETH,
				amount:  0.119996,
				address: "0x305c30dDc9DBCd1E831D8c894790AE0835B9D65d",
			},
			{
				txId:    "0x7f3cf60b639e426cf0423e2503f788341c8e1c8bb1fa5ca4e3163cf57723c9b3",
				coin:    db.CoinTypeUSDTERC20,
				amount:  2169.080917,
				address: "0x35df6C0ECA8AE63D489cd28ECfeA811fA8Fc5Bb1",
			},
			{
				txId:    "0x1a1626b7705b5ce3f3088ddc5dec6c49404f91f1d8b447ef811145a423dbfd0f",
				coin:    db.CoinTypeUSDCERC20,
				amount:  92.30606,
				address: "0xc1DA119E98158894F96Cf20C687F7D70B99Fc724",
			},
			{
				txId:    "0x604545cf837c51b6403c46c681a21f7fda461d4d4ec42373e865942d31cb65c3",
				coin:    db.CoinTypeDAIERC20,
				amount:  3260,
				address: "0x06Ac0C1C504218af3448E00ba1924455183D042C",
			},
			{
				txId:    "0x4c9b99f1b772c65f7c5f8afaca4e2ea756b2bf05b2d6ae787c041d6293a3ec8c",
				coin:    db.CoinTypeWBTCERC20,
				amount:  0.03719682,
				address: "0x6cC5F688a315f3dC28A7781717a9A798a59fDA7b",
			},
			{
				txId:    "0xa4c5c596de4f6a2ccfeafe9b2efb8c21e417c6e28c2ed3e4d7f0ff5d40af087a",
				coin:    db.CoinTypeUNIERC20,
				amount:  489.142016,
				address: "0xd5417e96Dd04363c675E41Ee6F30bF788412C719",
			},
			{
				txId:    "0x77b6d9dd4ac59b69aba72910e92cc50bfdc78dba2490a8883557b6e88c48918b",
				coin:    db.CoinTypeLINKERC20,
				amount:  1151.69485237,
				address: "0x59E0cDA5922eFbA00a57794faF09BF6252d64126",
			},
			{
				txId:    "0x1dde05711dd394ca61064c4c2c176abf39c91061428787ea348dfd1112d0181b",
				coin:    db.CoinTypeAAVEERC20,
				amount:  0.056373359965793961,
				address: "0x37F606d50815439DC163d79e379b0343889Bb480",
			},
			{
				txId:    "0x55b60ee6136991d342e6eb324d920f953b2ea74534f142c0df44a22a08a3f1c8",
				coin:    db.CoinTypeCRVERC20,
				amount:  25543.33,
				address: "0x28C6c06298d514Db089934071355E5743bf21d60",
			},
			{
				txId:    "0xe701832296c3f39e3b106e1dcfeaafcd9408664bb6d0cc284a39f12820fc1331",
				coin:    db.CoinTypeMATICERC20,
				amount:  2736.97309865,
				address: "0xe3b3233366961B2B926fe8aAbfa3C78382f8b997",
			},
			{
				txId:    "0x2f07144e97d625da9898b80530fedee3e7a9b47a05c7026862085378ee901311",
				coin:    db.CoinTypeSHIBERC20,
				amount:  95702450.003454334244215364,
				address: "0xaB5B038c647d2b4314BaC5b50004F39A913c122b",
			},
			{
				txId:    "0xab221809a9142c130b5ce706c17c8e244914cd5cf1146fc66b6ac0f65736638a",
				coin:    db.CoinTypeBNBERC20,
				amount:  0.039304267849757063,
				address: "0xC5fa84d2859AFCfDFB9D2c183F8A4E54F47051B8",
			},
			{
				txId:    "0x41477f74742757e3d74fb4a56f2f8461987281dbcffda34b0f77d73755622a9f",
				coin:    db.CoinTypeATOMERC20,
				amount:  1.455433,
				address: "0xF6FFC8f338213caC947426A0400df7B72Ad9408c",
			},
			{
				txId:    "0xe6e3fd45f3b650418562b1b650d9b9caffe99f459535dc17acf1917d256a82b6",
				coin:    db.CoinTypeARBERC20,
				amount:  76.8948,
				address: "0x0bEFf47d6A93D0dAd004B8383613df1bf0be1096",
			},
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
					amount, err := verifyETHBasedTxHandler(ctx, q, &verifyTxHandlerData[listener.ETHTx]{invoice: expectedInvoice, tx: txs[0]})

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
			{
				txId:    "0xcf7080e638bac6d921dd291bf03872683d5da7733bad5ea16d2e02566e6402d6",
				coin:    db.CoinTypeETH,
				amount:  0.119996,
				address: "0x305c30dDc9DBCd1E831D8c894790AE0835B9D65d",
			},
			{
				txId:    "0xcefde37e7f025a0bd9e4ad8282f05d945302738f3ea20329dacbe8274f781430",
				coin:    db.CoinTypeUSDTERC20,
				amount:  2169.080917,
				address: "0x35df6C0ECA8AE63D489cd28ECfeA811fA8Fc5Bb1",
			},
			{
				txId:    "0x88f44a4f8a8e9a603ebb50340ff671ba021907556409bb3c3cfc02d3b7272765",
				coin:    db.CoinTypeUSDCERC20,
				amount:  92.30606,
				address: "0xc1DA119E98158894F96Cf20C687F7D70B99Fc724",
			},
			{
				txId:    "0x4645d7db2b6b8e72b5878352eb4394a40e520a97736bb5817de21d020bb2104f",
				coin:    db.CoinTypeDAIERC20,
				amount:  3260,
				address: "0x06Ac0C1C504218af3448E00ba1924455183D042C",
			},
			{
				txId:    "0xf7f375e517acd55360c3d4d79302220988960e19dee33c478671cfa613163be2",
				coin:    db.CoinTypeWBTCERC20,
				amount:  0.03719682,
				address: "0x6cC5F688a315f3dC28A7781717a9A798a59fDA7b",
			},
			{
				txId:    "0x2d2d02b4d212cf51aff4008beebac54eff9d867709b733e4b9a7c460d9a96d07",
				coin:    db.CoinTypeUNIERC20,
				amount:  489.142016,
				address: "0xd5417e96Dd04363c675E41Ee6F30bF788412C719",
			},
			{
				txId:    "0xaf597ca3eb3470db4c605cfef71dc48ddd4b64acbe50af29ae9cd3c026b65772",
				coin:    db.CoinTypeLINKERC20,
				amount:  1151.69485237,
				address: "0x59E0cDA5922eFbA00a57794faF09BF6252d64126",
			},
			{
				txId:    "0xc81197f55426c0d44ebb7c36dbab4950356350f71f834711abb33f277dd0522b",
				coin:    db.CoinTypeAAVEERC20,
				amount:  0.056373359965793961,
				address: "0x37F606d50815439DC163d79e379b0343889Bb480",
			},
			{
				txId:    "0x2eff85c61ee573c2d91b327231e4b8b90f92e60282359bd6f9f2847e12134f86",
				coin:    db.CoinTypeCRVERC20,
				amount:  25543.33,
				address: "0x28C6c06298d514Db089934071355E5743bf21d60",
			},
			{
				txId:    "0xd522f05195e21f29be58588bbb25cd6fd89be2daece91d2fb1d99cc3daed623b",
				coin:    db.CoinTypeMATICERC20,
				amount:  2736.97309865,
				address: "0xe3b3233366961B2B926fe8aAbfa3C78382f8b997",
			},
			{
				txId:    "0x746a74de08c5742452c17fe8f95b92d49245c2a667a933a714c523c4fa05b090",
				coin:    db.CoinTypeSHIBERC20,
				amount:  95702450.003454334244215364,
				address: "0xaB5B038c647d2b4314BaC5b50004F39A913c122b",
			},
			{
				txId:    "0x367ce2a4e33c171a1a894d33e34afa09e615924a12967c33a83f8a2e4bc6d324",
				coin:    db.CoinTypeBNBERC20,
				amount:  0.039304267849757063,
				address: "0xC5fa84d2859AFCfDFB9D2c183F8A4E54F47051B8",
			},
			{
				txId:    "0x087f4f7314985afcc7a2b6c6395e4cdb9f1be54a6e58fcfabc6df7e148ec0f26",
				coin:    db.CoinTypeATOMERC20,
				amount:  1.455433,
				address: "0xF6FFC8f338213caC947426A0400df7B72Ad9408c",
			},
			{
				txId:    "0x25370db89f39c3f2f4bea1555482087e57a77f51a19019e382b493a4ca387087",
				coin:    db.CoinTypeARBERC20,
				amount:  76.8948,
				address: "0x0bEFf47d6A93D0dAd004B8383613df1bf0be1096",
			},
		}

		for _, v := range data {
			t.Run(string(v.coin), func(t *testing.T) {
				test.RunInTransaction(t, dbConn, func(t *testing.T, tx pgx.Tx) {
					// Given
					q := db.New(dbConn).WithTx(tx)
					userId, _, _ := createUserWithEthData(ctx, q)

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
					amount, err := verifyETHBasedTxHandler(ctx, q, &verifyTxHandlerData[listener.ETHTx]{invoice: expectedInvoice, tx: txs[0]})

					// Assert
					assert.NoError(t, err)
					assert.Equal(t, float64(0), amount)
				})
			})
		}
	})
}
