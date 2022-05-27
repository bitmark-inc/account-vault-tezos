package tezos

import (
	"math/big"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

type TransferParam struct {
	To      string
	TokenID string
}

func (t TransferParam) Build() (*transferParam, error) {
	// address
	to_, err := tezos.ParseAddress(t.To)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	// token
	tk, ok := new(big.Int).SetString(t.TokenID, 10)
	if !ok {
		return nil, ErrInvalidTokenID
	}
	return &transferParam{
		To:      to_,
		TokenID: tk,
	}, nil
}

type transferParam struct {
	To      tezos.Address
	TokenID *big.Int
}

// Transfer transfer a FA2 token
func (w *Wallet) Transfer(contr string, tp TransferParam) (*rpc.Receipt, error) {
	ca, err := tezos.ParseAddress(contr)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	// construct a new contract
	con := contract.NewContract(ca, w.rpcClient)
	tp_, err := tp.Build()

	// construct an FA2 token
	token := con.AsFA2(0)
	token.TokenId.Set(tp_.TokenID)

	// construct simple transfer arguments
	args := token.Transfer(
		w.privateKey.Address(), // from
		tp_.To,                 // to
		tezos.NewZ(1),          // amount
	)
	args.WithDestination(con.Address())

	return w.Send(args)
}
