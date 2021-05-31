package server

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"time"
)

const (
	connStr = "root:12345678@tcp(huobichain-dev-02.sinnet.huobiidc.com:3306)/heco_test?charset=utf8&parseTime=True&loc=Local"
)

func InitDB(source string) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.New(mysql.Config{DSN:source}), &gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		return nil, err
	}
	sql,err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sql.SetMaxIdleConns(0)
	return gdb, err
}

func TestMigration(t *testing.T) {
	db, err := InitDB(connStr)
	if err != nil {
		t.Error(err)
	}

	db.AutoMigrate(&migration{})

	now := time.Now()
	for _, m := range list {
		if !now.After(m.Date()) {
			break
		}
		db.Where("name = ? and expected_date = ?", m.Name(), m.Date()).First(&migration{})
		if db.Error != nil && db.Error.Error() != "record not found" {
			continue
		}
		//if !db.Where("name = ? and expected_date = ?", m.Name(), m.Date()).First(&migration{}).RecordNotFound() {
		//	continue
		//}

		logger.Infof("Start to migrate the %s", m.Name())
		if err := m.Apply(); err != nil {
			logger.Errorf("Failed to migrate %s caused by error %v", m.Name(), err)
			if err := m.Rollback(); err != nil {
				logger.Errorf("Failed to rollback migration %s caused by error %v", m.Name(), err)
			}

			t.Error(err)
		}

		if err := db.Create(&migration{
			Name:         m.Name(),
			ExpectedDate: m.Date(),
		}).Error; err != nil {
			logger.Errorf("Failed to save migration %s due to error %v", m.Name(), err)
			t.Error(err)
		}
	}

}
