package mysql

import (
	_ "github.com/go-sql-driver/mysql" // inject mysql driver to go sql
	"github.com/jinzhu/gorm"
)

func InitDB(source string) (*gorm.DB, error) {
	gdb, err := gorm.Open("mysql", source)
	gdb.DB().SetMaxIdleConns(0)
	return gdb, err
}
