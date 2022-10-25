package feralfilefeature

import (
	"encoding/hex"
	"math/big"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	tz "blockwatch.cc/tzgo/tezos"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

type MintEditionParam struct {
	Owner  string             `json:"owner"`
	Tokens []MintEditionToken `json:"tokens"`
}

func (m MintEditionParam) Build() (*mintEditionParam, error) {
	// address
	ow, err := tz.ParseAddress(m.Owner)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	var tks []mintEditionToken
	for _, tk := range m.Tokens {
		t, err := tk.Build()
		if err != nil {
			return nil, err
		}
		tks = append(tks, *t)
	}
	return &mintEditionParam{
		Owner:  ow,
		Tokens: tks,
	}, nil
}

type MintEditionToken struct {
	IPFSLink  string `json:"ipfs_link"`
	ArtworkID string `json:"artwork_id"`
	Edition   int64  `json:"edition"`
}

func (m MintEditionToken) Build() (*mintEditionToken, error) {
	a, err := hex.DecodeString(m.ArtworkID)
	if err != nil {
		return nil, err
	}

	return &mintEditionToken{
		ArtworkID: a,
		IPFSLink:  []byte(m.IPFSLink),
		edition:   big.NewInt(m.Edition),
	}, nil
}

type mintEditionParam struct {
	Owner  tz.Address
	Tokens []mintEditionToken
}

type mintEditionToken struct {
	IPFSLink  []byte
	ArtworkID []byte
	edition   *big.Int
}

type mintEditionArgs struct {
	contract.TxArgs
	Editions []mintEditionParam
}

var _ contract.CallArguments = (*mintEditionArgs)(nil)

func (p mintEditionParam) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for _, v := range p.Tokens {
		rs.Args = append(rs.Args,
			micheline.NewPair(
				micheline.NewSeq(
					NewElt(
						micheline.NewString(""),
						micheline.NewBytes(v.IPFSLink),
					),
				),
				micheline.NewPair(
					micheline.NewBytes(v.ArtworkID),
					micheline.NewNat(v.edition),
				),
			),
		)
	}
	return rs
}

func (p mintEditionArgs) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for i, v := range p.Editions {
		rs.Args = append(rs.Args,
			micheline.NewPair(
				micheline.NewBytes(v.Owner.Bytes22()),
				micheline.NewSeq(),
			),
		)
		rs.Args[i].Args[1] = v.Prim()
	}
	return rs
}

// mintEditions mint edition tokens for artworks
func MintEditions(w *tezos.Wallet, con *contract.Contract, mes []MintEditionParam) (*string, error) {
	var mes_ []mintEditionParam
	for _, me := range mes {
		me_, err := me.Build()
		if err != nil {
			return nil, err
		}
		mes_ = append(mes_, *me_)
	}

	args := mintEditionArgs{
		Editions: mes_,
	}

	args.Params = micheline.Parameters{
		Entrypoint: "mint_editions",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}
