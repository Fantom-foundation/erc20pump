// Package erc20pump implements the application server entry.
package main

import (
	"erc20pump/internal/cfg"
	"flag"
	"github.com/ethereum/go-ethereum/common"
)

// config loads configuration from cli flags.
func config() *cfg.Config {
	con := cfg.Config{}
	var addr string

	flag.StringVar(&con.OperaURI, "opera", "https://rpcapi.fantom.network", "Address of the Fantom Opera RPC interface.")
	flag.Uint64Var(&con.StartBlock, "block", 0, "Numeric ID of the first loaded block.")
	flag.StringVar(&addr, "contract", "0x0", "Address of the contract being scanned for ERC20 transfers.")
	flag.StringVar(&con.AwsRegion, "awsregion", "eu-central-1", "The AWS region to upload the JSONs to")
	flag.StringVar(&con.AwsStream, "awsstream", "", "The Kinesis stream to upload the JSONs to (keep empty to generate local json files)")
	flag.Parse()

	// decode contract address
	con.ScanContract = common.HexToAddress(addr)
	return &con
}
