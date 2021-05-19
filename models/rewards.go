package models

import (
	"context"
	"github.com/op/go-logging"
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

//todo use this during server mode rather than UT env
//func (rw *Reward) BeforeCreate() error {
//	db := MDB(context.Background()).First(&Reward{}, "epoch_index = ? and validator_addr = ?", rw.EpochIndex, rw.ValidatorAddr)
//	if db.RecordNotFound() {
//		return nil
//	}
//
//	return errors.ConflictErrorf(errors.EPIndexExist, "Epoch Index %d along with Validator %s exists", rw.EpochIndex, rw.ValidatorAddr)
//}
//
//func (ep *Epoch) BeforeCreate() error {
//	db := MDB(context.Background()).First(&Epoch{}, "epoch_index = ?", ep.EpochIndex)
//	if db.RecordNotFound() {
//		return nil
//	}
//
//	return errors.ConflictErrorf(errors.EPIndexExist, "Epoch Index %d exists", ep.EpochIndex)
//}



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

func saveEpoch(ctx context.Context, info *BlockchainInfo) error {
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
	return nil
}


func processDBErr(err error, log *logging.Logger, fmt string, args ...interface{}) error {
	log.Errorf(fmt, args...)
	return errors.DatabaseToAPIError(err)
}