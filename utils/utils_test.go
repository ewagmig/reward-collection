package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"strconv"
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
	config.SetConfigName("common-backend")
	config.SetConfigType("yaml")

	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	val := config.GetString("server.archiveNodeUrl")
	t.Log(val)

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

func TestConv(t *testing.T) {
	blk := uint64(100)
	blkByte := []byte(strconv.FormatUint(blk, 16))
	blkHex := hex.EncodeToString(blkByte)
	t.Log(blkHex)

	blkB, _ := hex.DecodeString(blkHex)
	resp, _ := strconv.ParseUint(string(blkB), 16, 64)
	t.Log(resp)
}

func TestHexutils(t *testing.T) {
	blk := uint64(100)
	blkhex := EncodeUint64(blk)
	t.Log(blkhex)

	resp, _ := DecodeUint64(blkhex)
	t.Log(resp)
}