package source

import (
	mdb "github.com/ewagmig/rewards-collection/common/db"
	"github.com/ewagmig/rewards-collection/models"
	"github.com/ewagmig/rewards-collection/server"
	"github.com/op/go-logging"
	"github.com/spf13/cast"
	"time"
)

var migrationLogger = logging.MustGetLogger("baas.migration")

func init() {
	err := server.Register(&createTables20200706{})
	if err != nil {
		migrationLogger.Errorf("createTables20200706 can not be registered")
	}
}

type createTables20200706 struct{}

func (c *createTables20200706) Name() string {
	return "createTables20200706"
}

func (c *createTables20200706) Date() time.Time {
	return cast.ToTime("2017-08-17 15:46:45")
}

func (c *createTables20200706) Apply() error {
	tables := []interface{}{
		&models.Reward{},
		&models.Epoch{},
	}

	err := mdb.Get().AutoMigrate(tables...).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *createTables20200706) Rollback() error {
	return nil
}
