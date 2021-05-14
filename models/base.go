package models

import (
	"context"
	mdb "github.com/ewagmig/rewards-collection/common/db"
	"github.com/jinzhu/gorm"
	"time"
)



// IDBase contains a ID field which can be used as the base definition for
// other model definitions.
type IDBase struct {
	ID uint `json:"id" gorm:"primary_key"`
}

// AtBase contains fields of time at which may create/update/delete model.
type AtBase struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}



func MDB(ctx context.Context) *gorm.DB {
	db := mdb.Get()
	return db
}
