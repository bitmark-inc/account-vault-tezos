package feralfilev1

import (
	"encoding/json"
	"errors"
	"fmt"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/rpc"
	tz "blockwatch.cc/tzgo/tezos"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

var (
	ErrWrongChainID     = errors.New("Connected node serve different chain from setting")
	ErrInvalidRpcNode   = errors.New("Invalid rpc node")
	ErrSignFailed       = errors.New("Failed to sign with provided data")
	ErrInvalidAddress   = errors.New("Invalid address provided")
	ErrInvalidPublicKey = errors.New("Invalid public key provided")
	ErrInvalidSignature = errors.New("Invalid signature provided")
	ErrInvalidTokenID   = errors.New("Invalid tokenID provided")
)

type FeralfileExhibitionV1Contract struct {
	contractAddress string
}

func FeralfileExhibitionV1ContractFactory(contractAddress string) tezos.Contract {
	return &FeralfileExhibitionV1Contract{
		contractAddress: contractAddress,
	}
}

// FIXME: TODO
// Deploy deploys the smart contract to tezos blockchain
func (c *FeralfileExhibitionV1Contract) Deploy(wallet *tezos.Wallet, arguments json.RawMessage) (string, string, error) {
	return "", "", nil
}

// Call is the entry function for account vault to interact with a smart contract.
func (c *FeralfileExhibitionV1Contract) Call(wallet *tezos.Wallet, method string, arguments json.RawMessage) (*rpc.Receipt, error) {
	ca, err := tz.ParseAddress(c.contractAddress)
	if err != nil {
		return nil, ErrInvalidAddress
	}
	// construct a new contract
	contract := contract.NewContract(ca, wallet.RPCClient())

	switch method {
	case "transfer":
		var params TransferParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return transfer(wallet, contract, params)
	case "authorized_transfers":
		var params []AuthTransferParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return authTransfers(wallet, contract, params)
	case "register_artworks":
		var params []RegisterArtworkParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return registerArtworks(wallet, contract, params)
	case "mint_editions":
		var params []MintEditionParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return mintEditions(wallet, contract, params)
	default:
		return nil, fmt.Errorf("unsupported method")
	}
}

func init() {
	tezos.RegisterContract("FeralfileExhibitionV1", FeralfileExhibitionV1ContractFactory)
}
