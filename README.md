# Ethereum Transaction Replay Tool

## Introduction
This project enhances Ethereum transaction processing efficiency through a restructured database that enables parallel execution of transactions. By optimizing how transactions are handled, the tool aims to meet the increasing demands on Ethereum's network, thus enabling faster and more efficient processing capabilities.

## Environment
- Golang version >= 1.19.0

## Project Structure
Core data directory:
+ `research`

## Usage

### Setup
Clone the project and execute the following command to build:
```bash
make all
```


This command compiles the source and generates two executable files in the build/bin directory: `geth` for creating the archival database and `evm` for transaction replay.

1. #### Data Recording
+ Obtain the compiled `geth` from build/bin/geth and copy it to a work directory, for example, `ReplayerSpace`.
In `ReplayerSpace`, create a JSON file to configure the archival database. Example configuration:
```json
{
"dir": "./",
"Names": [
"account",
"code",
"codeChange",
"storage",
"info"]
}
```
+ Use `geth` to import blockchain data in RLP format:
```bash
./geth --datadir ./geth_data import ./block/0To100w.rlp
```

2. #### Transaction Replay
+ Replay a specific transaction:
```bash
./evm replay blockNumber txIndex
```

+ Replay transactions within a specific block range:
```bash
./evm replay --range startBlockNumber endBlockNumber
```


## Note
This project has been tested up to the first 12 million blocks. Transactions beyond this range have not been experimented with.