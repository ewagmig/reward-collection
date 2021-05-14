package models

import (
	"time"
)

// Reward is a reward fetching per epoch and store the data in table
// [TABLE]
type Reward struct {
	IDBase
	ValidatorAddr       string     `json:"validator_addr"`
	Rewards			    string     `json:"rewards"`
	EpochIndex          int64      `json:"epoch_index"`
	BlockHeight         int64      `json:"block_height"`
	Distributed         bool       `json:"distributed"`
	LastTxCreatedAt     *time.Time `json:"last_tx_created_at"`
	AtBase
}

func (Reward) TableName() string {
	return "rewards"
}

type Epoch struct {
	IDBase
	Distributed         bool       `json:"distributed"`
	LastTxCreatedAt     *time.Time `json:"last_tx_created_at"`
	AtBase
}

func (Epoch) TableName() string {
	return "epochs"
}
