package tezos

import (
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/rpc"
)

// Contract is an interface defines how a vault interact with the smart contract
type Contract interface {
	Deploy(wallet *Wallet, arguments json.RawMessage) (address string, txID string, err error)
	Call(wallet *Wallet, method string, arguments json.RawMessage) (tx *rpc.Receipt, err error)
}

// ContractFactory is a function that takes an address and return a Contract instance
type ContractFactory func(string) Contract

// contracts is a singleton of all registered contract factory function
var contracts = map[string]ContractFactory{}

// GetContract returns
func GetContract(name string) ContractFactory {
	return contracts[name]
}

// RegisterContract registers a contract
func RegisterContract(name string, contract ContractFactory) error {
	if _, ok := contracts[name]; ok {
		return fmt.Errorf("duplicated contract name has registered")
	}

	contracts[name] = contract
	return nil
}
