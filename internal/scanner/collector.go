// Package scanner performs the scanning task.
package scanner

import (
	"bytes"
	"erc20pump/internal/cfg"
	"erc20pump/internal/scanner/cache"
	"erc20pump/internal/scanner/rpc"
	"erc20pump/internal/trx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"sync"
	"time"
)

// logCollector represents a service responsible for collecting patches of transfers
type logCollector struct {
	input      chan types.Log
	output     chan trx.BlockchainTransaction
	sigStop    chan bool
	currentTrx *trx.BlockchainTransaction
	tokens     map[common.Address]trx.Token
	rpc        *rpc.Adapter
	cache      *cache.MemCache
	wg         *sync.WaitGroup
}

// LogTopicProcessor represents a map of base log topic to transaction type.
var LogTopicProcessor = map[common.Hash]func(*types.Log, func(common.Address) trx.Token) trx.Erc20Transaction{
	common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"): decodeErc20Transfer,
	/* common.HexToHash("0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"): "APPROVAL", */
}

// newCollector creates a new log collector instance.
func newCollector(_ *cfg.Config, in chan types.Log, rpc *rpc.Adapter, cache *cache.MemCache) *logCollector {
	return &logCollector{
		input:   in,
		output:  make(chan trx.BlockchainTransaction, 25),
		tokens:  make(map[common.Address]trx.Token),
		sigStop: make(chan bool, 1),
		rpc:     rpc,
		cache:   cache,
	}
}

// run the log collector service.
func (lc *logCollector) run(wg *sync.WaitGroup) {
	lc.wg = wg

	wg.Add(1)
	go lc.collect()
}

// stop signals the log collector thread to terminate.
func (lc *logCollector) stop() {
	lc.sigStop <- true
}

// collect interesting transactions and build collections for sending.
func (lc *logCollector) collect() {
	// auto-close pending transaction if no new event arrived in given time
	tick := time.NewTicker(5 * time.Second)

	defer func() {
		tick.Stop()
		close(lc.output)

		log.Println("log collector terminated")
		lc.wg.Done()
	}()

	for {
		select {
		case <-lc.sigStop:
			return

		case <-tick.C:
			lc.newTransaction(nil)

		case ev := <-lc.input:
			tick.Reset(5 * time.Second)
			lc.process(ev)
		}
	}
}

// process log event into the collectors' transaction.
func (lc *logCollector) process(ev types.Log) {
	// is this the same chain trx?
	if lc.currentTrx == nil || bytes.Compare(lc.currentTrx.TXHash.Bytes(), ev.TxHash.Bytes()) != 0 {
		lc.newTransaction(&ev)
	}

	// do we have a decoder for this type of event?
	decode, ok := LogTopicProcessor[ev.Topics[0]]
	if !ok {
		log.Println("decoder lookup failed")
		return
	}

	// add decoded tx to the current transaction group
	lc.currentTrx.Transactions = append(lc.currentTrx.Transactions, decode(&ev, lc.token))
}

// newTransaction closes the current transaction, if any, and makes a new one.
func (lc *logCollector) newTransaction(ev *types.Log) {
	// submit the current transaction
	if lc.currentTrx != nil {
		log.Println("closing group", lc.currentTrx.TXHash.String())
		lc.output <- *lc.currentTrx
	}

	// no new log, just closing
	if ev == nil {
		lc.currentTrx = nil
		return
	}

	// make a new transaction record
	lc.currentTrx = &trx.BlockchainTransaction{
		TXHash:       ev.TxHash,
		From:         lc.sender(ev.TxHash),
		To:           lc.recipient(ev.TxHash),
		BlockNumber:  hexutil.Uint64(ev.BlockNumber),
		Timestamp:    lc.timestamp(ev.BlockNumber),
		Transactions: make([]trx.Erc20Transaction, 0),
	}

	log.Println("new group", ev.TxHash.String())
}

// decodeErc20Transfer decodes ERC20 transfer event log record into ERC20 trx structure.
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func decodeErc20Transfer(ev *types.Log, token func(common.Address) trx.Token) trx.Erc20Transaction {
	return trx.Erc20Transaction{
		Token:     token(ev.Address),
		Type:      "TRANSFER",
		Sender:    common.BytesToAddress(ev.Topics[1].Bytes()),
		Recipient: common.BytesToAddress(ev.Topics[2].Bytes()),
		Amount:    hexutil.Big(*new(big.Int).SetBytes(ev.Data[:32])),
	}
}

// timestamp provides time of the block by block number.
func (lc *logCollector) timestamp(blk uint64) hexutil.Uint64 {
	ts, err := lc.cache.BlockTime(blk, lc.rpc.BlockTime)
	if err != nil {
		log.Fatalf("block timestamp not available for %d; %s", blk, err.Error())
		return 0
	}
	return hexutil.Uint64(ts)
}

// recipient provides recipient of a transaction by its hash.
func (lc *logCollector) recipient(tx common.Hash) common.Address {
	a, err := lc.cache.TrxRecipient(tx, lc.rpc.TrxRecipient)
	if err != nil {
		log.Fatalf("no recipient available for %s; %s", tx.String(), err.Error())
		return common.Address{}
	}
	return a
}

// sender provides signing address of a transaction by its hash.
func (lc *logCollector) sender(tx common.Hash) common.Address {
	a, err := lc.rpc.TrxSender(tx)
	if err != nil {
		log.Fatalf("no sender available for %s; %s", tx.String(), err.Error())
		return common.Address{}
	}
	return a
}

// token provides an ERC20 detail structure based on token contract address.
func (lc *logCollector) token(adr common.Address) trx.Token {
	// do we already know the token?
	tok, ok := lc.tokens[adr]
	if ok {
		return tok
	}

	// we need to pull the data from RPC
	name, err := lc.rpc.Erc20Name(adr)
	if err != nil {
		log.Println("token name lookup failed", err.Error(), adr.Hex())
		name = "unknown"
	}

	symbol, err := lc.rpc.Erc20Symbol(adr)
	if err != nil {
		log.Println("token symbol lookup failed", err.Error(), adr.Hex())
		name = "-"
	}

	decimals, err := lc.rpc.Erc20Decimals(adr)
	if err != nil {
		log.Println("token decimals lookup failed", err.Error(), adr.Hex())
		decimals = 0
	}

	tok = trx.Token{
		Address:  adr,
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
	}

	log.Println("new token found", tok.Name, "/", tok.Symbol, "[", tok.Decimals, "]")
	lc.tokens[adr] = tok

	return tok
}
