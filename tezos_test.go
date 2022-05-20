package tezos

import (
	"encoding/hex"
	"testing"

	ed25519hd "github.com/bitmark-inc/go-ed25519-hd"
	"github.com/stretchr/testify/assert"
)

type wallet struct {
	seed       string
	account    string
	rpcURL     string
	network    string
	index      uint
	derivePath string
}

func TestNewWallet(t *testing.T) {
	for _, w := range testWallet() {
		s, _ := hex.DecodeString(w.seed)
		nw, err := NewWallet(s, w.network, w.rpcURL)
		assert.Nil(t, err)
		assert.EqualValues(t, w.account, nw.Account())
		assert.EqualValues(t, nw.rpcClient, nw.RPCClient())
		assert.EqualValues(t, nw.chainID, nw.ChainID())
	}

	for _, ww := range testWrongNetworkWallet() {
		s, _ := hex.DecodeString(ww.seed)
		_, err := NewWallet(s, ww.network, ww.rpcURL)
		assert.EqualError(t, err, ErrWrongChainID.Error())
	}

	for _, ww := range testWrongSeedSizeWallet() {
		s, _ := hex.DecodeString(ww.seed)
		_, err := NewWallet(s, ww.network, ww.rpcURL)
		assert.EqualError(t, err, ed25519hd.ErrWrongSeedSize.Error())
	}

	for _, ww := range testWrongRpcNodeWallet() {
		s, _ := hex.DecodeString(ww.seed)
		_, err := NewWallet(s, ww.network, ww.rpcURL)
		assert.EqualError(t, err, ErrInvalidRpcNode.Error())
	}
}

func TestDeriveAccount(t *testing.T) {
	for _, w := range testDeriveWallet() {
		s, _ := hex.DecodeString(w.seed)
		nw, err := NewWallet(s, w.network, w.rpcURL)
		assert.Nil(t, err)
		dw, err := nw.DeriveAccount(w.index)
		assert.Nil(t, err)
		assert.EqualValues(t, w.account, dw.Account())
	}

	for _, ww := range testWrongDerivePathWallet() {
		s, _ := hex.DecodeString(ww.seed)
		nw, err := NewWallet(s, ww.network, ww.rpcURL)
		assert.Nil(t, err)
		_, err = nw.DeriveAccount(ww.index)
		assert.EqualError(t, err, ed25519hd.ErrWrongDerivePath.Error())
	}
}

func testWallet() []wallet {
	return []wallet{
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1TFmv27hNN1CV4XFP5TceGzsmDCrWTdWpd",
			rpcURL:  "https://ithacanet.ecadinfra.com/",
			network: "testnet",
			index:   0,
		},
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1TFmv27hNN1CV4XFP5TceGzsmDCrWTdWpd",
			rpcURL:  "https://mainnet-node.madfish.solutions",
			network: "livenet",
			index:   0,
		},
	}
}

func testDeriveWallet() []wallet {
	return []wallet{
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1TFmv27hNN1CV4XFP5TceGzsmDCrWTdWpd",
			rpcURL:  "https://ithacanet.ecadinfra.com/",
			network: "testnet",
			index:   0,
		},
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1b4FWeKgXysDkeeHMaxy516PXB3Lni6Rpa",
			rpcURL:  "https://mainnet-node.madfish.solutions",
			network: "livenet",
			index:   1,
		},
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1ZRqZEaiwyrMGtZDfxhtMjqijaNy5oFpgK",
			rpcURL:  "https://mainnet-node.madfish.solutions",
			network: "livenet",
			index:   2,
		},
	}
}

func testWrongNetworkWallet() []wallet {
	return []wallet{
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1TFmv27hNN1CV4XFP5TceGzsmDCrWTdWpd",
			rpcURL:  "https://ithacanet.ecadinfra.com/",
			network: "livenet",
		},
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			account: "tz1TFmv27hNN1CV4XFP5TceGzsmDCrWTdWpd",
			rpcURL:  "https://mainnet-node.madfish.solutions",
			network: "testnet",
		},
	}
}

func testWrongSeedSizeWallet() []wallet {
	return []wallet{
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707feded",
			rpcURL:  "https://ithacanet.ecadinfra.com/",
			network: "livenet",
		},
		{
			seed:    "063cafb67",
			rpcURL:  "https://mainnet-node.madfish.solutions",
			network: "testnet",
		},
	}
}

func testWrongRpcNodeWallet() []wallet {
	return []wallet{
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			rpcURL:  "https://google.com",
			network: "livenet",
		},
	}
}

func testWrongDerivePathWallet() []wallet {
	return []wallet{
		{
			seed:    "063cafb67a29cb2c567a4ecba7edc856a54403952272bffd492caaf9095a9442b208d9f0d2b75a7b1cda59819c245949b9d7e4826e7ace8e19a970a080707fed",
			rpcURL:  "https://mainnet-node.madfish.solutions",
			network: "livenet",
			index:   2147483648,
		},
	}
}
