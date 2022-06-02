package feralfilev1

import (
	"math/big"

	"blockwatch.cc/tzgo/contract"
	tz "blockwatch.cc/tzgo/tezos"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

const (
	DefaultAccountIndex = 0
	MAINNETChainID      = "NetXdQprcVkpaWU"
	ITHACANETChainID    = "NetXnHfVqm9iesp"
)

type TransferParam struct {
	To      string `json:"to"`
	TokenID string `json:"token_id"`
}

func (t TransferParam) Build() (*transferParam, error) {
	// address
	to_, err := tz.ParseAddress(t.To)
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
	To      tz.Address
	TokenID *big.Int
}

// transfer transfer a FA2 token
func transfer(w *tezos.Wallet, con *contract.Contract, tp TransferParam) (*string, error) {
	tp_, err := tp.Build()
	if err != nil {
		return nil, err
	}

	// construct a FA2 token
	token := con.AsFA2(0)
	token.TokenId.Set(tp_.TokenID)

	// construct simple transfer arguments
	args := token.Transfer(
		w.PrivateKey().Address(), // from
		tp_.To,                   // to
		tz.NewZ(1),               // amount
	)
	args.WithDestination(con.Address())

	return w.Send(args)
}
