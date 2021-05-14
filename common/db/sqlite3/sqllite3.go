package sqlite3

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3" // inject sqlite driver to go sql
)

func InitDB(source string) (*gorm.DB, error) {
	if len(source) == 0 {
		source = "console.db"
	}
	gdb, err := gorm.Open("sqlite3", source)
	gdb.DB().SetMaxOpenConns(1)
	return gdb, err
}
