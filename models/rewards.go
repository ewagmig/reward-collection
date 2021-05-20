package models

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"github.com/starslabhq/rewards-collection/errors"
	"time"
)

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
}

func newBlockHelper() *blockHelper {
	archNode := viper.GetString("server.archiveNodeUrl")
	if len(archNode) == 0 {
		blockslogger.Errorf("No archNode config!")
		return nil
	}
	return &blockHelper{
		ArchNode: archNode,
	}
}

//todo use this during server mode rather than UT env
func (rw *Reward) BeforeCreate() error {
	db := MDB(context.Background()).First(&Reward{}, "epoch_index = ? and validator_addr = ?", rw.EpochIndex, rw.ValidatorAddr)
	if db.RecordNotFound() {
		return nil
	}

	return errors.ConflictErrorf(errors.EPIndexExist, "Epoch Index %d along with Validator %s exists", rw.EpochIndex, rw.ValidatorAddr)
}

func (ep *Epoch) BeforeCreate() error {
	db := MDB(context.Background()).First(&Epoch{}, "epoch_index = ?", ep.EpochIndex)
	if db.RecordNotFound() {
		return nil
	}

	return errors.ConflictErrorf(errors.EPIndexExist, "Epoch Index %d exists", ep.EpochIndex)
}



func SaveVals(ctx context.Context, valInfos []*ValRewardsInfo) error {
	for _, val := range valInfos {
		if err := saveValReward(ctx, val); err != nil {
			blockslogger.Errorf("Create rewards error '%v'", err)
			continue
		}
	}
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
		blockslogger.Errorf("Create rewards error '%v'", err)
		tx.Rollback()
		return processDBErr(err, blockslogger, "Failed to create rewards caused by error %v", err)
	}
	tx.Commit()
	return nil
}

//saveEpoch for server mode
func saveEpoch(ctx context.Context, info *BlockchainInfo) error {
	blockslogger.Infof("[Epoch Index %d ] Start to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())
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
		blockslogger.Errorf("Create epoch error '%v'", err)
		tx.Rollback()
		return processDBErr(err, blockslogger, "Failed to create epoch caused by error %v", err)
	}
	tx.Commit()
	blockslogger.Infof("[Epoch Index %d ] Finish to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())

	return nil
}

// saveEpochForTest just for testing w/o server mode
func saveEpochForTest(ctx context.Context, info *BlockchainInfo, db *gorm.DB) error {
	blockslogger.Infof("[Epoch Index %d ] Start to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())
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
		blockslogger.Errorf("Create epoch error '%v'", err)
		tx.Rollback()
		return processDBErr(err, blockslogger, "Failed to create epoch caused by error %v", err)
	}
	tx.Commit()
	blockslogger.Infof("[Epoch Index %d ] Finish to store epoch data for with fees %s", info.EpochIndex,info.TotalFees.String())

	return nil
}


func processDBErr(err error, log *logging.Logger, fmt string, args ...interface{}) error {
	log.Errorf(fmt, args...)
	return errors.DatabaseToAPIError(err)
}

//SaveEpochData save the block chain info into database periodically
func (helper *blockHelper)SaveEpochData(ctx context.Context, epochIndex uint64) error {
	info, err := GetEpochFees(helper.ArchNode, epochIndex)
	if err != nil {
		blockslogger.Errorf("Get epoch info error '%v'", err)
		return errors.BadRequestError(errors.EthCallError, "Get epoch info error")
	}
	//begin to save data into mysql backend
	err = saveEpoch(ctx, info)
	if err != nil {
		return err
	}
	return nil
}

func (helper *blockHelper)SaveEpochDataForTest(ctx context.Context, epochIndex uint64, db *gorm.DB) error {
	info, err := GetEpochFees(helper.ArchNode, epochIndex)
	if err != nil {
		blockslogger.Errorf("Get epoch info error '%v'", err)
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

func ProcessEpoch(ctx context.Context) (LaIndex uint64, err error) {
	helper := newBlockHelper()
	epstore := helper.GetStoreEPIndex(ctx)
	laInfo := ScramChainInfo(helper.ArchNode)
	if laInfo.EpochIndex > epstore {
		blockslogger.Warningf("Current store EP Index is %d, missing epoch data from %d to %d", epstore, epstore+1, laInfo.EpochIndex)
		epgap := laInfo.EpochIndex - epstore
		for i := epgap; i > 0; i -- {
			err = helper.SaveEpochData(ctx, laInfo.EpochIndex - i + 1)
			if err != nil {
				return uint64(0), err
			}
		}
	}

	return laInfo.EpochIndex, nil
}

func SyncEpochBackground() {
	var (
		ctx        = context.Background()
	)
	epIndex, err := ProcessEpoch(ctx)
	if err != nil{
		blockslogger.Errorf("Failed to sync background with epoch data parsing with error %v in epoch index %d", err, epIndex)
	}
	blockslogger.Debugf("Sync epoch success with latest epoch index %d", epIndex)
}