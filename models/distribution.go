package models

import (
	"context"
	"github.com/op/go-logging"
	"github.com/starslabhq/rewards-collection/errors"
	"gorm.io/gorm"
	"math/big"
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
func PreSend(ctx context.Context) (bool, map[string]*big.Int, error){
	valDs, err := PumpDistInfo(ctx)
	if err != nil {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return false, nil, err
	}
	if len(valDs) == 0 {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return false, nil, err
	}

	var startEpoch int64

	valmap := make(map[string]*big.Int)
	for _, vald := range valDs {
		valmap[vald.ValAddr] = vald.Distribution
		startEpoch = vald.ThisEpoch
	}
	
	//todo some basic check before sending
	/*
	check before sending
	*/

	distributionlogger.Infof("Begin to send validator rewards info from epoch %d", startEpoch)

	return true, valmap, nil
}

//todo signing service gateway integration
//SendDistribution to send distribution to gateway signing service
//func SendDistribution() (nonce uint64)  {
//	
//}

//todo ContractEventListening return map[string]bool with address distribution send success or not
//func ContractEventListening() (map[string]bool. error){
//	/*
//	eth_newFilter
//
//	eth_getFilterChanges
//
//	*/
//}


//PostSend
//func PostSend() error {
//	/*
// 	//mapValStatus := map[string]bool
//	mapValStatus, err := ContractEventListening()
//	if err != nil{
//	error handling
//	}
//
//	for val, ok := range mapValStatus{
//		if ok {
//			upNum, err := updateDisInDB(ctx, val){
//				if err != nil{
//					error handling
//				}
//				if upNum != 36{
//					updateDB checklist
//				}
//			}
//		}
//
//	}
//
//	//abnormal solution
//	//resend with fixed nonce, higher gasprice
//	//some solution here
//
//	*/
//
//}

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
func PumpDistInfo(ctx context.Context) ([]*ValDist, error) {
	//todo use node election contract phase III abi interface to get validators
	//vals := rpcCongressGetAllVals()
	var valsDists []*ValDist
	vals := []string{}
	for _, val := range vals {
		valDist, err := fetchValDist(ctx, val)
		if err != nil {
			distributionlogger.Errorf("Fetch validator distribution error %v", err)
			continue
		}
		valsDists = append(valsDists, valDist)
	}
	return valsDists, nil
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
func fetchValToDisWithinEP(ctx context.Context, valAddr string, epStart, epEnd int64) (*ValDist, error) {
	rws := []Reward{}
	valds := []*big.Int{}
	eplist := []int64{}
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
		ThisEpoch: epStart,
		LastEpoch: epEnd,
	}, nil
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