package db

import (
	"github.com/ewagmig/rewards-collection/common/db/mysql"
	"github.com/ewagmig/rewards-collection/common/db/pg"
	"github.com/ewagmig/rewards-collection/common/db/sqlite3"
	"github.com/ewagmig/rewards-collection/errors"
	"github.com/ewagmig/rewards-collection/utils"
	"github.com/jinzhu/gorm"
	"sync"
)

var (
	gdb      *gorm.DB
	initOnce sync.Once
)

type DBType uint

const (
	MySQL DBType = iota
	PostgreSQL
	Sqlite
)

var skipBlockchainVerifyTables = []string{
	"user_to_blockchain_permissions",
}

// Init inits the database connection only once
func Init(dbType DBType, source string) error {
	var err error
	initOnce.Do(func() {
		switch dbType {
		case PostgreSQL:
			gdb, err = pg.InitDB(source)
		case Sqlite:
			gdb, err = sqlite3.InitDB(source)
		default:
			gdb, err = mysql.InitDB(source) // MySQL is default
		}
	})

	gdb.BlockGlobalUpdate(true)
	// add callbacks to check if user has permissions to access blockchain
	gdb.Callback().Query().Before("gorm:query").Register("baas:check_blockchains_query", queryCallback)
	gdb.Callback().RowQuery().Before("gorm:row_query").Register("baas:check_blockchains_row_query", queryCallback)
	gdb.Callback().Create().Before("gorm:create").Register("baas:check_blockchains_create", createCallback)
	// gdb.Callback().Delete().Before("gorm:delete").Register("concord:check_blockchains_delete", deleteCallback)
	// gdb.Callback().Create().Before("gorm:update").Register("concord:check_blockchains_update", updateCallback)
	return err

}

// Get gets the gorm.DB instance which is safe for concurrent use by multiple goroutines.
func Get() *gorm.DB {
	if gdb == nil {
		panic("db is nil")
	}

	return gdb
}

func createCallback(scope *gorm.Scope) {
	if scope.HasError() {
		return
	}

	blockchainIDs, ok := getBockchainIDs(scope)
	if !ok {
		return
	}

	err := checkBlockchainIDField(scope, blockchainIDs)
	if err != nil {
		scope.Err(err)
	}
}

func queryCallback(scope *gorm.Scope) {
	if scope.HasError() {
		return
	}

	blockchainIDs, ok := getBockchainIDs(scope)
	if !ok {
		return
	}

	addBlockchainCondition(scope, blockchainIDs)
}

func checkBlockchainIDField(scope *gorm.Scope, blockchainIDs []uint) error {
	if utils.StrInSlice(skipBlockchainVerifyTables, scope.TableName()) {
		return nil
	}

	field, ok := scope.FieldByName("BlockchainID")
	if !ok {
		return nil
	}

	bid, ok := field.Field.Interface().(uint)
	if !ok {
		return nil
	}

	if bid == 0 {
		return nil
	}

	if !utils.UintInArray(bid, blockchainIDs) {
		return errors.ForbiddenErrorf(errors.Forbidden, "you have no permission to access this blockchain(ID:%d)", bid)
	}

	return nil
}

func addBlockchainCondition(scope *gorm.Scope, blockchainIDs []uint) {
	switch {
	case scope.TableName() == "blockchains":
		scope.Search.Where("id in (?)", blockchainIDs)
	case scope.HasColumn("blockchain_id"):
		scope.Search.Where("blockchain_id in (?)", blockchainIDs)
	default:
	}
}

func getBockchainIDs(scope *gorm.Scope) ([]uint, bool) {
	blockchainIDsVal, ok := scope.DB().Get(utils.BAAS_DB_BLOCKCHAINS_KEY)
	if !ok {
		return nil, false
	}

	blockchainIDs, ok := blockchainIDsVal.([]uint)
	if !ok {
		return nil, false
	}

	if len(blockchainIDs) == 0 {
		blockchainIDs = append(blockchainIDs, 0)
	}

	return blockchainIDs, true
}
