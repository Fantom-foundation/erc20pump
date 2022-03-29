// Package cfg represents a structure of app config.
package cfg

import "github.com/ethereum/go-ethereum/common"

// Config represents the app configuration.
type Config struct {
	OperaURI     string
	StartBlock   uint64
	ScanContract common.Address

	AwsRegion   string
	AwsS3Bucket string
}
