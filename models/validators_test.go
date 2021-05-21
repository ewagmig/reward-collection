package models

import (
	"context"
	_ "github.com/go-sql-driver/mysql" // inject mysql driver to go sql
	"github.com/starslabhq/rewards-collection/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/big"
	"testing"
)

const (
	connStr = "root:12345_@tcp(localhost:3306)/heco_test?charset=utf8&parseTime=True&loc=Local"
)

func InitDB(source string) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.New(mysql.Config{DSN:source}), &gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		return nil, err
	}
	sql,err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sql.SetMaxIdleConns(0)
	return gdb, err
}

func TestSaveEpochData(t *testing.T) {
	archNode := "https://http-testnet.hecochain.com"
	blockInfo := ScramChainInfo(archNode)
	t.Log(blockInfo.EpochIndex)
	fees := GetBlockEpochRewards(archNode,blockInfo.EpochIndex)
	blockInfo.TotalFees = fees

	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}

	epochs := &Epoch{
		EpochIndex: int64(blockInfo.EpochIndex),
		ThisBlockNumber: blockInfo.ThisBlockNum.Int64(),
		LastBlockNumber: blockInfo.LastBlockNum.Int64(),
		TotalFees: blockInfo.TotalFees.String(),
	}

	err = db.Create(epochs).Error
	if err != nil {
		t.Error(err)
	}
}

func TestEthcallVal(t *testing.T) {
	archiveNode := "https://http-testnet.hecochain.com"
	blkNumHex := "latest"
	valAddr := "0x6301cdf018E8678067cf8f14aB99F6f2a906dB44"

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
	BlkNum := uint64(41096)
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
	archNode := "http://localhost:8545"
	epochIndex := uint64(41096)
	resp := GetBlockEpochRewards(archNode, epochIndex)
	//t.Log(resp.ThisBlockNum, resp.LastBlockNum, resp.EpochIndex)
	t.Log(resp)
}

func TestScramChainInfo(t *testing.T) {
	archNode := "https://http-testnet.hecochain.com"
	resp := ScramChainInfo(archNode)
	t.Log(resp.ThisBlockNum, resp.LastBlockNum, resp.EpochIndex)
}

func TestSum(t *testing.T) {
	a := []*big.Int{big.NewInt(1), big.NewInt(3), big.NewInt(5)}
	resp := sum(a)
	t.Log(resp)
}

func TestGetTxFeesByBatch(t *testing.T) {
	archNode := "http://localhost:8545"
	blkNum := big.NewInt(41096)
	resp, err := getBlockFeesByBatch(archNode, blkNum)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)

}

func TestGetRewards(t *testing.T) {
	params := &CallParams{
		EpochIndex: uint64(24166),
		ArchiveNode: "https://http-testnet.hecochain.com",
	}
	resp, err := GetRewards(params)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)

}

func TestSaveEpData(t *testing.T) {
	blkhelper := &blockHelper{
		ArchNode: "https://http-testnet.hecochain.com",
	}
	epIndex := uint64(24452)
	ctx := context.TODO()

	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}

	err = blkhelper.SaveEpochDataForTest(ctx, epIndex, db)
	if err != nil {
		t.Error(err)
	}
}

func TestGetFeesInEP(t *testing.T) {

	ctx := context.Background()
	epIndex := uint64(24338)
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}
	fees := getFeesInEPForUT(ctx, epIndex,db)
	t.Log(fees)
}

func TestStoreRewards(t *testing.T) {
	blkhelper := &blockHelper{
		ArchNode: "https://http-testnet.hecochain.com",
	}

	ctx := context.TODO()
	epIndex := uint64(24351)
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}
	err = blkhelper.SaveValsForUT(ctx, epIndex,db)
	if err != nil {
		t.Error(err)
	}
}

func TestFetchDisFromDB(t *testing.T) {
	ctx := context.TODO()
	EPs := 10
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}
	valAddr := "0x2"
	resp, err := fetchValDistForUT(ctx, EPs, valAddr, db)
	if err != nil{
		t.Error(err)
	}
	t.Log(resp)
}

func TestUpdateDB(t *testing.T) {
	ctx := context.TODO()
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}
	valD := &ValDist{
		"0x3",
		big.NewInt(100),
		int64(24518),
		int64(24520),
	}
	resp, err := updateDisInDBUT(ctx,valD, db)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func TestFindRWResults(t *testing.T) {
	ctx := context.TODO()
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}
	epStart := int64(24518)
	epEnd := int64(24520)
	valAddr := "0x1"
	resp, err := fetchValToDisWithinEPUT(ctx, valAddr, db, epStart, epEnd)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}