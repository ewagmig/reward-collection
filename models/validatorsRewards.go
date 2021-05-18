package models

import (
	"github.com/starslabhq/rewards-collection/errors"
	"github.com/starslabhq/rewards-collection/utils"
	"github.com/ybbus/jsonrpc"
	"math/big"
	"strings"
)

const (
	EP = uint64(200)
)

/*
	the epoch_index with relationship of epoch number and block height during one epoch:
	[epoch_index * EP, (epoch_index + 1) * EP -1), i.e.
	[0,199), [200, 399), [400, 599), .etc
*/

type CallParams struct {
	//ArchiveNode could be fetched from consumer input or the default configuration from yaml file
	ArchiveNode		string	`json:"archive_node,omitempty"`
	PeriodBlocks	uint64	`json:"period_blocks,omitempty"`
	EpochIndex		uint64  `json:"epoch_index,omitempty"`
}

type ValRewardsInfo struct {
	ValAddr		string			`json:"val_addr"`
	Rewards		*big.Int		`json:"rewards"`
}

type ValidatorInfo struct {
	FeeAddr		string
	Status      string
	Coins 		string
	HBIncoming  string
}

func GetRewards(params *CallParams) ([]*ValRewardsInfo, error){
	valsRewardsInfos, err := GetDistributionPerEpoch(params.ArchiveNode, params.EpochIndex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	return valsRewardsInfos, nil
}

//GetDistributionPerEpoch to get distribution per epoch
func GetDistributionPerEpoch(archiveNode string, epochIndex uint64) ([]*ValRewardsInfo, error) {
	totalRewards, err := getDeltaRewards(epochIndex, archiveNode)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	valRewInfo, err := calcuDistInEpoch(epochIndex, totalRewards, archiveNode)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	return valRewInfo, nil
}

func callContract(ContractAddr, archNode string) (string, error)  {
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := ContractAddr

	//todo integrate with contract, fnSig and args
	//fn distribution signature in smart contract
	fnSig := "0x000000"
	args := "000000000000000000000000"
	dataOb := fnSig + args

	resp, err := client.Call("eth_call",map[string]interface{}{
		"to": validatorContractAddr,
		"data": dataOb,
	})
	if err != nil {
		return "",errors.BadRequestError(errors.EthCallError, err)
	}

	result, err := resp.GetString()
	if err != nil {
		return "",errors.BadRequestError(errors.EthCallError, err)
	}
	return result, nil
}

//ContractMonitoring
//todo ContractMonitoring is ready to listen the event from distribution contract
//func ContractMonitoring()  {
//	
//}

//GetRewardsAtBlock return the total rewards at the specific block number
func GetRewardsAtBlock(archiveNode string, blkNum uint64) (*big.Int, error) {
	//get all the validators
	blkhex := utils.EncodeUint64(blkNum)

	//get all active validators
	allVals, err := jsonrpcEthCallGetActVals(archiveNode, blkhex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	//fetch the values from current block number
	sumRewardsAtBlk := &big.Int{}
	for _, val := range allVals {
		valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, blkhex, val)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}
		//remove zero before next operation
		valReward := valInfo.HBIncoming
		valReward = removeConZero(valReward)
		valReward = "0x" + valReward
		rewardsInBig, err := utils.DecodeBig(valReward)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}
		sumRewardsAtBlk = sumRewardsAtBlk.Add(sumRewardsAtBlk, rewardsInBig)
	}
	return sumRewardsAtBlk, nil
}

/*
during the epoch means the block numbers between ecpochIndex * EP, (epochIndex + 1) * EP - 1

 */

//getDeltaRewards return the rewards during the epoch number of blocks
func getDeltaRewards(epochIndex uint64, archiveNode string) (*big.Int, error) {
	epochStartNum := epochIndex * EP
	epochEndNum := (epochIndex + 1) * EP - 1

	rewardsAtStart, err := GetRewardsAtBlock(archiveNode, epochStartNum)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	rewardsAtEnd, err := GetRewardsAtBlock(archiveNode, epochEndNum)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	deltaRewards := &big.Int{}

	deltaRewards.Sub(rewardsAtEnd, rewardsAtStart)

	return deltaRewards, nil
}

//calDistr: 50% per NumberOfActiveVal, 40% per Staking Coins, 10% per stakingOfCoins
func calcuDistInEpoch(epochIndex uint64, rewards *big.Int, archiveNode string) (valsInfo []*ValRewardsInfo, err error) {
	epochEndNum := (epochIndex + 1) * EP -1
	//make distribution of sumRewards
	rewardsPerActNums := new(big.Int)
	rewardsPerActNums.Div(rewards, new(big.Int).SetInt64(int64(2)))
	actValSet, err := jsonrpcEthCallGetActVals(archiveNode, utils.EncodeUint64(epochEndNum))
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	//Here the ActNum should be 11
	ActNum := len(actValSet)
	perActReward := new(big.Int)
	perActReward.Div(rewardsPerActNums, new(big.Int).SetInt64(int64(ActNum)))

	rewardsPerStakingCoins := new(big.Int)
	rewardsDouble :=new(big.Int)
	rewardsDouble.Mul(rewards, new(big.Int).SetInt64(int64(2)))
	rewardsPerStakingCoins.Div(rewardsDouble, new(big.Int).SetInt64(int64(5)))

	//get the total staking
	totalStakeStr, err := jsonrpcEthCallGetTotalStaking(archiveNode, utils.EncodeUint64(epochEndNum))
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	tatoalStakeHex := parseGetAllStaking(totalStakeStr)
	totalCoinsInEpoch, err := utils.DecodeBig(tatoalStakeHex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	//sharePerCoin
	sharePerCoin := new(big.Int)
	sharePerCoin.Div(rewardsPerStakingCoins, totalCoinsInEpoch)

	for _, actVal := range actValSet {
		val, err := jsonrpcEthCallGetValInfo(archiveNode, utils.EncodeUint64(epochEndNum), actVal)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}

		//convert to bigInt
		valCoin := "0x" + val.Coins
		valC, err := utils.DecodeBig(valCoin)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}

		perCoinsReward := new(big.Int)
		perCoinsReward.Mul(sharePerCoin, valC)

		actValRewards := new(big.Int)
		actValRewards.Add(perActReward, perCoinsReward)

		valInfo := &ValRewardsInfo{
			ValAddr: actVal,
			Rewards: actValRewards,
		}
		valsInfo = append(valsInfo, valInfo)
	}


	//get all valAddr with prefix "0x"
	allVals, err := rpcCongressGetAllVals(archiveNode, utils.EncodeUint64(epochEndNum))
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	//remove "0x" prefix in allVals
	var vals []string
	for _, v := range allVals {
		v = strings.TrimPrefix(v, "0x")
		vals = append(vals, v)
	}

	valsbs := utils.StringArrayDiff(vals, actValSet)
	if len(valsbs) == 0 {
		return nil, errors.BadRequestErrorf(errors.EthCallError, "No standby val")
	}
	sbValNums := len(valsbs)

	rewardsPerStandbyNums := new(big.Int)
	rewardsPerStandbyNums.Div(rewards, new(big.Int).SetInt64(int64(10)))

	perSBReward := new(big.Int)
	perSBReward.Div(rewardsPerStandbyNums, new(big.Int).SetInt64(int64(sbValNums)))

	for _, sbv := range valsbs{
		valInfo := &ValRewardsInfo{
			ValAddr: sbv,
			Rewards: perSBReward,
		}
		valsInfo = append(valsInfo, valInfo)
	}

	//remaining is the remaining of rewards - perActNums - perStakingCoins - perStandbyNums
	//todo how to handle the remaining rewards after each epoch calculation?
	//remainingRewards := new(big.Int)
	//remainingRewards.Sub(rewards, rewardsPerActNums)
	//remainingRewards.Sub(remainingRewards, rewardsPerStakingCoins)
	//remainingRewards.Sub(remainingRewards, rewardsPerStandbyNums)


	//mock response
	return valsInfo, nil
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

//todo check whether this abi is avail in smart contract
//jsonrpcEthCallGetTotalStaking to fetch all active validators
func jsonrpcEthCallGetTotalStaking(archNode, blkNumHex string) (string, error) {
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//use the json_rpc api, e.g.{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000f000", "data":"0x8a11d7c9000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5"}, "latest"],"id":1}
	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x000000000000000000000000000000000000f000"
	//fn getActiveValidators signature in smart contract
	getTtoalStakingPrefix := "0xc253c384"

	resp, err := client.Call("eth_call", map[string]interface{}{
		"to": validatorContractAddr,
		"data":getTtoalStakingPrefix,
	}, blkNumHex)
	if err != nil {
		return "nil",errors.BadRequestError(errors.EthCallError, err)
	}

	vals, err := resp.GetString()
	if err != nil {
		return "nil",errors.BadRequestError(errors.EthCallError, err)
	}


	return vals, nil

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
	strlen = removeConZero(strlen)
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

//parseGetAllStaking return the staking value
func parseGetAllStaking(stake string) (string) {
	stake = strings.TrimPrefix(stake, "0x")
	//remove all the concess zeros
	stake = removeConZero(stake[0:64])
	return "0x" + stake
}

//removeConZero
func removeConZero(str string) (string) {
	var index int
	sb := []byte(str)
	for i := 0; i < len(sb); i++ {
		if sb[i] == 48 {
			continue
		} else {
			index = i
			break
		}
	}
	return str[index:]
}