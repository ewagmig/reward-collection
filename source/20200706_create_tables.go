package source

import (
	"time"

	"github.com/op/go-logging"

	"github.com/spf13/cast"
	mdb "github.com/starslabhq/rewards-collection/common/db"
	"github.com/starslabhq/rewards-collection/models"
	"github.com/starslabhq/rewards-collection/server"
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
		&models.SendRecord{},
	}

	err := mdb.Get().AutoMigrate(tables...)
	if err != nil {
		return err
	}

	//err = models.CreateMetadata(context.Background(), "BaasVersion", []string{"1.0.0"})
	//if err != nil {
	//	return err
	//}
	//
	//return models.CreateMetadata(context.Background(), "MSPType", []string{"Standard Fabric-CA", "Internal CA"})
	return nil
}

func (c *createTables20200706) Rollback() error {
	return nil
}
