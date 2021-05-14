package pg

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // inject pg driver to go sql
)

func InitDB(source string) (*gorm.DB, error) {
	return gorm.Open("postgres", source)
}
