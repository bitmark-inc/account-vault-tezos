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
	DefaultSignPrefix   = "Tezos Signed Message:"
)

var (
	ErrWrongChainID                  = errors.New("Connected node serve different chain from setting")
	ErrInvalidRpcNode                = errors.New("Invalid rpc node")
	ErrSignFailed                    = errors.New("Failed to sign with provided data")
	ErrInvalidAddress                = errors.New("Invalid address provided")
	ErrInvalidTimestamp              = errors.New("Invalid timestamp provided")
	ErrInvalidPublicKey              = errors.New("Invalid public key provided")
	ErrInvalidSignature              = errors.New("Invalid signature provided")
	ErrInvalidTokenID                = errors.New("Invalid tokenID provided")
	ErrTransferAmountLowerThanSetFee = errors.New("Transfer amount lower than set fee")
	ErrExceedSettingFee              = errors.New("Actual cost is more than setting")
)

func buildDerivePath(index uint) string {
	return fmt.Sprintf("m/44'/1729'/%d'/0'", index)
}

type Wallet struct {
	chainID      tezos.ChainIdHash
	masterKey    ed25519hd.PrivateKey
	privateKey   tezos.PrivateKey
	accountIndex uint
	rpcClient    *rpc.Client
}

type TransferXTZParam struct {
	To     string
	Amount int64
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

	// chainID := GHOSTNETChainID
	if network == "livenet" {
		if !c.ChainId.Equal(tezos.DefaultParams.ChainId) {
			return nil, ErrWrongChainID
		}
	}

	return &Wallet{
		chainID:      c.ChainId,
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
func (w *Wallet) SignAuthTransferMessage(to, contractAddress, tokenID string, expiry time.Time) (string, error) {
	// timestamp
	ts := big.NewInt(expiry.Unix())

	// address
	ad, err := tezos.ParseAddress(to)
	if err != nil {
		return "", ErrInvalidAddress
	}

	contractAddr, err := tezos.ParseAddress(contractAddress)
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

	ctp := micheline.Prim{
		Type:  micheline.PrimBytes,
		Bytes: contractAddr.Bytes22(),
	}

	adp := micheline.Prim{
		Type:  micheline.PrimBytes,
		Bytes: ad.Bytes22(),
	}
	tkp := micheline.Prim{
		Type: micheline.PrimInt,
		Int:  tk,
	}

	m := append(append(append(tsp.Pack(), ctp.Pack()...), adp.Pack()...), tkp.Pack()...)
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

	if w.chainID.Equal(tezos.GhostnetParams.ChainId) {
		op.WithParams(tezos.GhostnetParams)
	} else if w.chainID.Equal(tezos.JakartanetParams.ChainId) {
		op.WithParams(tezos.JakartanetParams)
	} else if w.chainID.Equal(tezos.KathmandunetParams.ChainId) {
		op.WithParams(tezos.KathmandunetParams)
	} else {
		op.WithParams(tezos.DefaultParams)
	}

	return w.send(op, opts)
}

// SendOperations will send list of operations to tezos blockchain and return hash
func (w *Wallet) SendOperations(ops []codec.Operation) (*string, error) {
	opts := &rpc.CallOptions{
		TTL:    tezos.DefaultParams.MaxOperationsTTL - 2,
		MaxFee: 10_000_000,
	}

	op := codec.NewOp().WithTTL(opts.TTL)
	for _, o := range ops {
		op.WithContents(o)
	}

	if w.chainID.Equal(tezos.GhostnetParams.ChainId) {
		op.WithParams(tezos.GhostnetParams)
	} else if w.chainID.Equal(tezos.JakartanetParams.ChainId) {
		op.WithParams(tezos.JakartanetParams)
	} else if w.chainID.Equal(tezos.KathmandunetParams.ChainId) {
		op.WithParams(tezos.KathmandunetParams)
	} else {
		op.WithParams(tezos.DefaultParams)
	}

	return w.send(op, opts)
}

func (w *Wallet) SimulateXTZTransferFee(txs []TransferXTZParam) (*int64, error) {
	opts := &rpc.CallOptions{
		TTL: tezos.DefaultParams.MaxOperationsTTL - 2,
	}
	op := codec.NewOp()
	for _, tx := range txs {
		ad, err := tezos.ParseAddress(tx.To)
		if err != nil {
			return nil, ErrInvalidAddress
		}
		// construct a transfer operation
		op.WithTransfer(ad, tx.Amount)
	}

	ctx := context.Background()

	signer := w.rpcClient.Signer

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

	bufferFee := int64(0)
	// simulate to estimate cost
	sim, err := w.rpcClient.Simulate(ctx, op, opts)
	if err != nil {
		// FIXME: hack around way to get the estimate fee using the lowest transfer amount
		// have to find out a better way like include the tzStats SDK to get the current balance of account
		// we don't have a better way to get the est. fee for tezos in tzgo now if the transfer amount is almost(same) as the entire account balance

		// New Op
		op = codec.NewOp()

		// construct a transfer operation with 0.000001 xtz to do brief fee estimation
		for _, tx := range txs {
			ad, err := tezos.ParseAddress(tx.To)
			if err != nil {
				return nil, ErrInvalidAddress
			}
			op.WithTransfer(ad, 1)
		}

		// set source on all ops
		op.WithSource(key.Address())

		// auto-complete op with branch/ttl, source counter, reveal
		err = w.rpcClient.Complete(ctx, op, key)
		if err != nil {
			return nil, err
		}

		// simulate to estimate cost
		sim, err = w.rpcClient.Simulate(ctx, op, opts)
		if err != nil {
			return nil, err
		}

		bufferFee = int64(1000)
	}

	op.WithLimits(sim.MinLimits(), rpc.GasSafetyMargin)

	c := sim.TotalCosts()
	tc := c.Burn + op.Limits().Fee + bufferFee

	return &tc, nil
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
	return w.chainID.Hash.String()
}

// Account returns the private key
func (w *Wallet) PrivateKey() tezos.PrivateKey {
	return w.privateKey
}

// TransferXTZ transfer the xtz to destination
func (w *Wallet) TransferXTZ(to string, amount int64) (*string, error) {
	return w.BatchTransferXTZ(
		[]TransferXTZParam{
			{
				To:     to,
				Amount: amount,
			},
		},
	)
}

// BatchTransferXTZ transfer the xtz to destinations
func (w *Wallet) BatchTransferXTZ(txs []TransferXTZParam) (*string, error) {
	opts := &rpc.CallOptions{
		TTL:    tezos.DefaultParams.MaxOperationsTTL - 2,
		MaxFee: 1_000_000,
	}
	op := codec.NewOp()
	for _, tx := range txs {
		ad, err := tezos.ParseAddress(tx.To)
		if err != nil {
			return nil, ErrInvalidAddress
		}
		// construct a transfer operation
		op.WithTransfer(ad, tx.Amount)
	}

	return w.send(op, opts)
}

// convert an ed25519 hd private key to tzgo private key
func toTzgoPrivateKey(edk ed25519hd.PrivateKey) tezos.PrivateKey {
	key := tezos.PrivateKey{
		Type: tezos.KeyTypeEd25519,
	}
	key.Data = append(edk.Key, edk.GetPublicKey()...)
	return key
}
