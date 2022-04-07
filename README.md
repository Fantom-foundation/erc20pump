# ERC20 Transaction Pump
The application connects to a chain node and collects ERC20 transactions 
from a given block number up, following the head when reached. The transactions
are sent to a consumer.

## Building
You need Golang v.1.17 or newer and the usual build environment. 
A preferred way of building the app is via provided Makefile.

```shell
make
```

## Running
The application provides usual parameters help via `-h` option.

The application requires to have AWS credentials configured in `~/.aws/credentials` to be able to upload into AWS.
Check [AWS doc for detail](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

```shell
Usage of build/erc20pump:
  -awsregion string
    	The AWS region to upload the JSONs to (default "eu-central-1")
  -awsstream string
    	The Kinesis stream to upload the JSONs to (keep empty to generate local json files)
  -block uint
    	Numeric ID of the first loaded block.
  -contract string
    	Address of the contract being scanned for ERC20 transfers. (default "0x0")
  -opera string
    	Address of the Fantom Opera RPC interface. (default "https://rpcapi.fantom.network")
```
