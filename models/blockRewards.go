package models

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/op/go-logging"
	"github.com/starslabhq/rewards-collection/errors"
	"math/big"
	"time"
)

var blockslogger = logging.MustGetLogger("blocks.scraper.models")

/*
eth.getBlockByNumber:transactions, iterate transactions[],
to eth.getTransaction:gasPrice. to eth.getTransactionReceipt:gasUsed,
multiple(gasPrice, gasUsed), sum all
*/

var (
	Client *ethclient.Client
	//wg sync.WaitGroup
	//RPCClient *rpc.Client
)

type BlockchainInfo struct {
	LastBlockNum *big.Int
	ThisBlockNum *big.Int
	Blocks       []BlockInfo
	Transactions []TransactionInfo
	EpochIndex	 uint64
	TotalFees 	*big.Int
}

type BlockInfo struct {
	Num              *big.Int
	Timestamp        time.Time
	Hash             string
	TransactionCount int
	BlockFee         *big.Int
}

type TransactionInfo struct {
	Hash            string
	To              string
	Value           *big.Int
	Data            string
	ContractAddress string
	Fee             *big.Int
}


// GetBlockchainInfo retrieve top-level information about the blockchain
func GetBlockchainInfo(archNode string) *BlockchainInfo {
	blockchainInfo := &BlockchainInfo{
	}

	// RPC call to retrieve the latest block
	lastBlockNum, err := getBlockNumber(archNode)
	blockslogger.Infof("The latest block numer is: %v", lastBlockNum)
	if err != nil{
		blockslogger.Errorf("Get latest block error: %v", err)
	}

	maxBlock := EP
	//get mod of the lastBlockNum % EP > 10 to prevent `reorg` issue
	del := new(big.Int)
	del.Mod(lastBlockNum, big.NewInt(int64(EP)))
	if del.Cmp(big.NewInt(int64(10))) == -1 {
		//wait for 40 seconds more to recall the blockNumber
		time.Sleep(30 * time.Second)
		lastBlockNum, err = getBlockNumber(archNode)
		if err != nil{
			blockslogger.Errorf("Get latest block error: %v", err)
		}
		del.Mod(lastBlockNum, big.NewInt(int64(EP)))
		if del.Cmp(big.NewInt(int64(10))) == -1 {
			blockslogger.Warningf("No height updated, need to check the chain stability!")
			return nil
		}
	}

	//lastBlockNum in epoch should be modified as below:
	lb := new(big.Int)
	blockchainInfo.LastBlockNum = lb.Sub(lastBlockNum, del.Add(del, big.NewInt(int64(1))))

	//thisBlockNum in epoch
	tb := new(big.Int)
	blockchainInfo.ThisBlockNum = tb.Sub(lastBlockNum, del.Add(del, big.NewInt(int64(199))))

	epindex := new(big.Int)
	blockchainInfo.EpochIndex = epindex.Div(tb, big.NewInt(int64(200))).Uint64()

	fees := big.NewInt(0)
	for i := uint64(0); i < maxBlock; i++ {
		blockNum := big.NewInt(0).Set(blockchainInfo.LastBlockNum).Sub(blockchainInfo.LastBlockNum, big.NewInt(int64(i)))

		//init the Client
		Client, err = ethclient.Dial(archNode)
		if err != nil {
			blockslogger.Warningf("Dial archNode error!")
			return nil
		}
		defer Client.Close()
		// retrieve the block, which includes all of the transactions
		block, err := Client.BlockByNumber(context.TODO(), blockNum)
		if err != nil {
			blockslogger.Warningf("Error getting block %v by number: %v", blockNum, err)
			continue
		}

		// store the block info in a struct
		hash := block.Hash().Hex()

		//get the blockFee
		blockFee, err := getBlockFeesByBatch(archNode, blockNum)
		if blockFee == nil {
			blockFee = big.NewInt(0)
		}
		if err != nil {
			blockslogger.Warningf("Error getting block  fee %v by number: %v", blockNum, err)
			continue
		}

		blockInfo := BlockInfo{
			Num:              big.NewInt(0).Set(blockNum),
			Timestamp:        time.Unix(int64(block.Time()), 0),
			Hash:             hash,
			TransactionCount: len(block.Transactions()),
			BlockFee: big.NewInt(0).Set(blockFee),
		}

		fees.Add(fees, blockFee)
		// append the block info to the blockchain info struct
		blockchainInfo.Blocks = append(blockchainInfo.Blocks, blockInfo)

	}
	blockchainInfo.TotalFees = fees

	return blockchainInfo
}

//GetBlockFeeForBlock Deprecated adds the transactions and blockFee for ThisBlockNum into the BlockchainInfo struct
//func GetBlockFeeForBlock(archNode string, blockNumber *big.Int) (blockFee *big.Int, err error){
//	if blockNumber == nil {
//		blockslogger.Warningf("No block number to retrieve transactions from")
//		return
//	}
//
//	//init the Client
//	Client, err := ethclient.Dial(archNode)
//	if err != nil {
//		blockslogger.Warningf("Dial archNode error!")
//		return nil, errors.BadRequestErrorf(errors.EthCallError, "Dial archNode error: %v", err)
//	}
//	defer Client.Close()
//
//	// retrieve the block, which includes all of the transactions
//	block, err := Client.BlockByNumber(context.TODO(), blockNumber)
//	if err != nil {
//		blockslogger.Warningf("Error getting block %v by number: %v", blockNumber, err)
//		return nil, errors.BadRequestErrorf(errors.EthCallError, "Error getting block %v by number: %v", blockNumber, err)
//	}
//
//	//scrape all the transaction fees and return the blockFee
//	blockFee = big.NewInt(0)
//	for _, transaction := range []*types.Transaction(block.Transactions()) {
//		// retrieve transaction receipt
//		receipt, err := Client.TransactionReceipt(context.TODO(), transaction.Hash())
//		if err != nil {
//			blockslogger.Warningf("Error getting transaction receipt: %v", err)
//			return nil, errors.BadRequestErrorf(errors.EthCallError, "Error getting transaction receipt: %v", err)
//		}
//
//		transactionInfo := TransactionInfo{
//			Fee:             big.NewInt(0).Mul(transaction.GasPrice(), big.NewInt(int64(receipt.GasUsed))),
//		}
//		blockFee.Add(blockFee, transactionInfo.Fee)
//	}
//	return blockFee, nil
//}

//getBlockFeesByBatch fetch the blockFees by batch
func getBlockFeesByBatch(archNode string, blockNumber *big.Int) (*big.Int, error){
	if blockNumber == nil {
		blockslogger.Warningf("No block number to retrieve transactions from")
		return nil, errors.BadRequestErrorf(errors.EthCallError, "Dial archNode error")
	}

	//init the client
	client, err := ethclient.Dial(archNode)
	rpcclient, err := rpc.Dial(archNode)
	if err != nil {
		blockslogger.Warningf("Dial archNode error!")
		return nil, errors.BadRequestErrorf(errors.EthCallError, "Dial archNode error: %v", err)
	}
	defer client.Close()
	defer rpcclient.Close()

	// retrieve the block, which includes all of the transactions
	block, err := client.BlockByNumber(context.TODO(), blockNumber)
	if err != nil {
		blockslogger.Warningf("Error getting block %v by number: %v", blockNumber, err)
		return nil, errors.BadRequestErrorf(errors.EthCallError, "Error getting block %v by number: %v", blockNumber, err)
	}

	//scrape all the transaction fees and return the blockFee
	txs := block.Transactions()
	gasFee := big.NewInt(0)
	if len(txs) > 0 {
		batch := make([]rpc.BatchElem, len(txs))
		for i, tx := range txs {
			batch[i] = rpc.BatchElem{
				Method: "eth_getTransactionReceipt",
				Args:   []interface{}{tx.Hash()},
				Result: new(types.Receipt),
			}
		}
		if err := rpcclient.BatchCall(batch); err != nil {
			return nil, fmt.Errorf("failed to get tx receipts: %v", err)
		}
		for i, tx := range txs {
			txFee := new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(batch[i].Result.(*types.Receipt).GasUsed)))
			gasFee = gasFee.Add(gasFee, txFee)
		}
	}
	return gasFee, nil
}


//GetBlockFee Deprecated adds the transactions and blockFee for ThisBlockNum into the BlockchainInfo struct
//func GetBlockFee(client *ethclient.Client, archNode string, blockNumber *big.Int) (blockFee *big.Int, err error){
//	if blockNumber == nil {
//		blockslogger.Warningf("No block number to retrieve transactions from")
//		return
//	}
//
//	//init the Client
//	//Client, err := ethclient.Dial(archNode)
//	//if err != nil {
//	//	blockslogger.Warningf("Dial archNode error!")
//	//	return nil, errors.BadRequestErrorf(errors.EthCallError, "Dial archNode error: %v", err)
//	//}
//	//defer Client.Close()
//
//	// retrieve the block, which includes all of the transactions
//	block, err := client.BlockByNumber(context.TODO(), blockNumber)
//	if err != nil {
//		blockslogger.Warningf("Error getting block %v by number: %v", blockNumber, err)
//		return nil, errors.BadRequestErrorf(errors.EthCallError, "Error getting block %v by number: %v", blockNumber, err)
//	}
//
//	//scrape all the transaction fees and return the blockFee
//	blockFee = fetchBlkFees(client, archNode, []*types.Transaction(block.Transactions()))
//
//	return blockFee, nil
//}


func getBlockNumber(archNode string) (*big.Int, error) {
	RPCClient, err:= rpc.Dial(archNode)
	if err != nil {
	return nil, errors.BadRequestErrorf(errors.EthCallError, "RPC Dial node error: %v", err)
	}
	defer RPCClient.Close()

	var lastBlockStr string
	err = RPCClient.Call(&lastBlockStr, "eth_blockNumber")
	if err != nil {
		blockslogger.Errorf("Can't get latest block: %v", err)
		return nil, errors.BadRequestErrorf(errors.EthCallError, "Can't get latest block: %v", err)
	}

	// translate from string (hex probably) to *big.Int
	lastBlockNum := big.NewInt(0)
	if _, ok := lastBlockNum.SetString(lastBlockStr, 0); !ok {
		blockslogger.Errorf("Unable to parse last block string: %v", lastBlockStr)
		return nil, errors.BadRequestErrorf(errors.EthCallError, "Unable to parse last block string: %v", lastBlockStr)
	}
	return lastBlockNum, nil
}


//func retrieveTx(client *ethclient.Client, archNode string, tx *types.Transaction) (txfee *big.Int, erc error){
//	//Client, err := ethclient.Dial(archNode)
//	//defer Client.Close()
//	//if err != nil {
//	//	blockslogger.Warningf("Dial archNode error!")
//	//	return nil, errors.BadRequestErrorf(errors.EthCallError, "Dial archNode error: %v", err)
//	//}
//	receipt, err := client.TransactionReceipt(context.TODO(), tx.Hash())
//	if err != nil {
//		blockslogger.Warningf("Error getting transaction receipt: %v", err)
//		return nil, errors.BadRequestErrorf(errors.EthCallError, "Error getting transaction receipt: %v", err)
//	}
//	fee := big.NewInt(0).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed)))
//	return fee, nil
//}


//func fetchBlkFees(client *ethclient.Client,archNode string, transactions []*types.Transaction) *big.Int {
//	totals := make(chan *big.Int)
//	var wg sync.WaitGroup // number of working goroutines
//	for _, tr := range transactions {
//		wg.Add(1)
//		// worker
//		go func(tr *types.Transaction) {
//			defer wg.Done()
//			txFee, err := retrieveTx(client,archNode, tr)
//			if err != nil {
//				blockslogger.Errorf("Get the goroutine function error %v", err)
//				return
//			}
//			if txFee == nil {
//				txFee = big.NewInt(0)
//			}
//			totals <- txFee
//		}(tr)
//	}
//
//	// closer
//	go func() {
//		wg.Wait()
//		close(totals)
//	}()
//
//	total := big.NewInt(0)
//	for size := range totals {
//		total.Add(total, size)
//	}
//	return total
//}


//func fetchEpochFees(archNode string, transactions []*types.Transaction) *big.Int {
//	totals := make(chan *big.Int)
//	var wg sync.WaitGroup // number of working goroutines
//	for _, tr := range transactions {
//		wg.Add(1)
//		// worker
//		go func(tr *types.Transaction) {
//			defer wg.Done()
//			txFee, err := retrieveTx(archNode, tr)
//			if err != nil {
//				blockslogger.Errorf("Get the goroutine function error %v", err)
//				return
//			}
//			if txFee == nil {
//				txFee = big.NewInt(0)
//			}
//			totals <- txFee
//		}(tr)
//	}
//
//	// closer
//	go func() {
//		wg.Wait()
//		close(totals)
//	}()
//
//	total := big.NewInt(0)
//	for size := range totals {
//		total.Add(total, size)
//	}
//	return total
//}


//sum the big int
func sum(a []*big.Int) *big.Int{
	total := big.NewInt(0)
	for _, v := range a {
		total.Add(total,v)
	}
	return total
}