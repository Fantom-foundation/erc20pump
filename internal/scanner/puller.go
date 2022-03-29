// Package scanner performs the scanning task.
package scanner

import (
	"bytes"
	"erc20pump/internal/cfg"
	"erc20pump/internal/scanner/rpc"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"sync"
	"time"
)

// logBufferCapacity represents the capacity of collected log records.
const logBufferCapacity = 100

// defaultLogsWindowSize represents the maximal number of blocks we try to pull at once.
const defaultLogsWindowSize = 5

// logPuller represents log record pulling service
type logPuller struct {
	output        chan types.Log
	topBlock      uint64
	currentBlock  uint64
	sigStop       chan bool
	wg            *sync.WaitGroup
	rpc           *rpc.Adapter
	topics        [][]common.Hash
	txRecipients  map[common.Hash]common.Address
	contractMatch func(rc *common.Address) bool
}

// newPuller creates a new puller service.
func newPuller(cfg *cfg.Config, rpc *rpc.Adapter) *logPuller {
	// build a list of topics we want to scan for
	topics := [][]common.Hash{make([]common.Hash, 0, len(LogTopicProcessor))}
	for t := range LogTopicProcessor {
		topics[0] = append(topics[0], t)
	}

	// make the puller
	return &logPuller{
		output:       make(chan types.Log, logBufferCapacity),
		topBlock:     0,
		currentBlock: cfg.StartBlock,
		sigStop:      make(chan bool, 1),
		topics:       topics,
		rpc:          rpc,
		contractMatch: func(rc *common.Address) bool {
			return bytes.Compare(rc.Bytes(), cfg.ScanContract.Bytes()) == 0
		},
	}
}

// run the log puller service.
func (lp *logPuller) run(wg *sync.WaitGroup) {
	lp.wg = wg

	wg.Add(1)
	go lp.scan()
}

// stop signals the log puller thread to terminate.
func (lp *logPuller) stop() {
	lp.sigStop <- true
}

// scan the blockchain for log records of interest.
func (lp *logPuller) scan() {
	tick := time.NewTicker(500 * time.Millisecond)
	info := time.NewTicker(5 * time.Second)

	defer func() {
		tick.Stop()
		info.Stop()
		close(lp.output)

		fmt.Println("log puller terminated")
		lp.wg.Done()
	}()

	var logs []types.Log
	var log types.Log
	for {
		// terminate if requested
		select {
		case <-lp.sigStop:
			return
		case <-tick.C:
			lp.fetchHead()
		case <-info.C:
			fmt.Println("scanner at #", lp.currentBlock, "head at #", lp.topBlock)
		default:
		}

		// do we have a log record to process?
		if logs == nil || len(logs) == 0 {
			logs = lp.nextLogs()
			continue
		}

		// get the next log and process
		log, logs = logs[0], logs[1:]
		lp.process(log)
	}
}

// fetchHead updates the current known head block index.
func (lp *logPuller) fetchHead() {
	var err error

	lp.topBlock, err = lp.rpc.TopBlock()
	if err != nil {
		fmt.Println("error pulling the current head", err.Error())
	}
}

// nextLogs pulls the next set of log records from the backend server.
func (lp *logPuller) nextLogs() []types.Log {
	// do we even have anything to pull?
	if lp.currentBlock >= lp.topBlock {
		return nil
	}

	// what is our current target?
	target := lp.currentBlock + defaultLogsWindowSize
	if target > lp.topBlock {
		target = lp.topBlock
	}

	// pull the data from remote server
	logs, err := lp.rpc.GetLogs(lp.topics, lp.currentBlock, target)
	if err != nil {
		fmt.Println("failed to pull logs", err.Error())
		return nil
	}

	// clear tx recipients map, if it makes sense
	if len(logs) > 0 {
		lp.txRecipients = make(map[common.Hash]common.Address)
	}

	// advance current block
	lp.currentBlock = target + 1
	return logs
}

// process given event log record.
func (lp *logPuller) process(ev types.Log) {
	var err error

	// do we know the transaction recipient?
	rec, ok := lp.txRecipients[ev.TxHash]
	if !ok {
		rec, err = lp.rpc.TrxRecipient(ev.TxHash)
		if err != nil {
			fmt.Println("can not get tx recipient:", err.Error())
			return
		}

		// remember the transaction recipient in case we have more logs from this tx
		lp.txRecipients[ev.TxHash] = rec
	}

	// is the recipient interesting?
	if !lp.contractMatch(&rec) {
		return
	}

	fmt.Println("match", rec.String(), "on", ev.TxHash.String())

	// this one is what we're looking for
	lp.output <- ev
}
