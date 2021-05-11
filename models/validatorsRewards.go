package models

import (
	"encoding/hex"
	"github.com/ewagmig/rewards-collection/errors"
	"github.com/onrik/ethrpc"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc"
	"math/big"
	"strconv"
	"strings"
)

type CallParams struct {
	//ArchiveNode could be fetched from consumer input or the default configuration from yaml file
	ArchiveNode		string	`json:"archive_node,omitempty"`
	PeriodBlocks	uint64	`json:"period_blocks"`
	CurrBlkNum		uint64  `json:"curr_blk_num"`
}

type ValidatorInfo struct {
	ValAddr		[20]byte	`json:"val_addr"`
	Rewards		big.Int		`json:"rewards"`
}

type ValidatorsInfo []ValidatorInfo

func GetState(params *CallParams) (valsInfo *ValidatorsInfo, err error){
	var archiveNode string
	if len(params.ArchiveNode) != 0 {
		archiveNode = params.ArchiveNode
	} else {
		archiveNode = viper.GetString("server.archiveNodeUrl")
	}
	//get all the validators
	blkbyte := []byte(strconv.FormatUint(params.CurrBlkNum, 16))
	blkhex := hex.EncodeToString(blkbyte)

	allVals, err := rpcCongressGetAllVals(archiveNode, blkhex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	//fetch the current block height values
	for _, val := range allVals {
		_, err := jsonrpcEthCallGetValInfo(archiveNode, blkhex, val)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}

	}


	//fetch the period first block height values
	epoch := params.PeriodBlocks
	blkP0 := params.CurrBlkNum - epoch + 1
	blkp0byte := []byte(strconv.FormatUint(blkP0, 16))
	blkp0hex := hex.EncodeToString(blkp0byte)

	//fetch the previous block height values during period
	for _, val := range allVals {
		_, err := jsonrpcEthCallGetValInfo(archiveNode, blkp0hex, val)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}

	}


	return nil, nil

}


//jsonrpcEthCallGetValInfo used to eth_call validator info
func jsonrpcEthCallGetValInfo(archNode, blkNumHex, addrHex string) (string, error){
	//init a new json rpc client
	client := ethrpc.New(archNode)

	//use the json_rpc api, e.g.{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000f000", "data":"0x8a11d7c9000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5"}, "latest"],"id":1}
	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x000000000000000000000000000000000000f000"
	//fn getValidatorInfo signature in smart contract
	getValInfoPrefix := "0x8a11d7c9"
	addrPrefix := "000000000000000000000000"
	valAddr := strings.TrimPrefix(addrHex, "0x")
	dataOb := getValInfoPrefix + addrPrefix + valAddr
	params := ethrpc.T{
		To: validatorContractAddr,
		Data: dataOb,
	}

	resp, err := client.EthCall(params,blkNumHex)
	if err != nil {
		return "",errors.BadRequestError(errors.EthCallError, err)
	}
	return resp, nil
}


//jsonrpcEthCallGetActVals is just used for get active validators, for all validators including active and standby ones, congress.getValidators() should be used
func jsonrpcEthCallGetActVals(archNode, blkNumHex string) (string, error) {
	//init a new json rpc client
	client := ethrpc.New(archNode)

	//use the json_rpc api, e.g.{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000f000", "data":"0x8a11d7c9000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5"}, "latest"],"id":1}
	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x000000000000000000000000000000000000f000"
	//fn getActiveValidators signature in smart contract
	getValsPrefix := "0x8a11d7c9"
	dataOb := getValsPrefix
	params := ethrpc.T{
		To: validatorContractAddr,
		Data: dataOb,
	}

	resp, err := client.EthCall(params,blkNumHex)
	if err != nil {
		return "",errors.BadRequestError(errors.EthCallError, err)
	}

	//try to split the string to []string with padding

	return resp, nil

}

//rpcCongressGetAllVals is used for congress api querying, the ethrpc is not suitable anymore use another json rpc client
func rpcCongressGetAllVals(archNode, blkNumHex string) ([]string, error) {
	rpcClient := jsonrpc.NewClient(archNode)
	resp, err := rpcClient.Call("congress_getValidators", blkNumHex)
	if err != nil {
		return nil,errors.BadRequestError(errors.CongressGetValsError, err)
	}
	if resp.Error != nil {
		return nil,errors.BadRequestError(errors.CongressGetValsError, err)
	}

	//make unmarshalling with the response
	strvals := []string{}
	err = resp.GetObject(&strvals)
	if err != nil {
		return nil,errors.BadRequestError(errors.CongressGetValsError, err)
	}

	return strvals, nil
}