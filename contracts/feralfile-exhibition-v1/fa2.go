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

// transfer transfer FA2 tokens
func transfer(w *tezos.Wallet, con *contract.Contract, tps []TransferParam) (*string, error) {
	// construct transfer arguments
	args := contract.NewFA2TransferArgs()
	for _, tp := range tps {
		tp_, err := tp.Build()
		if err != nil {
			return nil, err
		}
		args.WithTransfer(
			w.PrivateKey().Address(),
			tp_.To,
			(tz.Z)(*tp_.TokenID),
			tz.NewZ(1),
		)
	}
	args.WithDestination(con.Address())
	args.Optimize()

	return w.Send(args)
}
