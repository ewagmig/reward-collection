package models

import (
	"github.com/starslabhq/rewards-collection/utils"
	"math/big"
	"testing"
)

func TestEthcallVal(t *testing.T) {
	archiveNode := "http://localhost:8545"
	blkNumHex := "latest"
	valAddr := "0x1aa397e02fb3abba1072b431e92b0f90fe60993c"

	valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, blkNumHex, valAddr)
	if err != nil {
		t.Error(err)
	}
	t.Log(valInfo)
}

func TestGetActVals(t *testing.T) {
	archiveNode := "https://http-mainnet-node.defibox.com"
	blkNumHex := utils.EncodeUint64(uint64(4813199))
	vals, err := jsonrpcEthCallGetActVals(archiveNode, blkNumHex)
	if err != nil {
		t.Error(err)
	}
	t.Log(vals)
}

func TestGetAllVals(t *testing.T) {
	archiveNode := "http://localhost:8545"
	blkNumHex := "latest"
	vals, err := rpcCongressGetAllVals(archiveNode, blkNumHex)
	if err != nil {
		t.Error(err)
	}
	t.Log(vals)
}

func TestStrSplit(t *testing.T) {
	valInfo := "0x000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000394c549ef0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000000"
	resp, err := splitValInfo(valInfo)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func TestStrSplitArr(t *testing.T) {
	vals := "0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000005000000000000000000000000192bbe6143d57fee4d0e6fd6ec55d9c83bd5d6c90000000000000000000000001aa397e02fb3abba1072b431e92b0f90fe60993c00000000000000000000000038e439a4abead544e0f11a323d4091f58f5431ad000000000000000000000000b4675e493f17b84828e70f18fddce3c55ec67d6f000000000000000000000000c48bfe79065ddfd8d84d535f47c480bf38d568ce"
	resp, err := splitVals(vals)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)

	val0 := resp[0]
	t.Log(val0)
}

func TestGetRewardAtBlk(t *testing.T) {
	ArchiveNode := "http://localhost:8545"
	BlkNum := uint64(600)
	totalRewards, err := GetRewardsAtBlock(ArchiveNode,BlkNum)
	if err != nil {
		t.Error(err)
	}
	t.Log(totalRewards)
}

func TestRemoveConZero(t *testing.T) {
	str := "00000878000"
	resp := removeConZero(str)
	t.Log(resp)
}

func TestGetDeltaRewards(t *testing.T) {
	epochIndex := uint64(7)
	archiveNode := "http://localhost:8545"
	resp, err := getDeltaRewards(epochIndex, archiveNode)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)

}

func TestGetBlockRewards(t *testing.T) {
	archNode := "https://http-mainnet-node.defibox.com"
	epochIndex := uint64(24063)
	resp := GetBlockEpochRewards(archNode, epochIndex)
	//t.Log(resp.ThisBlockNum, resp.LastBlockNum, resp.EpochIndex)
	t.Log(resp)
}

func TestScramChainInfo(t *testing.T) {
	archNode := "https://http-mainnet-node.defibox.com"
	resp := ScramChainInfo(archNode)
	t.Log(resp.ThisBlockNum, resp.LastBlockNum, resp.EpochIndex)
}

func TestSum(t *testing.T) {
	a := []*big.Int{big.NewInt(1), big.NewInt(3), big.NewInt(5)}
	resp := sum(a)
	t.Log(resp)
}

func TestGetTxFeesByBatch(t *testing.T) {
	archNode := "https://http-mainnet-node.defibox.com"
	blkNum := big.NewInt(4811960)
	resp, err := getBlockFeesByBatch(archNode, blkNum)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)

}

func TestGetRewards(t *testing.T) {
	params := &CallParams{
		EpochIndex: uint64(24063),
		ArchiveNode: "https://http-mainnet-node.defibox.com",
	}
	resp, err := GetRewards(params)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)

}