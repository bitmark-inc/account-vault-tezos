package feralfilev2

import (
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/contract"
	tz "blockwatch.cc/tzgo/tezos"

	tezos "github.com/bitmark-inc/account-vault-tezos"
	fff "github.com/bitmark-inc/account-vault-tezos/contracts/feralfile-feature"
)

type FeralfileExhibitionV2Contract struct {
	contractAddress string
}

func FeralfileExhibitionV1ContractFactory(contractAddress string) tezos.Contract {
	return &FeralfileExhibitionV2Contract{
		contractAddress: contractAddress,
	}
}

// FIXME: TODO
// Deploy deploys the smart contract to tezos blockchain
func (c *FeralfileExhibitionV2Contract) Deploy(wallet *tezos.Wallet, arguments json.RawMessage) (string, string, error) {
	return "", "", nil
}

// Call is the entry function for account vault to interact with a smart contract.
func (c *FeralfileExhibitionV2Contract) Call(wallet *tezos.Wallet, method string, arguments json.RawMessage) (*string, error) {
	ca, err := tz.ParseAddress(c.contractAddress)
	if err != nil {
		return nil, fff.ErrInvalidAddress
	}
	// construct a new contract
	contract := contract.NewContract(ca, wallet.RPCClient())

	switch method {
	case "transfer":
		var params []fff.TransferParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return fff.Transfer(wallet, contract, params)
	case "authorized_transfer":
		var params []fff.AuthTransferParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return fff.AuthTransfer(wallet, contract, params)
	case "register_artworks":
		var params []fff.RegisterArtworkParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return fff.RegisterArtworks(wallet, contract, params)
	case "mint_editions":
		var params []fff.MintEditionParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return fff.MintEditions(wallet, contract, params)
	case "update_edition_metadata":
		var params []fff.UpdateEditionMetadataParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return fff.UpdateEditionMetadata(wallet, contract, params)
	case "burn_editions":
		var params []fff.BurnEditionsParam
		if err := json.Unmarshal(arguments, &params); err != nil {
			return nil, err
		}
		return fff.BurnEditions(wallet, contract, params)
	default:
		return nil, fmt.Errorf("unsupported method")
	}
}

func init() {
	tezos.RegisterContract("FeralfileExhibitionV2", FeralfileExhibitionV1ContractFactory)
}
