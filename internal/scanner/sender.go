// Package scanner performs the scanning task.
package scanner

import (
	"bytes"
	"encoding/json"
	"erc20pump/internal/cfg"
	"erc20pump/internal/trx"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// sender represents a sub-service responsible for sending collected transactions
type sender struct {
	input    chan trx.BlockchainTransaction
	uploader *s3manager.Uploader
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
		input:   in,
		uploader: s3manager.NewUploader(sess),
		bucket:  config.AwsS3Bucket,
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

	fmt.Printf("storing data \"%s\"\n", string(data))

	// put the data into a file
	result, err := se.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(se.bucket),
		Key:    aws.String(tx.TXHash.String()+".json"),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		log.Fatalf("Failed to upload into S3; %s", err)
		return
	}
	log.Printf("Uploaded tx %s into S3 as %s", tx.TXHash.String(), result.Location)
}
