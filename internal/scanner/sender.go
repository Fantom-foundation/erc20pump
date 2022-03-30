// Package scanner performs the scanning task.
package scanner

import (
	"bytes"
	"encoding/json"
	"erc20pump/internal/cfg"
	"erc20pump/internal/trx"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// sender represents a sub-service responsible for sending collected transactions
type sender struct {
	input    chan trx.BlockchainTransaction
	uploader *s3manager.Uploader
	queue    []trx.BlockchainTransaction
	lastSent time.Time
	bucket   string
	sigStop  chan bool
	wg       *sync.WaitGroup
}

// newSender creates a new transaction sender instance.
func newSender(config *cfg.Config, in chan trx.BlockchainTransaction) *sender {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.AwsRegion),
	}))

	return &sender{
		input:    in,
		uploader: s3manager.NewUploader(sess),
		lastSent: time.Now(),
		bucket:   config.AwsS3Bucket,
		sigStop:  make(chan bool, 1),
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
	if se.bucket == "" {
		se.save(tx)
		return
	}

	// add to queue
	se.queue = append(se.queue, tx)

	if len(se.queue) >= 40 || time.Now().Sub(se.lastSent) > 2*time.Minute {
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

	// put the data into a file
	result, err := se.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(se.bucket),
		Key:    aws.String(se.queue[0].TXHash.String() + ".json"),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		log.Fatalf("Failed to upload into S3; %s", err)
		return
	}
	log.Printf("Uploaded %d transactions into S3 as %s", len(se.queue), result.Location)

	// empty the queue
	se.queue = make([]trx.BlockchainTransaction, 0, 100)
}
