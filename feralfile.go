package tezos

import (
	"math/big"
	"time"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
)

type AuthTransferParam struct {
	From      string
	PK        string
	Timestamp time.Time
	Txs       []AuthTransaction
}

type AuthTransaction struct {
	To        string
	Signature string
	Amount    string
	TokenID   string
}

type authTransferParam struct {
	From      tezos.Address
	PK        tezos.Key
	Timestamp big.Int
	Txs       []authTransaction
}

type authTransaction struct {
	To        tezos.Address
	Signature tezos.Signature
	Amount    big.Int
	TokenID   big.Int
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
					micheline.NewBig(&v.TokenID),
					micheline.NewPair(
						micheline.NewNat(&v.Amount),
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
						micheline.NewBig(&v.Timestamp),
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
func (w *Wallet) SignAuthTransferMessage(ts time.Time, to, tokenID string) (string, error) {
	m, err := buildAuthTransferMessage(ts, to, tokenID)
	if err != nil {
		return "", err
	}
	return w.SignMessage(m)
}

// AuthTransfer call the authorized transfer entrypoint define in FeralFile contract
func (w *Wallet) AuthTransfer(contr string, ap []AuthTransferParam) (*rpc.Receipt, error) {
	w.rpcClient.Signer = signer.NewFromKey(w.privateKey)

	ca, err := tezos.ParseAddress(contr)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	con := contract.NewContract(ca, w.rpcClient)

	ap_, err := buildAuthTransferParam(ap)
	if err != nil {
		return nil, err
	}

	args := authTransferArgs{
		Transfers: ap_,
	}

	args.Params = micheline.Parameters{
		Entrypoint: "authorized_transfer",
		Value:      args.Prim(),
	}
	args.WithDestination(con.Address())

	return w.Send(&args)
}

// buildAuthTransferMessage build the authorized transfer message
func buildAuthTransferMessage(timestamp time.Time, to, tokenID string) ([]byte, error) {
	// timestamp
	ts := big.NewInt(timestamp.Unix())

	// address
	ad, err := tezos.ParseAddress(to)
	if err != nil {
		return nil, ErrInvalidAddress
	}

	// token
	n := new(big.Int)
	tk, ok := n.SetString(tokenID, 10)
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

func buildAuthTransferParam(a []AuthTransferParam) ([]authTransferParam, error) {
	rs := []authTransferParam{}
	for _, aa := range a {
		from_, err := tezos.ParseAddress(aa.From)
		if err != nil {
			return nil, ErrInvalidAddress
		}
		pk_, err := tezos.ParseKey(aa.PK)
		if err != nil {
			return nil, ErrInvalidPublicKey
		}
		txs, err := buildAuthTransactions(aa.Txs)
		if err != nil {
			return nil, err
		}
		rs = append(rs, authTransferParam{
			From:      from_,
			PK:        pk_,
			Timestamp: *big.NewInt(aa.Timestamp.Unix()),
			Txs:       txs,
		})
	}

	return rs, nil
}

func buildAuthTransactions(a []AuthTransaction) ([]authTransaction, error) {
	rs := []authTransaction{}
	for _, aa := range a {
		sig_, err := tezos.ParseSignature(aa.Signature)
		if err != nil {
			return nil, ErrInvalidSignature
		}
		// big int token
		n := new(big.Int)
		tk, _ := n.SetString(aa.TokenID, 10)
		to_, err := tezos.ParseAddress(aa.To)
		if err != nil {
			return nil, ErrInvalidAddress
		}
		rs = append(rs, authTransaction{
			Signature: sig_,
			TokenID:   *tk,
			To:        to_,
			Amount:    *big.NewInt(1),
		})
	}

	return rs, nil
}
