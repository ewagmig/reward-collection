package models

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/op/go-logging"
	"github.com/starslabhq/rewards-collection/errors"
	"github.com/starslabhq/rewards-collection/utils"
	"gorm.io/gorm"
	"math/big"
	"strings"
)

var (
	distributionlogger = logging.MustGetLogger("rewards.distribution.models")
	EPDuration = int64(36)
)

type ValDist struct {
	ValAddr			string
	Distribution 	*big.Int
	ThisEpoch		int64
	LastEpoch		int64
}

//[]ValMapRewards is used to send to contract for rewards distribution, two slices: []string{}, []big.Int
type ValMapRewards struct {
	ValAddr			string
	Rewards 		*big.Int
}

//PreSend to pump distribution from database, then take some check before sending
func PreSend(ctx context.Context, epStart, epEnd uint64, archiveNode string) (bool, map[string]*big.Int, error){
	valmap, err := PumpDistInfo(ctx, epStart, epEnd, archiveNode)
	if err != nil {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return false, nil, err
	}
	if len(valmap) == 0 {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return false, nil, err
	}


	//todo some basic check before sending
	/*
	check before sending
	*/

	distributionlogger.Infof("Begin to send validator rewards info from epoch %d", epStart)

	return true, valmap, nil
}

//todo signing service gateway service integration
//todo logic on the resend
//SendDistribution to send distribution to gateway signing service
//func SendDistribution() (nonce uint64)  {
//	getTransactionReceipt
//}

//todo check if send or not
//SignTxToContract, just for Testing
//func SignTxToContract(data string) {
//	archNode := "https://http-testnet.hecochain.com"
//	client, err := ethclient.Dial(archNode)
//	if err != nil {
//		return
//	}
//	defer client.Close()
//
//
//
//}


//todo check send success or not
//ContractEventListening to trace the log of event NotifyRewardSummary after the contract notifyReward
func ContractEventListening(archnode, txhash string) (uint64, uint64, error){
	//use archnode instead for active tracing
	client, err := ethclient.Dial(archnode)
	if err != nil {
		return 0, 0, err
	}
	defer client.Close()
	receipt, err := client.TransactionReceipt(context.TODO(), common.Hash(utils.HexToHash(txhash)))
	if err != nil{
		return 0,0, err
	}

	//catch the receipt status
	if receipt.Status == uint64(1){
		distributionlogger.Debugf("The transaction is success!")
	}
	//take action to handle the receipt logs
	//event NotifyRewardSummary(uint256 inputLength, uint256 okLength), there is no indexed, only in data field
	datastr := hex.EncodeToString(receipt.Logs[0].Data)

	//split the datastr
	datastr = strings.TrimPrefix(datastr, "0x")
	inputLenStr := datastr[:64]
	inputLenStr = removeConZero(inputLenStr)
	inLen, err := utils.DecodeUint64("0x" + inputLenStr)
	if err != nil {
		return 0, 0, nil
	}

	okLenStr := datastr[64:128]
	okLenStr = removeConZero(okLenStr)
	okLen, err := utils.DecodeUint64("0x" + okLenStr)
	if err != nil {
		return 0, 0, nil
	}

	//todo logic
	distributionlogger.Infof("The input pools number is %d, and the success execution in contract is %d", inLen, okLen)
	if inLen != okLen {
		distributionlogger.Errorf("There have been data mismatch during execution in contract!")
	}

	return inLen, okLen, nil
}


//PostSend
func PostSend(ctx context.Context) error {
	//mapValStatus := map[string]bool
	//todo make some coordination on the input params
	archnode, txhash := "", ""
	inLen, okLen, err := ContractEventListening(archnode, txhash)
	if err != nil{
		return err
	}

	if inLen != okLen {
		distributionlogger.Debugf("There have been data mismatch during execution in contract! Take some action to check!")

	}

	//normal process
	vals := []*ValDist{}
	for _, valD := range vals {
		affectedRows, err := updateDisInDB(ctx, valD)
		if affectedRows == int64(0) || err != nil {
			distributionlogger.Errorf("The updating distributed flag error with val addr %s", valD.ValAddr)
			continue
		}
	}



	//abnormal solution
	//resend with fixed nonce, higher gasprice
	//some solution here

	distributionlogger.Infof("Begin to send validator rewards info from epoch %d", vals[0].ThisEpoch)
	return nil

}

//updateDisInDB to update distribution in Database
func updateDisInDB(ctx context.Context, valD *ValDist) (int64, error) {
	rw := Reward{}
	eplist := []int64{}
	deltaEP := valD.LastEpoch - valD.ThisEpoch + 1
	for i := valD.ThisEpoch; i <= valD.LastEpoch; i ++ {
		eplist = append(eplist, i)
	}
	db := MDB(ctx).Model(&rw).Where("validator_addr = ? and epoch_index IN ?", valD.ValAddr, eplist).Updates(map[string]interface{}{"distributed": true})
	if db.RowsAffected != deltaEP || db.Error != nil {
		distributionlogger.Errorf("Update distribution in db error")
		return 0, errors.BadRequestError(errors.DatabaseError, "Update distribution in db error")
	}
	return db.RowsAffected, nil
}

//PumpDistInfo to pump the distribution info from database
func PumpDistInfo(ctx context.Context, epStart, epEnd uint64, archiveNode string) (map[string]*big.Int, error) {
	valMapDist := make(map[string]*big.Int)
	//get the vals at the end of this period
	vals, err  := rpcCongressGetAllVals(epEnd, archiveNode)
	if err != nil {
		return nil, err
	}
	for _, val := range vals{
		valdis, err := fetchValToDisWithinEP(ctx,val,epStart,epEnd)
		if err != nil {
			return nil, err
		}
		valMapDist[val] = valdis.Distribution
	}

	return valMapDist, nil
}


//fetchValDist to fetch all the distribution rewards through some EPs
func fetchValDist(ctx context.Context, valAddr string) (*ValDist, error) {
	valds := []*big.Int{}
	rws := []Reward{}
	//fetch the undis reward
	edrw, err := fetchValEdDis(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	//fetch the todis reward
	torw, err := fetchValToDis(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	rw := MDB(ctx).Where("distributed = ? and validator_addr = ? and epoch_index > ? and epoch_index <= ?", 0, valAddr, edrw.EpochIndex, torw.EpochIndex).FindInBatches(&rws, 36, func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})
	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)

	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
		ThisEpoch: edrw.EpochIndex + int64(1),
		LastEpoch: torw.EpochIndex,
	}, nil
}

//fetchValToDisWithinEP to fetch val epoch rewards during a epoch range
func fetchValToDisWithinEP(ctx context.Context, valAddr string, epStart, epEnd uint64) (*ValDist, error) {
	rws := []Reward{}
	valds := []*big.Int{}
	eplist := []uint64{}
	for i := epStart; i <= epEnd; i ++ {
		eplist = append(eplist, i)
	}
	deltaEP := epEnd - epStart + 1
	//db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	rw := MDB(ctx).Where("validator_addr = ? and epoch_index IN ?", valAddr, eplist).FindInBatches(&rws, int(deltaEP), func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})

	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)
	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
		ThisEpoch: int64(epStart),
		LastEpoch: int64(epEnd),
	}, nil
}

func FetchTotalRewardsEPs(ctx context.Context, epStart, epEnd uint64) (*big.Int, error) {
	eps := []Epoch{}
	total := []*big.Int{}
	eplist := []uint64{}
	for i := epStart; i <= epEnd; i ++ {
		eplist = append(eplist, i)
	}
	deltaEP := epEnd - epStart + 1

	ep := MDB(ctx).Where("epoch_index IN ?", eplist).FindInBatches(&eps, int(deltaEP), func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, ep := range eps{
			rwbig, ok := new(big.Int).SetString(ep.TotalFees, 10)
			if ok{
				total = append(total, rwbig)
			}
		}
		return nil
	})
	distributionlogger.Debugf("The rows affected should be %d", ep.RowsAffected)

	totald := sum(total)
	return totald, nil
}


//just for UT
func fetchValDistForUT(ctx context.Context, EPs int, valAddr string, db *gorm.DB) (*ValDist, error) {
	valds := []*big.Int{}
	rws := []Reward{}
	//fetch the distributed already reward
	edrw, err := fetchValEdDisUT(ctx, valAddr, db)
	if err != nil {
		return nil, err
	}

	//fetch the todis reward
	torw, err := fetchValToDisUT(ctx, valAddr, db)
	if err != nil {
		return nil, err
	}

	rw := db.Where("distributed = ? and validator_addr = ? and epoch_index > ? and epoch_index <= ?", 0, valAddr, edrw.EpochIndex, torw.EpochIndex).FindInBatches(&rws, EPs, func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})
	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)

	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
	}, nil
}


func fetchValEdDis(ctx context.Context, valAddr string) (*Reward, error) {
	rw := &Reward{}
	MDB(ctx).Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 1, valAddr).First(&rw)
	return rw, nil
}

func fetchValToDis(ctx context.Context, valAddr string) (*Reward, error) {
	rw := &Reward{}
	MDB(ctx).Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	return rw, nil
}

//just for UT
func fetchValEdDisUT(ctx context.Context, valAddr string, db *gorm.DB) (*Reward, error) {
	rw := &Reward{}
	db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 1, valAddr).First(&rw)
	return rw, nil
}

//just for UT
func fetchValToDisUT(ctx context.Context, valAddr string, db *gorm.DB) (*Reward, error) {
	rw := &Reward{}
	db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	return rw, nil
}

//jsut for UT within EP range
func fetchValToDisWithinEPUT(ctx context.Context, valAddr string, db *gorm.DB, epStart, epEnd int64) (*ValDist, error) {
	rws := []Reward{}
	valds := []*big.Int{}
	eplist := []int64{}
	for i := epStart; i <= epEnd; i ++ {
		eplist = append(eplist, i)
	}
	deltaEP := epEnd - epStart + 1
	//db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	rw := db.Where("validator_addr = ? and epoch_index IN ?", valAddr, eplist).FindInBatches(&rws, int(deltaEP), func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})

	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)
	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
		ThisEpoch: epStart,
		LastEpoch: epEnd,
	}, nil
}

//just for UT
func updateDisInDBUT(ctx context.Context, valD *ValDist, tx *gorm.DB) (int64, error) {
	rw := Reward{}
	eplist := []int64{}
	deltaEP := valD.LastEpoch - valD.ThisEpoch + 1
	for i := valD.ThisEpoch; i <= valD.LastEpoch; i ++ {
		eplist = append(eplist, i)
	}

	db := tx.Model(&rw).Where("validator_addr = ? and epoch_index IN ?", valD.ValAddr, eplist).Updates(map[string]interface{}{"distributed": true})

	if db.RowsAffected != deltaEP || db.Error != nil {
		distributionlogger.Errorf("Update distribution in db error")
		return 0, errors.BadRequestError(errors.DatabaseError, "Update distribution in db error")
	}
	return db.RowsAffected, nil
}