package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	mdb "github.com/starslabhq/rewards-collection/common/db"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func PrintJSON(data interface{}) {
	r, _ := StructToJSON(data)
	fmt.Print(r)
}

func YAMLToStruct(yamlPath string, result interface{}) error {
	content, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, result)
}

func StructToYAML(s interface{}) (string, error) {
	r, err := yaml.Marshal(s)
	return string(r), err
}

func StructToJSON(s interface{}) (string, error) {
	r, err := json.Marshal(s)
	return string(r), err
}

func InitDBConnectionString() error {
	defaultDB := viper.GetString("database.default")
	if defaultDB == "" {
		defaultDB = "mysql"
	}
	connStr := viper.GetString(fmt.Sprintf("database.%s.connection", defaultDB))

	if connStr == "" {
		return fmt.Errorf("Invalid connection string: %s" + connStr)
	}

	var err error
	switch defaultDB {
	case "mysql":
		err = mdb.Init(mdb.MySQL, connStr)
	case "postgres":
		err = mdb.Init(mdb.PostgreSQL, connStr)
	case "sqlite":
		err = mdb.Init(mdb.Sqlite, connStr)
	default:
		return fmt.Errorf("Invalid default database type")
	}
	if err != nil {
		return err
	}

	return nil
}