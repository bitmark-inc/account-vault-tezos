package tezos

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	ed25519hd "github.com/bitmark-inc/go-ed25519-hd"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
)

const (
	DefaultAccountIndex = 0
	MAINNETChainID      = "NetXdQprcVkpaWU"
	ITHACANETChainID    = "NetXnHfVqm9iesp"
	DefaultSignPrefix   = "Tezos Signed Message:"
)

var (
	ErrWrongChainID     = errors.New("Connected node serve different chain from setting")
	ErrInvalidRpcNode   = errors.New("Invalid rpc node")
	ErrSignFailed       = errors.New("Failed to sign with provided data")
	ErrInvalidAddress   = errors.New("Invalid address provided")
	ErrInvalidTimestamp = errors.New("Invalid timestamp provided")
	ErrInvalidPublicKey = errors.New("Invalid public key provided")
	ErrInvalidSignature = errors.New("Invalid signature provided")
	ErrInvalidTokenID   = errors.New("Invalid tokenID provided")
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

	c, err := rpc.NewClient(rpcURL, nil)
	if err != nil {
		return nil, err
	}

	// Set default signer to wallet private key
	c.Signer = signer.NewFromKey(key)

	if err := c.Init(context.Background()); err != nil {
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

// signMessage sign a specific message from privateKey
func (w *Wallet) signMessage(message []byte) (string, error) {
	// force add prefix to message to prevent possible attack
	m := append([]byte(DefaultSignPrefix), message...)
	// pack the message to tezos bytes
	mp := micheline.Prim{
		Type:  micheline.PrimBytes,
		Bytes: m,
	}
	dm := tezos.Digest(mp.Pack())
	sig, err := w.privateKey.Sign(dm[:])
	if err != nil {
		return "", ErrSignFailed
	}
	return sig.Generic(), nil
}

// SignAuthTransferMessage sign the authorized transfer message from privateKey
func (w *Wallet) SignAuthTransferMessage(to, tokenID string, expiry time.Time) (string, error) {
	// timestamp
	ts := big.NewInt(expiry.Unix())

	// address
	ad, err := tezos.ParseAddress(to)
	if err != nil {
		return "", ErrInvalidAddress
	}

	// token
	tk, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		return "", ErrInvalidTokenID
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

	m := append(append(tsp.Pack(), adp.Pack()...), tkp.Pack()...)
	return w.signMessage(m)
}

// Send will send a op to tezos blockchain and return hash
func (w *Wallet) Send(args contract.CallArguments) (*string, error) {
	opts := &rpc.CallOptions{
		TTL:    tezos.DefaultParams.MaxOperationsTTL - 2,
		MaxFee: 10_000_000,
	}

	op := codec.NewOp().WithTTL(opts.TTL)
	op.WithContents(args.Encode())

	if w.chainID == ITHACANETChainID {
		op.WithParams(tezos.IthacanetParams)
	} else {
		op.WithParams(tezos.DefaultParams)
	}

	return w.send(op, opts)
}

// send is a convenience wrapper for sending operations. It auto-completes gas and storage limit,
// ensures minimum fees are set, protects against fee overpayment, signs and broadcasts the final
// operation.
func (w *Wallet) send(op *codec.Op, opts *rpc.CallOptions) (*string, error) {
	ctx := context.Background()
	if opts == nil {
		opts = &rpc.DefaultOptions
	}

	signer := w.rpcClient.Signer
	if opts.Signer != nil {
		signer = opts.Signer
	}

	// identify the sender address for signing the message
	addr := opts.Sender
	if !addr.IsValid() {
		addrs, err := signer.ListAddresses(ctx)
		if err != nil {
			return nil, err
		}
		addr = addrs[0]
	}

	key, err := signer.GetKey(ctx, addr)
	if err != nil {
		return nil, err
	}

	// set source on all ops
	op.WithSource(key.Address())

	// auto-complete op with branch/ttl, source counter, reveal
	err = w.rpcClient.Complete(ctx, op, key)
	if err != nil {
		return nil, err
	}

	// simulate to check tx validity and estimate cost
	sim, err := w.rpcClient.Simulate(ctx, op, opts)
	if err != nil {
		return nil, err
	}

	// fail with Tezos error when simulation failed
	if !sim.IsSuccess() {
		return nil, sim.Error()
	}

	// apply simulated cost as limits to tx list
	if !opts.IgnoreLimits {
		op.WithLimits(sim.MinLimits(), rpc.GasSafetyMargin)
	}

	// check minFee calc against maxFee if set
	if opts.MaxFee > 0 {
		if l := op.Limits(); l.Fee > opts.MaxFee {
			return nil, fmt.Errorf("estimated cost %d > max %d", l.Fee, opts.MaxFee)
		}
	}

	// sign digest
	sig, err := signer.SignOperation(ctx, addr, op)
	if err != nil {
		return nil, err
	}
	op.WithSignature(sig)

	// broadcast
	hash, err := w.rpcClient.Broadcast(ctx, op)
	if err != nil {
		return nil, err
	}
	h := hash.String()
	return &h, nil
}

// RPCClient returns the Tezos RPC client which is bound to the wallet
func (w *Wallet) RPCClient() *rpc.Client {
	return w.rpcClient
}

// Account returns the tezos account address string
func (w *Wallet) Account() string {
	return w.privateKey.Address().String()
}

// ChainID returns the tezos wallet ChainID
func (w *Wallet) ChainID() string {
	return w.chainID
}

// Account returns the private key
func (w *Wallet) PrivateKey() tezos.PrivateKey {
	return w.privateKey
}

// convert an ed25519 hd private key to tzgo private key
func toTzgoPrivateKey(edk ed25519hd.PrivateKey) tezos.PrivateKey {
	key := tezos.PrivateKey{
		Type: tezos.KeyTypeEd25519,
	}
	key.Data = append(edk.Key, edk.GetPublicKey()...)
	return key
}
