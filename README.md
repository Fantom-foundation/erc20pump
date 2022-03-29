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

```shell
Usage of build/erc20pump:
  -contract string
    	Address of the contract being scanned for ERC20 transfers.
  -from uint
    	Numeric ID of the first loaded block.
  -opera string
    	Address of the chain node RPC interface.
```
