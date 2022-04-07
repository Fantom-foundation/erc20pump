// Package trx implements transaction types.
package trx

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// BlockchainTransaction represents a blockchain transaction.
type BlockchainTransaction struct {
	TXHash       common.Hash        `json:"hash"`
	BlockNumber  hexutil.Uint64     `json:"blockNumber"`
	Timestamp    hexutil.Uint64     `json:"timestamp"`
	From         common.Address     `json:"from"`
	To           common.Address     `json:"to"`
	Transactions []Erc20Transaction `json:"erc20Transactions"`
}

// Erc20Transaction represents an ERC20 token transaction as part of the blockchain transaction.
type Erc20Transaction struct {
	Token     Token          `json:"token"`
	Type      string         `json:"trxType"`
	Sender    common.Address `json:"sender"`
	Recipient common.Address `json:"recipient"`
	Amount    hexutil.Big    `json:"amount"`
}
