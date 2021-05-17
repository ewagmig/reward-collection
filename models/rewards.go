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
	ThisBlockNumber     int64      `json:"this_block_number"`
	LastBlockNumber     int64      `json:"last_block_number"`
	Distributed         bool       `json:"distributed"`
	LastTxCreatedAt     *time.Time `json:"last_tx_created_at"`
	AtBase
}

func (Reward) TableName() string {
	return "rewards"
}

type Epoch struct {
	IDBase
	EpochIndex          int64      `json:"epoch_index"`
	Distributed         bool       `json:"distributed"`
	LastTxCreatedAt     *time.Time `json:"last_tx_created_at"`
	Remaining			string	   `json:"remaining"`
	AtBase
}

func (Epoch) TableName() string {
	return "epochs"
}
