package mysql

import (
	_ "github.com/go-sql-driver/mysql" // inject mysql driver to go sql
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
