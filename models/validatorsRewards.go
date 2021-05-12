package models

import (
	"github.com/ewagmig/rewards-collection/errors"
	"github.com/ewagmig/rewards-collection/utils"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc"
	"math/big"
	"strings"
)

type CallParams struct {
	//ArchiveNode could be fetched from consumer input or the default configuration from yaml file
	ArchiveNode		string	`json:"archive_node,omitempty"`
	PeriodBlocks	uint64	`json:"period_blocks"`
	CurrBlkNum		uint64  `json:"curr_blk_num"`
}

type ValRewardsInfo struct {
	ValAddr		string		`json:"val_addr"`
	Rewards		big.Int		`json:"rewards"`
}

type ValRewardsInfos []ValRewardsInfo


type ValidatorInfo struct {
	FeeAddr		string
	Status      string
	Coins 		string
	HBIncoming  string
}

func GetState(params *CallParams) (valsInfo *ValRewardsInfos, err error){
	var archiveNode string
	if len(params.ArchiveNode) != 0 {
		archiveNode = params.ArchiveNode
	} else {
		archiveNode = viper.GetString("server.archiveNodeUrl")
	}
	//get all the validators
	blkhex := utils.EncodeUint64(params.CurrBlkNum)

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
	blkp0hex := utils.EncodeUint64(blkP0)

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
func jsonrpcEthCallGetValInfo(archNode, blkNumHex, addrHex string) (*ValidatorInfo, error){
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//use the json_rpc api, e.g.{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000f000", "data":"0x8a11d7c9000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5"}, "latest"],"id":1}
	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x000000000000000000000000000000000000f000"
	//fn getValidatorInfo signature in smart contract
	getValInfoPrefix := "0x8a11d7c9"
	addrPrefix := "000000000000000000000000"
	valAddr := strings.TrimPrefix(addrHex, "0x")
	dataOb := getValInfoPrefix + addrPrefix + valAddr

	resp, err := client.Call("eth_call",map[string]interface{}{
		"to": validatorContractAddr,
		"data": dataOb,
	},blkNumHex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	valInfo, err := resp.GetString()
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	//split the valInfo string into corresponding field
	validatorInfo, err := splitValInfo(valInfo)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	return validatorInfo, nil
}

//jsonrpcEthCallGetActVals to fetch all active validators
func jsonrpcEthCallGetActVals(archNode, blkNumHex string) ([]string, error) {
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//use the json_rpc api, e.g.{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000f000", "data":"0x8a11d7c9000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5"}, "latest"],"id":1}
	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x000000000000000000000000000000000000f000"
	//fn getActiveValidators signature in smart contract
	getValsPrefix := "0x9de70258"

	resp, err := client.Call("eth_call", map[string]interface{}{
		"to": validatorContractAddr,
		"data":getValsPrefix,
	}, blkNumHex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	vals, err := resp.GetString()
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	actVals, err := splitVals(vals)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	return actVals, nil

}



//rpcCongressGetAllVals is used for congress api querying, the ethrpc is not suitable anymore use another json rpc client
//The corresponding archive node should open the rpc "congress" api in addition to other normal apis.
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

//splitValInfo try to split the string into corresponding field according to validators smart contract
func splitValInfo(valInfo string) (*ValidatorInfo, error) {
	if len(valInfo) == 0 {
		return nil, errors.BadRequestErrorf(errors.EthCallError, "The valInfo is nil")
	}
	valInfo = strings.TrimPrefix(valInfo, "0x")
	return &ValidatorInfo{
		FeeAddr: valInfo[:64],
		Status: valInfo[64:128],
		Coins: valInfo[128:192],
		HBIncoming: valInfo[192:256],
	}, nil
}
//splitVals try to split the string into corresponding field val address according to validators smart contract
func splitVals(vals string) ([]string, error) {
	if len(vals) == 0 {
		return nil, errors.BadRequestError(errors.EthCallError, "The vals is nil")
	}

	vals = strings.TrimPrefix(vals, "0x")
	//remove all the zeros in length
	strlen := vals[64:128]
	strlen = strings.Replace(strlen,"0", "", -1)
	length := "0x" + strlen
	nLen, err := utils.DecodeUint64(length)
	if err != nil {
		return nil, errors.BadRequestError(errors.EthCallError, "decode hexstring error")
	}

	if nLen == 0 {
		return nil, errors.BadRequestError(errors.EthCallError, "The length is zero")
	}

	//make an array to hold all the val string elements
	valsArray := []string{}
	for i := uint64(0); i < nLen; i++ {
		valsArray = append(valsArray, vals[64*(3+i)-40:64*(3+i)])
	}
	return valsArray, nil
}