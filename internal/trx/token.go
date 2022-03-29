// Package trx implements transaction types.
package trx

import "github.com/ethereum/go-ethereum/common"

// Token represents a description of an ERC20 token.
type Token struct {
	Address  common.Address `json:"address"`
	Name     string         `json:"name"`
	Symbol   string         `json:"symbol"`
	Decimals uint8          `json:"decimals"`
}
