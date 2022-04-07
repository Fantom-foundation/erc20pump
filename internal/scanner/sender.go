// Package scanner performs the scanning task.
package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"erc20pump/internal/cfg"
	"erc20pump/internal/trx"
	"fmt"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// sender represents a sub-service responsible for sending collected transactions
type sender struct {
	input    chan trx.BlockchainTransaction
	uploader *kinesis.Kinesis
	queue    []trx.BlockchainTransaction
	lastSent   time.Time
	streamName string
	sigStop    chan bool
	wg       *sync.WaitGroup
}

// newSender creates a new transaction sender instance.
func newSender(config *cfg.Config, in chan trx.BlockchainTransaction) *sender {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.AwsRegion),
	}))

	return &sender{
		input:      in,
		uploader:   kinesis.New(sess),
		lastSent:   time.Now(),
		streamName: config.AwsStream,
		sigStop:    make(chan bool, 1),
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
		log.Println("sender terminated")
		se.wg.Done()
	}()

	for {
		select {
		case <-se.sigStop:
			return
		case tx := <-se.input:
			se.process(tx)
		}
	}
}

// process adds the transaction into queue, sends if the queue is log/old enough
func (se *sender) process(tx trx.BlockchainTransaction) {
	// store locally instead if no bucket is specified
	if se.streamName == "" {
		se.save(tx)
		return
	}

	// add to queue
	se.queue = append(se.queue, tx)

	if len(se.queue) >= 10 || time.Now().Sub(se.lastSent) > 2 * time.Minute {
		se.send()
		se.lastSent = time.Now()
	}
}

// save stores the transaction data locally to a file.
func (se *sender) save(tx trx.BlockchainTransaction) {
	log.Println("storing", tx.TXHash.String())

	// encode the transaction into a human-readable JSON struct
	data, err := json.MarshalIndent(tx, "", "    ")
	if err != nil {
		log.Println("can not encode to JSON", err.Error())
		return
	}

	// put the data into a file
	err = ioutil.WriteFile(tx.TXHash.String()+".json", data, 0644)
	if err != nil {
		log.Println("can not write JSON to file", err.Error())
	}
}

// send the data to S3
func (se *sender) send() {
	log.Printf("sending %d transactions", len(se.queue))

	// encode the transaction into a human-readable JSON struct
	data, err := json.MarshalIndent(se.queue, "", "    ")
	if err != nil {
		fmt.Println("can not encode transactions into JSON", err.Error())
		return
	}

	fmt.Printf("storing data \"%s\"\n", string(data))

	hash := md5.Sum(data)
	dataHash := hex.EncodeToString(hash[:])

	// put the data into the Kinesis data stream
	_, err = se.uploader.PutRecord(&kinesis.PutRecordInput{
		StreamName: &se.streamName,
		Data:       data,
		PartitionKey: &dataHash,
	})
	if err != nil {
		log.Fatalf("Failed to upload into Kinesis; %s", err)
		return
	}
	log.Printf("Uploaded %d transactions into Kinesis", len(se.queue))

	// empty the queue
	se.queue = make([]trx.BlockchainTransaction, 0, 100)
}
