package models

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/starslabhq/rewards-collection/errors"
	"github.com/starslabhq/rewards-collection/utils"
	"gorm.io/gorm"
	"math/big"
	"sync"
	"time"
)

const (
	RecordCreated   =	"created"
	RecordFailed	=	"failed"
	RecordSuccess	= 	"success"

)

//SendRecord is a table to store the send raw transaction record
//[Table]
type SendRecord struct {
	IDBase
	RawTx		string			`json:"raw_tx"`
	TxHash      string			`json:"tx_hash"`
	Stat      	string          `json:"stat"`
	Nonce		int64			`json:"nonce"`
	//GasPrice    int64			`json:"gas_price"`
	ThisEpoch	int64			`json:"this_epoch"`
	LastEpoch	int64			`json:"last_epoch"`
	AtBase
}

func (SendRecord)TableName() string {
	return "send_records"
}

// Reward is a reward fetching per validator and store the data in table
// [TABLE]
type Reward struct {
	IDBase
	ValidatorAddr       string     `json:"validator_addr"`
	Rewards			    string     `json:"rewards"`
	EpochIndex          int64      `json:"epoch_index"`
	Distributed         bool       `json:"distributed"`
	LastTxCreatedAt     *time.Time `json:"last_tx_created_at"`
	AtBase
}

func (Reward) TableName() string {
	return "rewards"
}

// Epoch is a reward fetching per epoch and store the data in table
// [TABLE]
type Epoch struct {
	IDBase
	EpochIndex          int64      `json:"epoch_index"`
	ThisBlockNumber     int64      `json:"this_block_number"`
	LastBlockNumber     int64      `json:"last_block_number"`
	Remaining			string	   `json:"remaining"`
	TotalFees			string	   `json:"total_fees"`
	LastTxCreatedAt     *time.Time `json:"last_tx_created_at"`
	AtBase
}

func (Epoch) TableName() string {
	return "epochs"
}

type blockHelper struct {
	ArchNode string
	mu    sync.RWMutex
}

func newBlockHelper() *blockHelper {
	archNode := viper.GetString("server.archiveNodeUrl")
	if len(archNode) == 0 {
		logrus.Errorf("No archNode config!")
		return nil
	}
	return &blockHelper{
		ArchNode: archNode,
	}
}

func (rw *Reward) BeforeCreate(tx *gorm.DB) error {
	db := tx.First(&Reward{}, "epoch_index = ? and validator_addr = ?", rw.EpochIndex, rw.ValidatorAddr)
	if db.Error != nil && db.Error.Error() == "record not found" {
		return nil
	}

	return errors.ConflictErrorf(errors.EPIndexExist, "Epoch Index %d along with Validator %s exists", rw.EpochIndex, rw.ValidatorAddr)
}

func (ep *Epoch) BeforeCreate(tx *gorm.DB) error {
	db := tx.First(&Epoch{}, "epoch_index = ?", ep.EpochIndex)
	if db.Error != nil && db.Error.Error() == "record not found" {
		return nil
	}

	return errors.ConflictErrorf(errors.EPIndexExist, "Epoch Index %d exists", ep.EpochIndex)
}

func (sr *SendRecord) BeforeCreate(tx *gorm.DB) error {
	db := tx.First(&SendRecord{}, "raw_tx = ?", sr.RawTx)
	if db.Error != nil && db.Error.Error() == "record not found" {
		return nil
	}

	return errors.ConflictErrorf(errors.EPIndexExist, "Send record %d exists", sr.Nonce)
}

//SaveVals to save vals info into database every epoch
func (helper *blockHelper)SaveVals(ctx context.Context, epochIndex uint64) error {
	rewards := getFeesInEPStore(ctx, epochIndex)
	vals, err := calcuDistInEpoch(epochIndex, rewards, helper.ArchNode)
	if err != nil{
		logrus.Errorf("Calculate rewards error '%v'", err)
		return err
	}

	logrus.Infof("[Epoch Index %d ] Start to store reward data for validators", epochIndex)

	RWs := []*Reward{}
	//use batch to insert data
	for _, val := range vals {
		rw := &Reward{
			EpochIndex: int64(val.EpochIndex),
			ValidatorAddr: val.ValAddr,
			Rewards: val.Rewards.String(),
		}
		RWs = append(RWs, rw)
	}

	//begin to create the rewards table
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	tx := MDB(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Create(RWs).Error; err != nil {
		logrus.Errorf("Create rewards error '%v'", err)
		tx.Rollback()
		return processDBErr(err, "failed to process database %v", err)
	}
	tx.Commit()

	logrus.Infof("[Epoch Index %d ] Finish to store reward data for validators", epochIndex)
	return nil

}

func SaveSendRecord(ctx context.Context, record *SendRecord) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	tx := MDB(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Create(record).Error; err != nil {
		logrus.Errorf("Create record error '%v'", err)
		tx.Rollback()
		return processDBErr(err, "Failed to store record caused by error %v", err)
	}
	tx.Commit()
	return nil
}

func UpdateSendRecord(ctx context.Context, txHash string) error {
	rw := SendRecord{}
	db := MDB(ctx).Model(&rw).Where("tx_hash = ? and stat = ?", txHash, RecordCreated).Updates(map[string]interface{}{"stat": RecordSuccess})
	if db.Error != nil {
		logrus.Errorf("Update send record in db error")
		return errors.BadRequestError(errors.DatabaseError, "Update distribution in db error")
	}
	logrus.Debugf("The updated column in send record tables is %d", db.RowsAffected)
	return nil
}

func UpdateSendRecordFailed(ctx context.Context, txHash string) error {
	rw := SendRecord{}
	db := MDB(ctx).Model(&rw).Where("tx_hash = ? and stat = ?", txHash, RecordCreated).Updates(map[string]interface{}{"stat": RecordFailed})
	if db.Error != nil {
		logrus.Errorf("Update send record in db error")
		return errors.BadRequestError(errors.DatabaseError, "Update distribution in db error")
	}
	logrus.Debugf("The updated column in send record tables is %d", db.RowsAffected)
	return nil
}


//SaveValsForUT just for UT testing
func (helper *blockHelper)SaveValsForUT(ctx context.Context, epochIndex uint64, tx *gorm.DB) error {
	rewards := getFeesInEPStoreForUT(ctx, epochIndex, tx)
	vals, err := mockCalcDisInEpoch(epochIndex, rewards)
	if err != nil{
		logrus.Errorf("Calculate rewards error '%v'", err)
	}

	logrus.Infof("[Epoch Index %d ] Start to store reward data for validators", epochIndex)

	RWs := []*Reward{}
	//use batch to insert data
	for _, val := range vals {
		rw := &Reward{
			EpochIndex: int64(val.EpochIndex),
			ValidatorAddr: val.ValAddr,
			Rewards: val.Rewards.String(),
		}
		RWs = append(RWs, rw)
	}

	//begin to create the rewards table
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	//tx := MDB(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Create(RWs).Error; err != nil {
		logrus.Errorf("Create rewards error '%v'", err)
		tx.Rollback()
		return processDBErr(err,  "Failed to create rewards caused by error %v", err)
	}
	tx.Commit()

	logrus.Infof("[Epoch Index %d ] Start to store reward data for validators", epochIndex)
	return nil

}

func saveValRewardForUT(ctx context.Context, valInfo *ValRewardsInfo, db *gorm.DB) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	//tx := MDB(ctx).Begin()
	tx := db
	defer tx.Rollback()

	valReward := &Reward{
		EpochIndex: int64(valInfo.EpochIndex),
		ValidatorAddr: valInfo.ValAddr,
		Rewards: valInfo.Rewards.String(),
	}

	if err := tx.Create(valReward).Error; err != nil {
		logrus.Errorf("Create rewards error '%v'", err)
		tx.Rollback()
		return processDBErr(err, "Failed to create rewards caused by error %v", err)
	}
	tx.Commit()
	return nil
}

func saveValReward(ctx context.Context, valInfo *ValRewardsInfo) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}
	tx := MDB(ctx).Begin()
	defer tx.Rollback()

	valReward := &Reward{
		EpochIndex: int64(valInfo.EpochIndex),
		ValidatorAddr: valInfo.ValAddr,
		Rewards: valInfo.Rewards.String(),
	}

	if err := tx.Create(valReward).Error; err != nil {
		logrus.Errorf("Create rewards error '%v'", err)
		tx.Rollback()
		return processDBErr(err,  "Failed to create rewards caused by error %v", err)
	}
	tx.Commit()
	return nil
}

//saveEpoch for server mode
func saveEpoch(ctx context.Context, info *BlockchainInfo) error {
	logrus.Infof("[Epoch Index %d ] Start to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())
	tx := MDB(ctx).Begin()
	defer tx.Rollback()

	//take action to parse table
	blockRewards := &Epoch{
		EpochIndex: int64(info.EpochIndex),
		ThisBlockNumber: info.ThisBlockNum.Int64(),
		LastBlockNumber: info.LastBlockNum.Int64(),
		TotalFees: info.TotalFees.String(),
	}

	if err := tx.Create(blockRewards).Error; err != nil {
		logrus.Errorf("Create epoch error '%v'", err)
		tx.Rollback()
		return processDBErr(err, "Failed to create epoch caused by error %v", err)
	}
	tx.Commit()
	logrus.Infof("[Epoch Index %d ] Finish to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())

	return nil
}

// saveEpochForTest just for testing w/o server mode
func saveEpochForTest(ctx context.Context, info *BlockchainInfo, db *gorm.DB) error {
	logrus.Infof("[Epoch Index %d ] Start to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())
	tx := db
	defer tx.Rollback()

	//take action to parse table
	blockRewards := &Epoch{
		EpochIndex: int64(info.EpochIndex),
		ThisBlockNumber: info.ThisBlockNum.Int64(),
		LastBlockNumber: info.LastBlockNum.Int64(),
		TotalFees: info.TotalFees.String(),
	}

	if err := tx.Create(blockRewards).Error; err != nil {
		logrus.Errorf("Create epoch error '%v'", err)
		tx.Rollback()
		return processDBErr(err, "Failed to create epoch caused by error %v", err)
	}
	tx.Commit()
	logrus.Infof("[Epoch Index %d ] Finish to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())

	return nil
}


func processDBErr(err error, fmt string, args ...interface{}) error {
	logrus.Errorf(fmt, args...)
	return errors.DatabaseToAPIError(err)
}

//SaveEpochData save the block chain info into database periodically
func (helper *blockHelper)SaveEpochData(ctx context.Context, epochIndex uint64) error {
	info, err := GetEpochFees(helper.ArchNode, epochIndex)
	if err != nil {
		logrus.Errorf("Get epoch info error '%v'", err)
		return errors.BadRequestError(errors.EthCallError, "Get epoch info error")
	}
	//begin to save data into mysql backend
	err1 := saveEpoch(ctx, info)
	if err1 != nil {
		logrus.Errorf("save Epoch data error %v", err1)
		return err1
	}
	return nil
}

type SetStartEPResponse struct {
	EpochIndex		int64		`json:"epoch_index"`
	ThisBlockNum	int64		`json:"this_block_num"`
	LastBlockNum	int64		`json:"last_block_num"`
}

//SetStartEpoch to set the start epoch for sync start point
func SetStartEpoch(ctx context.Context, archNode string, epochIndex uint64) (*SetStartEPResponse,error) {
	helper := &blockHelper{
		ArchNode: archNode,
	}
	err := helper.SaveEpochData(ctx, epochIndex)
	if err != nil {
		logrus.Errorf("Save epoch info error '%v'", err)
		return nil, errors.BadRequestError(errors.EthCallError, "Save epoch info error")
	}

	err = helper.SaveVals(ctx, epochIndex)
	if err != nil {
		logrus.Errorf("Save epoch rewards error '%v'", err)
		return nil, errors.BadRequestError(errors.EthCallError, "Save epoch rewards error")
	}
	//get the record in database to verify
	ep := &Epoch{}
	MDB(ctx).First(&ep, "epoch_index = ?", epochIndex)

	rb := &SetStartEPResponse{
		EpochIndex: ep.EpochIndex,
		ThisBlockNum: ep.ThisBlockNumber,
		LastBlockNum: ep.LastBlockNumber,
	}

	return rb, nil

}


func (helper *blockHelper)SaveEpochDataForTest(ctx context.Context, epochIndex uint64, db *gorm.DB) error {
	info, err := GetEpochFees(helper.ArchNode, epochIndex)
	if err != nil {
		logrus.Errorf("Get epoch info error '%v'", err)
		return errors.BadRequestError(errors.EthCallError, "Get epoch info error")
	}
	//begin to save data into mysql backend
	err = saveEpochForTest(ctx, info, db)
	if err != nil {
		return err
	}
	return nil
}

func (helper *blockHelper)GetStoreEPIndex(ctx context.Context) uint64 {
	ep := &Epoch{}
	MDB(ctx).Order("epoch_index DESC").First(&ep)
	return uint64(ep.EpochIndex)
}

func getFeesInEPStore(ctx context.Context, epIndex uint64) *big.Int {
	ep := &Epoch{}
	MDB(ctx).First(&ep, "epoch_index = ?", epIndex)
	fees, ok := new(big.Int).SetString(ep.TotalFees, 10)
	if ok {
		return fees
	}
	return nil
}

func getFeesInEPStoreForUT(ctx context.Context, epIndex uint64, db *gorm.DB) *big.Int {
	ep := &Epoch{}
	db.First(&ep, "epoch_index = ?", epIndex)
	fees, ok := new(big.Int).SetString(ep.TotalFees, 10)
	if ok {
		return fees
	}
	return nil
}

func getFeesInEPForUT(ctx context.Context, epIndex uint64, db *gorm.DB) *big.Int {
	ep := &Epoch{}
	db.First(&ep, "epoch_index = ?", epIndex)
	fees, ok := new(big.Int).SetString(ep.TotalFees, 10)
	if ok {
		return fees
	}
	return nil
}

func ProcessEpoch(ctx context.Context) (LaIndex uint64, err error) {
	helper := newBlockHelper()
	laInfo, err := helper.ProcessSync(ctx)
	if err != nil {
		logrus.Errorf("There is error with process Epoch %v", err)
		return uint64(0), err
	}
	return laInfo, nil
}

func (helper *blockHelper) ProcessSync(ctx context.Context) (LaIndex uint64, err error) {
	//helper.mu.Lock()
	epstore := helper.GetStoreEPIndex(ctx)
	laInfo := ScramChainInfo(helper.ArchNode)
	//helper.mu.Unlock()

	if laInfo.EpochIndex > epstore {
		logrus.Warningf("Current store EP Index is %d, missing epoch data from %d to %d", epstore, epstore+1, laInfo.EpochIndex)
		epgap := laInfo.EpochIndex - epstore
		logrus.Infof("The epoch gap is %d:", epgap)
		for i := epgap; i > 0; i -- {
			err1 := helper.SaveEpochData(ctx, laInfo.EpochIndex - i + 1)
			if err1 != nil {
				logrus.Errorf("The error for save Epoch data is %v", err1)
				return uint64(0), err1
			}
			//try to save rewards info into database
			err2 := helper.SaveVals(ctx, laInfo.EpochIndex - i + 1)
			if err2 != nil {
				logrus.Errorf("The error for save vals data is %v", err2)
				return uint64(0), err2
			}
		}
	}

	//check the last send record
	sr := &SendRecord{}
	MDB(ctx).Last(&sr).Where("stat = ?", RecordCreated)
	if sr == nil {
		logrus.Debugf("There is no pending send record to update")
		return laInfo.EpochIndex, nil
	}
	logrus.Debugf("The latest created send record is %v with txhash %s", sr, sr.TxHash)
	client, err1 := ethclient.Dial(helper.ArchNode)
	if err1 != nil {
		logrus.Errorf("Eth client dial error %v", err1)
		return laInfo.EpochIndex, nil
	}
	receipt, err2 := client.TransactionReceipt(context.Background(), common.Hash(utils.HexToHash(sr.TxHash)))
	if err2 != nil {
		//no broadcasting failed
		err4 := UpdateSendRecordFailed(ctx, sr.TxHash)
		if err4 != nil{
			logrus.Errorf("update send record error %v", err4)
			return laInfo.EpochIndex, nil
		}
		logrus.Errorf("Get transaction receipt error %v", err2)
		return laInfo.EpochIndex, nil
	}
	if receipt != nil{
		//pending success
		if receipt.Status == uint64(1){
			err3 := UpdateSendRecord(ctx, sr.TxHash)
			if err3 != nil{
				logrus.Errorf("update send record error %v", err3)
				return laInfo.EpochIndex, nil
			}
		} else {
		//pending failed
			err4 := UpdateSendRecordFailed(ctx, sr.TxHash)
			if err4 != nil{
				logrus.Errorf("update send record error %v", err4)
				return laInfo.EpochIndex, nil
			}
		}
	}

	return laInfo.EpochIndex, nil
}

//syncEpoch background
func SyncEpochBackground() {
	var (
		ctx        = context.Background()
	)
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()

	logrus.Infof("Begin to sync Epoch background")
	epIndex, err := ProcessEpoch(ctx)
	if err != nil{
		logrus.Errorf("Failed to sync background with epoch data parsing with error %v in epoch index %d", err, epIndex)
	}
	logrus.Infof("Sync epoch success with latest epoch index %d", epIndex)
}

//process Send background
func ProcessSendBackground() {
	var (
		ctx        = context.TODO()
	)
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	logrus.Infof("Begin to Process Send background")
	err := ProcessSend(ctx)
	if err != nil{
		logrus.Errorf("Failed to process send background with error %v", err)
	}
	logrus.Infof("Process send distribution success!")
}