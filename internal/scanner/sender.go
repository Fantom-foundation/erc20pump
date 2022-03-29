// Package scanner performs the scanning task.
package scanner

import (
	"encoding/json"
	"erc20pump/internal/cfg"
	"erc20pump/internal/trx"
	"fmt"
	"io/ioutil"
	"sync"
)

// sender represents a sub-service responsible for sending collected transactions
type sender struct {
	input   chan trx.BlockchainTransaction
	sigStop chan bool
	wg      *sync.WaitGroup
}

// newSender creates a new transaction sender instance.
func newSender(_ *cfg.Config, in chan trx.BlockchainTransaction) *sender {
	return &sender{
		input:   in,
		sigStop: make(chan bool, 1),
	}
}

// run the sender service.
func (se *sender) run(wg *sync.WaitGroup) {
	se.wg = wg

	wg.Add(1)
	go se.observe()
}

// stop signals the sender thread to terminate.
func (se *sender) stop() {
	se.sigStop <- true
}

// scan the blockchain for log records of interest.
func (se *sender) observe() {
	defer func() {
		fmt.Println("sender terminated")
		se.wg.Done()
	}()

	for {
		select {
		case <-se.sigStop:
			return
		case tx := <-se.input:
			se.send(tx)
		}
	}
}

// send the given transaction to the consumer.
func (se *sender) send(tx trx.BlockchainTransaction) {
	fmt.Println("storing", tx.TXHash.String())

	// encode the transaction into a human-readable JSON struct
	data, err := json.MarshalIndent(tx, "", "    ")
	if err != nil {
		fmt.Println("can not encode to JSON", err.Error())
		return
	}

	// put the data into a file
	// @todo Replace this with S3 buckets support.
	err = ioutil.WriteFile(tx.TXHash.String()+".json", data, 0644)
	if err != nil {
		fmt.Println("can not write JSON to file", err.Error())
	}
}
