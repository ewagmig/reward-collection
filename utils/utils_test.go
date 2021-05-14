package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"math/big"
	"strconv"
	"strings"
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

func TestSplitStr(t *testing.T) {
	str0 := "0x000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000394c549ef0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000000"
	str0 = strings.TrimPrefix(str0, "0x")
	//fee address
	str01 := str0[0:64]
	t.Log(str01)

	//val status
	str02 := str0[64:128]
	t.Log(str02)

	//coins
	str03 := str0[128:192]
	t.Log(str03)

	//hbIncoming
	str04 := str0[192:256]
	t.Log(str04)
}

func TestRemoveDuplicateStrings(t *testing.T) {
	arr1 := []string{"a", "b", "c"}
	arr2 := []string{"a", "b"}
	arr := StringArrayDiff(arr1, arr2)
	t.Log(arr)
}

func TestGetMod(t *testing.T) {
	big1 := new(big.Int).SetInt64(int64(3320))
	big2 := new(big.Int).SetInt64(int64(200))

	mod := new(big.Int)
	mod.Mod(big1, big2)

	v := mod.Int64()
	t.Log(v)
}