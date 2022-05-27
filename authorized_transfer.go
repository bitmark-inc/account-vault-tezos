package tezos

import (
	"math/big"
	"time"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

type AuthTransferMessageParam struct {
	To        string
	TokenID   string
	Timestamp time.Time
}

func (a AuthTransferMessageParam) Build() ([]byte, error) {
	// timestamp
	ts := big.NewInt(a.Timestamp.Unix())

	// address
	ad, err := tezos.ParseAddress(a.To)
	if err != nil {
		return nil, ErrInvalidAddress
	}

	// token
	tk, ok := new(big.Int).SetString(a.TokenID, 10)
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

	return append(append(tsp.Pack(), adp.Pack()...), tkp.Pack()...), nil
}

type AuthTransferParam struct {
	From      string
	PK        string
	Timestamp time.Time
	Txs       []AuthTransaction
}

func (a AuthTransferParam) Build() (*authTransferParam, error) {
	from_, err := tezos.ParseAddress(a.From)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	pk_, err := tezos.ParseKey(a.PK)
	if err != nil {
		return nil, ErrInvalidPublicKey
	}
	var txs []authTransaction
	for _, tx := range a.Txs {
		x, err := tx.Build()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *x)
	}
	return &authTransferParam{
		From:      from_,
		PK:        pk_,
		Timestamp: big.NewInt(a.Timestamp.Unix()),
		Txs:       txs,
	}, nil
}

type AuthTransaction struct {
	To        string
	Signature string
	TokenID   string
}

func (a AuthTransaction) Build() (*authTransaction, error) {
	sig_, err := tezos.ParseSignature(a.Signature)
	if err != nil {
		return nil, ErrInvalidSignature
	}
	tk, ok := new(big.Int).SetString(a.TokenID, 10)
	if !ok {
		return nil, ErrInvalidTokenID
	}
	to_, err := tezos.ParseAddress(a.To)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	return &authTransaction{
		Signature: sig_,
		TokenID:   tk,
		To:        to_,
		Amount:    big.NewInt(1),
	}, nil
}

type authTransferParam struct {
	From      tezos.Address
	PK        tezos.Key
	Timestamp *big.Int
	Txs       []authTransaction
}

type authTransaction struct {
	To        tezos.Address
	Signature tezos.Signature
	Amount    *big.Int
	TokenID   *big.Int
}

type authTransferArgs struct {
	contract.TxArgs
	Transfers []authTransferParam
}

var _ contract.CallArguments = (*authTransferArgs)(nil)

func (p authTransferParam) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for _, v := range p.Txs {
		rs.Args = append(rs.Args,
			micheline.NewPair(
				micheline.NewBytes(v.To.Bytes22()),
				micheline.NewPair(
					micheline.NewBig(v.TokenID),
					micheline.NewPair(
						micheline.NewNat(v.Amount),
						micheline.NewBytes(v.Signature.Bytes()),
					),
				),
			),
		)
	}
	return rs
}

func (p authTransferArgs) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for i, v := range p.Transfers {
		rs.Args = append(rs.Args,
			micheline.NewPair(
				micheline.NewBytes(v.From.Bytes22()),
				micheline.NewPair(
					micheline.NewBytes(v.PK.Bytes()),
					micheline.NewPair(
						micheline.NewBig(v.Timestamp),
						micheline.NewSeq(),
					),
				),
			),
		)
		rs.Args[i].Args[1].Args[1].Args[1] = v.Prim()
	}
	return rs
}

// SignAuthTransferMessage sign the authorized transfer message from privateKey
func (w *Wallet) SignAuthTransferMessage(am AuthTransferMessageParam) (string, error) {
	m, err := am.Build()
	if err != nil {
		return "", err
	}
	return w.SignMessage(m)
}

// AuthTransfer call the authorized transfer entrypoint define in FeralFile contract
func (w *Wallet) AuthTransfer(contr string, aps []AuthTransferParam) (*rpc.Receipt, error) {
	ca, err := tezos.ParseAddress(contr)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	con := contract.NewContract(ca, w.rpcClient)

	var aps_ []authTransferParam
	for _, ap := range aps {
		ap_, err := ap.Build()
		if err != nil {
			return nil, err
		}
		aps_ = append(aps_, *ap_)
	}

	args := authTransferArgs{
		Transfers: aps_,
	}

	args.Params = micheline.Parameters{
		Entrypoint: "authorized_transfer",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}
