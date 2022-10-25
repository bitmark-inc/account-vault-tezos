package feralfilefeature

import (
	"math/big"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

type BurnEditionsParam string

func (b BurnEditionsParam) Build() (*burnEditionsParam, error) {
	tk, ok := new(big.Int).SetString(string(b), 10)
	if !ok {
		return nil, ErrInvalidTokenID
	}
	btk := burnEditionsParam(*tk)
	return &btk, nil
}

type burnEditionsParam big.Int

type burnEditionsArgs struct {
	contract.TxArgs
	burnEditions []burnEditionsParam
}

var _ contract.CallArguments = (*updateEditionMetadataArgs)(nil)

func (p burnEditionsParam) Prim() micheline.Prim {
	b := big.Int(p)
	return micheline.NewBig(
		&b,
	)
}

func (p burnEditionsArgs) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for _, v := range p.burnEditions {
		rs.Args = append(rs.Args,
			v.Prim(),
		)
	}
	return rs
}

// burnEditions burn the editions
func BurnEditions(w *tezos.Wallet, con *contract.Contract, bes []BurnEditionsParam) (*string, error) {
	var _bes []burnEditionsParam
	for _, be := range bes {
		_be, err := be.Build()
		if err != nil {
			return nil, err
		}
		_bes = append(_bes, *_be)
	}

	args := burnEditionsArgs{
		burnEditions: _bes,
	}

	args.Params = micheline.Parameters{
		Entrypoint: "burn_editions",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}
