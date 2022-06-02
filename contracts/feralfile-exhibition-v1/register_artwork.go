package feralfilev1

import (
	"math/big"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"github.com/ethereum/go-ethereum/accounts/abi"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

type RegisterArtworkParam struct {
	ArtistName  string `json:"artist_name"`
	Fingerprint string `json:"fingerprint"`
	Title       string `json:"title"`
	MaxEdition  int64  `json:"max_edition"`
}

func (ra RegisterArtworkParam) Build() (*registerArtworkParam, error) {
	pfp, err := getPackedFingerprint(ra.Fingerprint)
	if err != nil {
		return nil, err
	}
	return &registerArtworkParam{
		ArtistName:  ra.ArtistName,
		Fingerprint: pfp,
		Title:       ra.Title,
		MaxEdition:  big.NewInt(ra.MaxEdition),
	}, nil
}

type registerArtworkParam struct {
	ArtistName  string
	Fingerprint []byte
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
						micheline.NewBytes(v.Fingerprint),
						micheline.NewBig(v.MaxEdition),
					),
				),
			),
		)
	}
	return rs
}

// registerArtworks register new artworks
func registerArtworks(w *tezos.Wallet, con *contract.Contract, ras []RegisterArtworkParam) (*string, error) {
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

// getPackedFingerprint returns the packed fingerprint. The value
// would be identical to the one generated from the ethereum solidity abi.encode.
// In this way we could keep the packed artwork fingerprint same as ethereum on tezos
func getPackedFingerprint(fingerprint string) ([]byte, error) {
	stringTy, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{
			Type: stringTy,
		},
	}

	bytes, err := args.Pack(fingerprint)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
