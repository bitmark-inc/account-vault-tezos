package feralfilev1

import (
	"math/big"
	"time"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	tz "blockwatch.cc/tzgo/tezos"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

type AuthTransferParam struct {
	From      string            `json:"from"`
	PK        string            `json:"pk"`
	Timestamp time.Time         `json:"timestamp"`
	Txs       []AuthTransaction `json:"txs"`
}

func (a AuthTransferParam) Build() (*authTransferParam, error) {
	from_, err := tz.ParseAddress(a.From)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	pk_, err := tz.ParseKey(a.PK)
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
	To        string `json:"to"`
	Signature string `json:"signature"`
	TokenID   string `json:"token_id"`
}

func (a AuthTransaction) Build() (*authTransaction, error) {
	sig_, err := tz.ParseSignature(a.Signature)
	if err != nil {
		return nil, ErrInvalidSignature
	}
	tk, ok := new(big.Int).SetString(a.TokenID, 10)
	if !ok {
		return nil, ErrInvalidTokenID
	}
	to_, err := tz.ParseAddress(a.To)
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
	From      tz.Address
	PK        tz.Key
	Timestamp *big.Int
	Txs       []authTransaction
}

type authTransaction struct {
	To        tz.Address
	Signature tz.Signature
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

// authTransfers call the authorized transfer entrypoint define in FeralFile contract
func authTransfers(w *tezos.Wallet, con *contract.Contract, aps []AuthTransferParam) (*rpc.Receipt, error) {
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
		Entrypoint: "authorized_transfers",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}
