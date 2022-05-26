package tezos

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	ed25519hd "github.com/bitmark-inc/go-ed25519-hd"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

const (
	DefaultAccountIndex = 0
	MAINNETChainID      = "NetXdQprcVkpaWU"
	ITHACANETChainID    = "NetXnHfVqm9iesp"
)

var (
	ErrWrongChainID   = errors.New("Connected node serve different chain from setting")
	ErrInvalidRpcNode = errors.New("Invalid rpc node")
	ErrSignFailed     = errors.New("Failed to sign with provided data")
	ErrInvalidAddress = errors.New("Invalid address provided")
	ErrInvalidTokenID = errors.New("Invalid tokenID provided")
)

func buildDerivePath(index uint) string {
	return fmt.Sprintf("m/44'/1729'/%d'/0'", index)
}

type Wallet struct {
	chainID      string
	masterKey    ed25519hd.PrivateKey
	privateKey   tezos.PrivateKey
	accountIndex uint
	rpcClient    *rpc.Client
}

// NewWallet creates a tezos wallet from a given seed
func NewWallet(seed []byte, network string, rpcURL string) (*Wallet, error) {
	pk, err := ed25519hd.GetMasterKeyFromSeed(seed)
	if err != nil {
		return nil, err
	}

	dpk, _ := pk.DeriveChildPrivateKey(buildDerivePath(DefaultAccountIndex))
	key := toTzgoPrivateKey(*dpk)

	c, _ := rpc.NewClient(rpcURL, nil)
	err = c.Init(context.Background())
	if err != nil {
		return nil, ErrInvalidRpcNode
	}

	chainID := ITHACANETChainID
	if network == "livenet" {
		chainID = MAINNETChainID
	}
	cChainID, _ := tezos.ParseChainIdHash(chainID)
	if !c.ChainId.Equal(cChainID) {
		return nil, ErrWrongChainID
	}

	return &Wallet{
		chainID:      chainID,
		masterKey:    *pk,
		privateKey:   key,
		accountIndex: DefaultAccountIndex,
		rpcClient:    c,
	}, nil
}

// DeriveAccount derive the specific index account from the master key
func (w *Wallet) DeriveAccount(index uint) (*Wallet, error) {
	dpk, err := w.masterKey.DeriveChildPrivateKey(buildDerivePath(index))
	if err != nil {
		return nil, err
	}
	key := toTzgoPrivateKey(*dpk)
	return &Wallet{
		chainID:      w.chainID,
		masterKey:    w.masterKey,
		privateKey:   key,
		accountIndex: index,
		rpcClient:    w.rpcClient,
	}, nil
}

// SignMessage sign a specific message from privateKey
func (w *Wallet) SignMessage(message []byte) (string, error) {
	dm := tezos.Digest(message)
	sig, err := w.privateKey.Sign(dm[:])
	if err != nil {
		return "", ErrSignFailed
	}
	return sig.Generic(), nil
}

// SignAuthTransferMessage sign the authorized transfer message from privateKey
func (w *Wallet) SignAuthTransferMessage(to, tokenID string) (string, error) {
	m, err := BuildAuthTransferMessage(to, tokenID)
	if err != nil {
		return "", err
	}
	return w.SignMessage(m)
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

// BuildAuthTransferMessage build the authorized transfer message
func BuildAuthTransferMessage(to, tokenID string) ([]byte, error) {
	// timestamp
	ts := big.NewInt(time.Now().Unix())

	// address
	ad, err := tezos.ParseAddress(to)
	if err != nil {
		return nil, ErrInvalidAddress
	}

	// token
	n := new(big.Int)
	tk, ok := n.SetString(tokenID, 10)
	if !ok {
		return nil, ErrInvalidTokenID
	}

	tsp := micheline.Prim{
		Type: micheline.PrimInt,
		Int:  ts,
	}
	adp := micheline.Prim{
		Type:  micheline.PrimBytes,
		Bytes: ad.Bytes22(),
	}
	tkp := micheline.Prim{
		Type: micheline.PrimInt,
		Int:  tk,
	}

	// return hex.EncodeToString(append(append(tsp.Pack(), adp.Pack()...), tkp.Pack()...)), nil
	return append(append(tsp.Pack(), adp.Pack()...), tkp.Pack()...), nil
}

// convert an ed25519 hd private key to tzgo private key
func toTzgoPrivateKey(edk ed25519hd.PrivateKey) tezos.PrivateKey {
	key := tezos.PrivateKey{
		Type: tezos.KeyTypeEd25519,
	}
	key.Data = append(edk.Key, edk.GetPublicKey()...)
	return key
}
