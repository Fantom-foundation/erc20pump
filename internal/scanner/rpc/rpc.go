// Package rpc implements Opera node communication wrappers through an adapter.
package rpc

import (
	"context"
	"erc20pump/internal/cfg"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	client "github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

// Adapter represents a communication interface to the Opera node.
type Adapter struct {
	rpc *client.Client
	ftm *ethclient.Client
}

// New creates a new RPC adapter.
func New(cfg *cfg.Config) (*Adapter, error) {
	con, err := connect(cfg.OperaURI)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		rpc: con,
		ftm: ethclient.NewClient(con),
	}, nil
}

// connects opens RPC connection to the Opera node.
func connect(uri string) (*client.Client, error) {
	c, err := client.Dial(uri)
	if err != nil {
		fmt.Println("can not connect Opera", err.Error())
		return nil, err
	}

	fmt.Println("Opera connected", uri)
	return c, nil
}

// TopBlock provides the numeric ID of the current blockchain head block.
func (a *Adapter) TopBlock() (uint64, error) {
	return a.ftm.BlockNumber(context.Background())
}

// GetLogs provides a slice of log records for the given topics and blocks range.
func (a *Adapter) GetLogs(topics [][]common.Hash, from uint64, to uint64) ([]types.Log, error) {
	return a.ftm.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Topics:    topics,
	})
}

// Transaction provides the transaction details.
func (a *Adapter) Transaction(tx common.Hash) (*types.Transaction, error) {
	trx, _, err := a.ftm.TransactionByHash(context.Background(), tx)
	if err != nil {
		fmt.Println("failed to get transaction", err.Error(), tx.String())
		return nil, err
	}
	return trx, nil
}
