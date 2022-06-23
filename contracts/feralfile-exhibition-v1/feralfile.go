package feralfilev1

import (
	"encoding/json"
	"errors"
	"fmt"

	"blockwatch.cc/tzgo/contract"
	tz "blockwatch.cc/tzgo/tezos"

	tezos "github.com/bitmark-inc/account-vault-tezos"
)

var (
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
func (c *FeralfileExhibitionV1Contract) Call(wallet *tezos.Wallet, method string, arguments json.RawMessage) (*string, error) {
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
	case "update_edition_metadata":
		var params []UpdateEditionMetadataParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return updateEditionMetadata(wallet, contract, params)
	case "burn_editions":
		var params []BurnEditionsParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return burnEditions(wallet, contract, params)
	default:
		return nil, fmt.Errorf("unsupported method")
	}
}

func init() {
	tezos.RegisterContract("FeralfileExhibitionV1", FeralfileExhibitionV1ContractFactory)
}
