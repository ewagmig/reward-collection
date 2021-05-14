package server

import (
	"sort"
	"time"

	mdb "github.com/ewagmig/rewards-collection/common/db"
)

// Migrater represents a migration operation.
type Migrater interface {
	// Name returns a string to describe this migration.
	Name() string
	// Date return the time when to apply this migration.
	Date() time.Time
	// Apply defines the logic to apply this migration.
	Apply() error
	// Rollback defines the logic to rollback this migration.
	Rollback() error
}

type migraterList []Migrater

func (m migraterList) Len() int           { return len(m) }
func (m migraterList) Less(i, j int) bool { return m[i].Date().Before(m[j].Date()) }
func (m migraterList) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

var list migraterList

type migration struct {
	ID           int    `gorm:"primary_key"`
	Name         string `gorm:"type:varchar(100)"`
	ExpectedDate time.Time
	CreatedAt    time.Time
}

// Register registers a migrater.
func Register(m Migrater) error {
	list = append(list, m)
	sort.Sort(list)
	return nil
}

// RunMigration apply all migrations.
func RunMigration() error {
	if len(list) == 0 {
		logger.Warning("No migration can be applied")
		return nil
	}

	db := mdb.Get()
	// Create migrations table if not exists.
	db.AutoMigrate(&migration{})

	now := time.Now()
	for _, m := range list {
		if !now.After(m.Date()) {
			break
		}

		if !db.Where("name = ? and expected_date = ?", m.Name(), m.Date()).First(&migration{}).RecordNotFound() {
			continue
		}

		logger.Infof("Start to migrate the %s", m.Name())
		if err := m.Apply(); err != nil {
			logger.Errorf("Failed to migrate %s caused by error %v", m.Name(), err)
			if err := m.Rollback(); err != nil {
				logger.Errorf("Failed to rollback migration %s caused by error %v", m.Name(), err)
			}

			return err
		}

		if err := db.Create(&migration{
			Name:         m.Name(),
			ExpectedDate: m.Date(),
		}).Error; err != nil {
			logger.Errorf("Failed to save migration %s due to error %v", m.Name(), err)
			return err
		}
	}
	return nil
}
