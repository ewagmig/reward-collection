package utils

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"testing"
)

func TestStrToStrSlice(t *testing.T) {
	s := "39.9075,116.3972"
	r := StrToStrSlice(s)
	t.Log(r)
}

func TestGetConfigYaml(t *testing.T) {
	path := "../conf"
	config := viper.New()

	config.AddConfigPath(path)
	config.SetConfigName("config")
	config.SetConfigType("yaml")

	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	val := config.GetStringMap("channels")
	t.Log(val)

	//chans := models.GetNestMap(val)
	//
	//t.Log(chans)

	//bcinfo, _ := models.GetBCInfo(chans)
	//
	//t.Log(bcinfo)
}

func TestMap(t *testing.T) {
	input := map[string]interface{}{
		"FirstName": "Mitchell",
		"LastName":  "Hashimoto",
		"City":      "San Francisco",
	}

	input_json, _ := json.Marshal(input)
	strjson := string(input_json)
	fmt.Printf("the json str is %s", strjson)
	t.Log(strjson)
}
