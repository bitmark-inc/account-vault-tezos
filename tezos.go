package tezos

import (
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	MAINNETChainID     = "NetXdQprcVkpaWU"
	ITHACANETChainID   = "NetXbhmtAbMukLc"
	HANGZHOUNETChainID = "NetXZSsxBpMQeAT"

	LiveMAINNET     Network = "MAINNET"
	TestITHACANET   Network = "ITHACANET"
	TestHANGZHOUNET Network = "HANGZHOUNET"
)

const DefaultDerivationPath = "m/44'/1729'/0'/0'"

type Wallet struct {
	chainID    string
	privateKey tezos.PrivateKey
	rpcClient  *rpc.Client
}

type Network string

func (n *Network) ChainID() string {
	switch *n {
	case LiveMAINNET:
		return MAINNETChainID
	case TestITHACANET:
		return ITHACANETChainID
	case TestHANGZHOUNET:
		return HANGZHOUNETChainID
	}
	return ""
}

// NewWallet creates a tezos wallet from a given seed
func NewWallet(seed []byte, network Network, rpcURL string) (*Wallet, error) {
	key := tezos.PrivateKey{
		Type: tezos.KeyTypeSecp256k1,
	}
	wallet, err := hdwallet.NewFromSeed(seed)
	if err != nil {
		return nil, err
	}

	path := hdwallet.MustParseDerivationPath(DefaultDerivationPath)
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}

	pk, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, err
	}
	key.Data = make([]byte, tezos.KeyTypeSecp256k1.SkHashType().Len())
	pk.D.FillBytes(key.Data)
	c, err := rpc.NewClient(rpcURL, nil)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		chainID:    network.ChainID(),
		privateKey: key,
		rpcClient:  c,
	}, nil
}

// RPCClient returns the Tezos RPC client which is bound to the wallet
func (w *Wallet) RPCClient() *rpc.Client {
	return w.rpcClient
}

// Account returns the tezos account address
func (w *Wallet) Account() string {
	return w.privateKey.Address().String()
}

// ChainID returns the tezos wallet ChainID
func (w *Wallet) ChainID() string {
	return w.chainID
}

// func (w *Wallet) Sign() error {

// }
