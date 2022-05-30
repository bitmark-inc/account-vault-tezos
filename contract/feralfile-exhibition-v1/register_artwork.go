package feralfilev1

import (
	"math/big"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

type RegisterArtworkParam struct {
	ArtistName  string
	Fingerprint string
	Title       string
	MaxEdition  int64
}

func (ra RegisterArtworkParam) Build() (*registerArtworkParam, error) {
	return &registerArtworkParam{
		ArtistName:  ra.ArtistName,
		Fingerprint: ra.Fingerprint,
		Title:       ra.Title,
		MaxEdition:  big.NewInt(ra.MaxEdition),
	}, nil
}

type registerArtworkParam struct {
	ArtistName  string
	Fingerprint string
	Title       string
	MaxEdition  *big.Int
}

type registerArtworkArgs struct {
	contract.TxArgs
	Artworks []registerArtworkParam
}

var _ contract.CallArguments = (*mintEditionArgs)(nil)

func (p registerArtworkArgs) Prim() micheline.Prim {
	rs := micheline.NewSeq()
	for _, v := range p.Artworks {
		rs.Args = append(rs.Args,
			micheline.NewPair(
				micheline.NewString(v.Title),
				micheline.NewPair(
					micheline.NewString(v.ArtistName),
					micheline.NewPair(
						micheline.NewString(v.Fingerprint),
						micheline.NewBig(v.MaxEdition),
					),
				),
			),
		)
	}
	return rs
}

// registerArtworks register new artworks
func registerArtworks(w *tezos.Wallet, con *contract.Contract, ras []RegisterArtworkParam) (*rpc.Receipt, error) {
	var ras_ []registerArtworkParam
	for _, ra := range ras {
		ra_, err := ra.Build()
		if err != nil {
			return nil, err
		}
		ras_ = append(ras_, *ra_)
	}

	args := registerArtworkArgs{
		Artworks: ras_,
	}

	args.Params = micheline.Parameters{
		Entrypoint: "register_artworks",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}
