package processor

import (
	"context"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/listener"
	"github.com/ethereum/go-ethereum/common"
)

const (
	transferMethodSignatureETHCompatible string = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

type tokenData struct {
	contractAddress string
	decimals        uint64
}

type indices struct {
	major uint32
	minor uint32
}

var (
	tokenDataETHCompatible map[db.CoinType]map[db.CoinType]tokenData = map[db.CoinType]map[db.CoinType]tokenData{
		// ERC20
		db.CoinTypeETH: {
			db.CoinTypeUSDTERC20:  {contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7", decimals: 10e5},
			db.CoinTypeUSDCERC20:  {contractAddress: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", decimals: 10e5},
			db.CoinTypeDAIERC20:   {contractAddress: "0x6B175474E89094C44Da98b954EedeAC495271d0F", decimals: 10e17},
			db.CoinTypeWBTCERC20:  {contractAddress: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", decimals: 10e7},
			db.CoinTypeUNIERC20:   {contractAddress: "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", decimals: 10e17},
			db.CoinTypeLINKERC20:  {contractAddress: "0x514910771AF9Ca656af840dff83E8264EcF986CA", decimals: 10e17},
			db.CoinTypeAAVEERC20:  {contractAddress: "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9", decimals: 10e17},
			db.CoinTypeCRVERC20:   {contractAddress: "0xD533a949740bb3306d119CC777fa900bA034cd52", decimals: 10e17},
			db.CoinTypeMATICERC20: {contractAddress: "0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0", decimals: 10e17},
			db.CoinTypeSHIBERC20:  {contractAddress: "0x95aD61b0a150d79219dCF64E1E6Cc01f0B64C4cE", decimals: 10e17},
			db.CoinTypeBNBERC20:   {contractAddress: "0xB8c77482e45F1F44dE1745F52C74426C631bDD52", decimals: 10e17},
			db.CoinTypeATOMERC20:  {contractAddress: "0x8D983cb9388EaC77af0474fA441C4815500Cb7BB", decimals: 10e5},
			db.CoinTypeARBERC20:   {contractAddress: "0xB50721BCf8d664c30412Cfbc6cf7a15145234ad1", decimals: 10e17},
		},

		// BEP20
		db.CoinTypeBNB: {
			db.CoinTypeBSCUSDBEP20: {contractAddress: "0x55d398326f99059fF775485246999027B3197955", decimals: 10e17},
			db.CoinTypeUSDCBEP20:   {contractAddress: "0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d", decimals: 10e17},
			db.CoinTypeDAIBEP20:    {contractAddress: "0x1AF3F329e8BE154074D8769D1FFa4eE058B1DBc3", decimals: 10e17},
			db.CoinTypeWBTCBEP20:   {contractAddress: "0x0555E30da8f98308EdB960aa94C0Db47230d2B9c", decimals: 10e7},
			db.CoinTypeUNIBEP20:    {contractAddress: "0xBf5140A22578168FD562DCcF235E5D43A02ce9B1", decimals: 10e17},
			db.CoinTypeLINKBEP20:   {contractAddress: "0xF8A0BF9cF54Bb92F17374d9e9A321E6a111a51bD", decimals: 10e17},
			db.CoinTypeAAVEBEP20:   {contractAddress: "0xfb6115445Bff7b52FeB98650C87f44907E58f802", decimals: 10e17},
			db.CoinTypeMATICBEP20:  {contractAddress: "0xCC42724C6683B7E57334c4E856f4c9965ED682bD", decimals: 10e17},
			db.CoinTypeSHIBBEP20:   {contractAddress: "0x2859e4544C4bB03966803b044A93563Bd2D0DD4D", decimals: 10e17},
			db.CoinTypeBUSDBEP20:   {contractAddress: "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56", decimals: 10e17},
			db.CoinTypeATOMBEP20:   {contractAddress: "0x0Eb3a705fc54725037CC9e008bDede697f62F335", decimals: 10e17},
			db.CoinTypeARBBEP20:    {contractAddress: "0xa050FFb3eEb8200eEB7F61ce34FF644420FD3522", decimals: 10e17},
			db.CoinTypeETHBEP20:    {contractAddress: "0x2170Ed0880ac9A755fd29B2688956BD959F933F8", decimals: 10e17},
			db.CoinTypeXRPBEP20:    {contractAddress: "0x1D2F0da169ceB9fC7B3144628dB156f3F6c60dBE", decimals: 10e17},
			db.CoinTypeADABEP20:    {contractAddress: "0x3EE2200Efb3400fAbB9AacF31297cBdD1d435D47", decimals: 10e17},
			db.CoinTypeTRXBEP20:    {contractAddress: "0xCE7de646e7208a4Ef112cb6ed5038FA6cC6b12e3", decimals: 10e5},
			db.CoinTypeDOGEBEP20:   {contractAddress: "0xbA2aE424d960c26247Dd6c32edC70B295c744C43", decimals: 10e7},
			db.CoinTypeLTCBEP20:    {contractAddress: "0x4338665CBB7B2485A8855A139b75D5e34AB0DB94", decimals: 10e17},
			db.CoinTypeBCHBEP20:    {contractAddress: "0x8fF795a6F4D97E7887C79beA79aba5cc76444aDf", decimals: 10e17},
			db.CoinTypeTWTBEP20:    {contractAddress: "0x4B0F1812e5Df2A09796481Ff14017e6005508003", decimals: 10e17},
			db.CoinTypeAVAXBEP20:   {contractAddress: "0x1CE0c2827e2eF14D5C4f29a091d735A204794041", decimals: 10e17},
			db.CoinTypeCAKEBEP20:   {contractAddress: "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82", decimals: 10e17},
		},
	}
)

func verifyETHBasedTxHandler(ctx context.Context, q *db.Queries, data *verifyTxHandlerData[listener.ETHTx]) (float64, error) {
	var amount float64 = 0

	if data.tx.IsDoubleSpendSeen() {
		return amount, nil
	}

	logsCount := len(data.tx.Logs)
	if logsCount > 0 {
		tokenData, ok := func() (tokenData, bool) {
			for _, v := range tokenDataETHCompatible {
				tokenData, ok := v[data.invoice.Coin]
				if ok {
					return tokenData, ok
				}
			}
			return tokenData{}, false
		}()
		if !ok {
			return amount, nil
		}

		for i := 0; i < logsCount; i++ {
			log := data.tx.Logs[i]
			if len(log.Topics) < 3 ||
				log.Topics[0].Hex() != transferMethodSignatureETHCompatible ||
				log.Address.Hex() != tokenData.contractAddress {
				continue
			}

			if common.BytesToAddress(log.Topics[2].Bytes()).Hex() == data.invoice.CryptoAddress {
				am, _ := new(big.Float).Quo(
					new(big.Float).SetInt(new(big.Int).SetBytes(log.Data)),
					new(big.Float).SetInt(new(big.Int).SetUint64(tokenData.decimals)),
				).Float64()
				amount += am
			}
		}
	} else if toAddr := data.tx.Tx.To(); toAddr != nil && data.invoice.CryptoAddress == toAddr.Hex() {
		amount += float64(data.tx.Tx.Value().Uint64()) / 1e18
	}

	return amount, nil
}

func deriveNextETHBasedECPubKeyHelper(indices indices, masterPubKey string) (*btcec.PublicKey, error) {
	mPub, err := hdkeychain.NewKeyFromString(masterPubKey)
	if err != nil {
		return nil, err
	}

	majMPub, err := mPub.Derive(indices.major)
	if err != nil {
		return nil, err
	}
	minMPub, err := majMPub.Derive(indices.minor)
	if err != nil {
		return nil, err
	}

	return minMPub.ECPubKey()
}
