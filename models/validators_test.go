package models

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	_ "github.com/go-sql-driver/mysql" // inject mysql driver to go sql
	"github.com/starslabhq/rewards-collection/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/big"
	"modernc.org/sortutil"
	"testing"
)

const (
	connStr = "root:12345678@tcp(huobichain-dev-02.sinnet.huobiidc.com:3306)/heco_test?charset=utf8&parseTime=True&loc=Local"
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

//func TestSaveEpochData(t *testing.T) {
//	archNode := "https://http-testnet.hecochain.com"
//	blockInfo := ScramChainInfo(archNode)
//	t.Log(blockInfo.EpochIndex)
//	fees := GetBlockEpochRewards(archNode,blockInfo.EpochIndex)
//	blockInfo.TotalFees = fees
//
//	db, err := InitDB(connStr)
//	if err != nil {
//		t.Error(err)
//	}
//
//	epochs := &Epoch{
//		EpochIndex: int64(blockInfo.EpochIndex),
//		ThisBlockNumber: blockInfo.ThisBlockNum.Int64(),
//		LastBlockNumber: blockInfo.LastBlockNum.Int64(),
//		TotalFees: blockInfo.TotalFees.String(),
//	}
//
//	err = db.Create(epochs).Error
//	if err != nil {
//		t.Error(err)
//	}
//}

//func TestEthcallVal(t *testing.T) {
//	archiveNode := "https://http-testnet.hecochain.com"
//	blkNumHex := "latest"
//	valAddr := "0x6301cdf018E8678067cf8f14aB99F6f2a906dB44"
//
//	valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, blkNumHex, valAddr)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(valInfo)
//}

func TestGetActVals(t *testing.T) {
	archiveNode := "https://http-mainnet-node.defibox.com"
	blkNumHex := utils.EncodeUint64(uint64(4813199))
	vals, err := jsonrpcEthCallGetActVals(archiveNode, blkNumHex)
	if err != nil {
		t.Error(err)
	}
	t.Log(vals)
}

//func TestGetAllVals(t *testing.T) {
//	archiveNode := "http://localhost:8545"
//	blkNumHex := "latest"
//	vals, err := rpcCongressGetAllVals(archiveNode, blkNumHex)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(vals)
//}

func TestStrSplit(t *testing.T) {
	valInfo := "0x000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000394c549ef0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000000"
	resp, err := splitValInfo(valInfo)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

//func TestStrSplitArr(t *testing.T) {
//	vals := "0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000005000000000000000000000000192bbe6143d57fee4d0e6fd6ec55d9c83bd5d6c90000000000000000000000001aa397e02fb3abba1072b431e92b0f90fe60993c00000000000000000000000038e439a4abead544e0f11a323d4091f58f5431ad000000000000000000000000b4675e493f17b84828e70f18fddce3c55ec67d6f000000000000000000000000c48bfe79065ddfd8d84d535f47c480bf38d568ce"
//	resp, err := splitVals(vals)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(resp)
//
//	val0 := resp[0]
//	t.Log(val0)
//}

//func TestGetRewardAtBlk(t *testing.T) {
//	ArchiveNode := "http://localhost:8545"
//	BlkNum := uint64(41096)
//	totalRewards, err := GetRewardsAtBlock(ArchiveNode,BlkNum)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(totalRewards)
//}

func TestRemoveConZero(t *testing.T) {
	str := "00000000000"
	resp := removeConZero(str)
	t.Log(resp)
}

//func TestGetDeltaRewards(t *testing.T) {
//	epochIndex := uint64(7)
//	archiveNode := "http://localhost:8545"
//	resp, err := getDeltaRewards(epochIndex, archiveNode)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(resp)
//
//}

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

//func TestGetRewards(t *testing.T) {
//	params := &CallParams{
//		EpochIndex: uint64(24166),
//		ArchiveNode: "https://http-testnet.hecochain.com",
//	}
//	resp, err := GetRewards(params)
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(resp)
//
//}

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

func TestBigSort(t *testing.T) {
	bigs := sortutil.BigIntSlice{big.NewInt(400), big.NewInt(200), big.NewInt(200), big.NewInt(300)}
	bigs.Sort()
	t.Log(bigs)

	small2 := bigs[:2]
	big3 := bigs[len(bigs)-3:]

	t.Log(small2)
	t.Log(big3)
}

func TestGet(t *testing.T) {
	archNode := "https://http-testnet.hecochain.com"
	epIndex := uint64(25037)
	params := &CallParams{
		ArchiveNode: archNode,
		EpochIndex: epIndex,
	}

	valinfo, err := GetRewards(params)
	if err != nil {
		t.Error(err)
	}
	t.Log(valinfo)
}

func TestGetVals(t *testing.T) {
	archNode := "https://http-testnet.hecochain.com"
	blockHex := "latest"

	vals, err := jsonrpcEthCallGetActVals(archNode, blockHex)
	if err != nil {
		t.Error(err)
	}
	t.Log(vals)
}

func TestCalcu(t *testing.T) {
	archNode := "https://http-testnet.hecochain.com"
	epIndex := uint64(25361)
	rewards := big.NewInt(116011615870000000)
	val, err := calcuDistInEpoch(epIndex, rewards, archNode)
	if err != nil {
		t.Error(err)
	}
	t.Log(val)
}

func TestSignGateway(t *testing.T) {
	valMapDist := make(map[string]*big.Int)
	valMapDist["000000000000000000000000532f39e49dc1a7f154a1d08ad6eaba6b0aa49a16"] = big.NewInt(643498595238095)
	dataStr, amstr := getNotifyAmountData(valMapDist)
	t.Log("The data string is", dataStr)
	t.Log("The amount string is", amstr)

	archNode := "https://http-testnet.hecochain.com"
	//sysAddr := "0xe2cdcf16d70084ac2a9ce3323c5ad3fa44cddbda"
	signGateway(archNode, sysAddr, valMapDist)
}

func TestCalcUT(t *testing.T) {
	var bigSort sortutil.BigIntSlice
	vals := []*big.Int{big.NewInt(64), big.NewInt(10), big.NewInt(3),big.NewInt(2), big.NewInt(2), big.NewInt(2), big.NewInt(1), big.NewInt(1), big.NewInt(1)}
	for _, v := range vals {
		bigSort = append(bigSort, v)
	}
	bigSort.Sort()
	rewards := big.NewInt(116011615870000000)

	rewardsPerNum := big.NewInt(0)
	rewardsPerNum.Mul(rewards, big.NewInt(1))
	rewardsPerNum.Div(rewards, big.NewInt(42))


	rewardsPerStakingCoins := new(big.Int)
	rewardsDouble :=new(big.Int)
	rewardsDouble.Mul(rewards, new(big.Int).SetInt64(int64(2)))
	rewardsPerStakingCoins.Div(rewardsDouble, new(big.Int).SetInt64(int64(5)))

	totalActCoins := sum(bigSort)


	reward1 := big.NewInt(0)
	reward1.Mul(big.NewInt(64), rewardsPerStakingCoins)
	reward1.Div(reward1, totalActCoins)

	reward1.Add(reward1, rewardsPerNum)
	t.Log(reward1)
}

func TestSortBigTable(t *testing.T) {
	slice := []int{1,2,3,4,5,5,6,7}
	l := len(slice)

	t.Log(slice)
	//t.Log(slice[:l-3])
	//t.Log(slice[l-3:])
	//t.Log(slice[l-5])
	//t.Log(slice[:3])
	t.Log(slice[l-1:])
}

func TestDecodeData(t *testing.T) {
	rawTXstring := "f9064b808509502f9000830f4240945caef96c490b5c357847214395ca384dc3d3b85e80b905e4f141d389000000000000000000000000000000000000000000000000fc79ee2d5d6746e2280a0e57834b0390c0e0257c0000000000000000000000000000000000000000000000007567e32efe4812f02d19216fe9c0709e4b7c612b000000000000000000000000000000000000000000000000e11a5fe5cfdf07373e1130628cc25b76580337b4000000000000000000000000000000000000000000000000b88c622adf4a878a7caee487c1356e13421b8267000000000000000000000000000000000000000000000000b763487cbed3ac6a7401f7740d19b4401b948402000000000000000000000000000000000000000000000000d87a1c95b941f633a7bc0735a57d8e90f9a187b70000000000000000000000000000000000000000000000007be5c02f3569f57d519621b68c8953fa9f2c071f000000000000000000000000000000000000000000000000cff389d6791d05a47bc3fc59319e7fbb01026a65000000000000000000000000000000000000000000000000f12a627bd37a326fb3e628017cf04538cf9a96250000000000000000000000000000000000000000000000000c7c3f651ca16346d000eeb9b6f78997ec9be28d0000000000000000000000000000000000000000000000004706040ed70f257288b1acc04de74372720025ba0000000000000000000000000000000000000000000000002512d871e388a97e35f5ae8c46344077166c9a07000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000fc79ee2d5d6746e2280a0e57834b0390c0e0257c0000000000000000000000000000000000000000000000007567e32efe4812f02d19216fe9c0709e4b7c612b000000000000000000000000000000000000000000000000e11a5fe5cfdf07373e1130628cc25b76580337b4000000000000000000000000000000000000000000000000b88c622adf4a878a7caee487c1356e13421b8267000000000000000000000000000000000000000000000000b763487cbed3ac6a7401f7740d19b4401b948402000000000000000000000000000000000000000000000000d87a1c95b941f633a7bc0735a57d8e90f9a187b70000000000000000000000000000000000000000000000007be5c02f3569f57d519621b68c8953fa9f2c071f000000000000000000000000000000000000000000000000cff389d6791d05a47bc3fc59319e7fbb01026a65000000000000000000000000000000000000000000000000f12a627bd37a326fb3e628017cf04538cf9a96250000000000000000000000000000000000000000000000000c7c3f651ca16346d000eeb9b6f78997ec9be28d0000000000000000000000000000000000000000000000004706040ed70f257288b1acc04de74372720025ba0000000000000000000000000000000000000000000000002512d871e388a97e35f5ae8c46344077166c9a07000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000fc49c20acf698f000000000000000000000000000000000000000000000000005ba8f64bd9c78a00000000000000000000000000000000000000000000000000261e07b6dd3c330000000000000000000000000000000000000000000000000030d36aa17624de0000000000000000000000000000000000000000000000000030d36aa17624de00000000000000000000000000000000000000000000000000261e07b6dd3c3300000000000000000000000000000000000000000000000000261e07b6dd3c330000000000000000000000000000000000000000000000000030d36aa17624de0000000000000000000000000000000000000000000000000030d36aa17624de0000000000000000000000000000000000000000000000000030d36aa17624de000000000000000000000000000000000000000000000000003b88cd8c0f0d8900000000000000000000000000000000000000000000000000731de759ebc5401ca0b936fd524a800b4fa0d027ac20d22410b617db7040c24d1f8022ef592016ffbca047719611f72fb5bdcaa85c31e36c47398c9730381d44bd25a5f7f0bb691ac699"
	var tx types.Transaction
	rawtx,err := hex.DecodeString(rawTXstring)
	if err != nil{
		t.Error(err)
	}
	rlp.DecodeBytes(rawtx, &tx)

	t.Log(tx.Hash().Hex())

}

func TestValidator(t *testing.T) {
	valMapDist := make(map[string]*big.Int)
	valMapDist["000000000000000000000000532f39e49dc1a7f154a1d08ad6eaba6b0aa49a16"] = big.NewInt(643498595238095)
	archNode := "https://http-testnet.hecochain.com"

	encResp, err := signGateway(archNode, sysAddr, valMapDist)
	if err != nil {
		t.Error(err)
	}

	encData := encResp.Data.EncryptData
	t.Log("The enc signed tx is", encData)

	targetUrl := "http://huobichain-dev-02.sinnet.huobiidc.com:5005/validate/cross/check"
	accKey := Key{
		AccessKey: AccessKey,
		SecretKey: SecretKey,
	}

	validaReq := ValidatorReq{
		EncryptData: encResp.Data.EncryptData,
		Cipher: encResp.Data.Extra.Cipher,
	}

	rawTx, ok := ValidateEnc(validaReq, targetUrl, accKey)

	t.Log("The raw tx is", rawTx)
	t.Log("The Ok status is", ok)

	rpcClient, _ := rpc.Dial(archNode)
	_ = rpcClient.CallContext(context.Background(),nil,"eth_sendRawTransaction", rawTx)

}

func TestGetRewardsInEPs(t *testing.T) {
	rs := []string{"3455121063333332", "3455121063333332", "4442298509999998", "4442298509999998", "4442298509999998", "4442298509999998", "5429475956666664", "8391008296666662", "23198669996666652", "4442298509999998", "2467943616666666", "2726113765714285", "2726113765714285", "5149326001904761", "14236371887619046", "2120310706666666", "2120310706666666", "2726113765714285", "2726113765714285", "2726113765714285", "3331916824761904", "1514507647619047", "2025642211666666", "2604397129285713", "2604397129285713", "3183152046904760", "2025642211666666", "2604397129285713", "2604397129285713", "2604397129285713", "4919416799761901", "13600740564047606", "1446887294047619"}

	var rbig []*big.Int
	for _, v := range rs{
		rwbig, ok := new(big.Int).SetString(v, 10)
		if ok{
			rbig = append(rbig, rwbig)
		}
	}

	total := sum(rbig)
	t.Log(total)

}

func TestGetDataAmount(t *testing.T) {
	valMapDist := make(map[string]*big.Int)

	valMapDist["0000000000000000000000000c7c3f651ca16346d000eeb9b6f78997ec9be28d"] = big.NewInt(5330510894999998)
	valMapDist["0000000000000000000000002512d871e388a97e35f5ae8c46344077166c9a07"] = big.NewInt(4145952918333332)
	//valMapDist["0000000000000000000000004706040ed70f257288b1acc04de74372720025ba"] = big.NewInt(6515068871666664)
	//valMapDist["0000000000000000000000007567e32efe4812f02d19216fe9c0709e4b7c612b"] = big.NewInt(5330510894999998)
	//valMapDist["0000000000000000000000007be5c02f3569f57d519621b68c8953fa9f2c071f"] = big.NewInt(10068742801666662)
	//valMapDist["000000000000000000000000b763487cbed3ac6a7401f7740d19b4401b948402"] = big.NewInt(5330510894999998)
	//valMapDist["000000000000000000000000b88c622adf4a878a7caee487c1356e13421b8267"] = big.NewInt(2961394941666666)
	//valMapDist["000000000000000000000000d87a1c95b941f633a7bc0735a57d8e90f9a187b7"] = big.NewInt(27837112451666652)
	//valMapDist["000000000000000000000000e11a5fe5cfdf07373e1130628cc25b76580337b4"] = big.NewInt(4145952918333332)
	//valMapDist["000000000000000000000000f12a627bd37a326fb3e628017cf04538cf9a9625"] = big.NewInt(5330510894999998)
	//valMapDist["000000000000000000000000fc79ee2d5d6746e2280a0e57834b0390c0e0257c"] = big.NewInt(5330510894999998)



	dataStr, amstr := getNotifyAmountData(valMapDist)

	t.Log(dataStr, amstr)
}

func TestCalcuHex(t *testing.T) {
	hexstr := "0000000000000000000000000000000000000000000000000012f012485e29be"
	hexstr = "0x" + removeConZero(hexstr)
	t.Log("Hex string is", hexstr)

	big, _ := hexutil.DecodeBig(hexstr)
	t.Log(big.String())

}

func TestGetRewardsInEPsut(t *testing.T) {
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}
	valAddr := "000000000000000000000000d87a1c95b941f633a7bc0735a57d8e90f9a187b7"
	epStart := int64(26080)
	epEnd := int64(26082)
	valDist, err := fetchValToDisWithinEPUT(context.Background(), valAddr, db, epStart, epEnd)
	if err != nil {
		t.Error(err)
	}
	t.Log(valDist)
}