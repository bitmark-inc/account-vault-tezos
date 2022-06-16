package feralfilev1

import (
	"math/big"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

type UpdateEditionMetadataParam struct {
	TokenID  string `json:"token_id"`
	IPFSLink string `json:"ipfs_link"`
}

func (u UpdateEditionMetadataParam) Build() (*updateEditionMetadataParam, error) {
	tk, ok := new(big.Int).SetString(u.TokenID, 10)
	if !ok {
		return nil, ErrInvalidTokenID
	}
	return &updateEditionMetadataParam{
		TokenID:  tk,
		IPFSLink: []byte(u.IPFSLink),
	}, nil
}

type updateEditionMetadataParam struct {
	TokenID  *big.Int
	IPFSLink []byte
}

type updateEditionMetadataArgs struct {
	contract.TxArgs
	updateEditions []updateEditionMetadataParam
}

var _ contract.CallArguments = (*updateEditionMetadataArgs)(nil)

func (p updateEditionMetadataParam) Prim() micheline.Prim {
	return micheline.NewPair(
		micheline.NewBig(p.TokenID),
		micheline.NewSeq(
			newElt(
				micheline.NewString(""),
				micheline.NewBytes(p.IPFSLink),
			),
		),
	)
}

func (p updateEditionMetadataArgs) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for _, v := range p.updateEditions {
		rs.Args = append(rs.Args,
			v.Prim(),
		)
	}
	return rs
}

// updateEditionMetadata update the edition token metadata
func updateEditionMetadata(w *tezos.Wallet, con *contract.Contract, uem []UpdateEditionMetadataParam) (*string, error) {
	var _uem []updateEditionMetadataParam
	for _, ue := range uem {
		ue_, err := ue.Build()
		if err != nil {
			return nil, err
		}
		_uem = append(_uem, *ue_)
	}

	args := updateEditionMetadataArgs{
		updateEditions: _uem,
	}

	args.Params = micheline.Parameters{
		Entrypoint: "update_edition_metadata",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}
