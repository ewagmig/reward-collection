package models

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/starslabhq/rewards-collection/errors"
	"github.com/starslabhq/rewards-collection/utils"
	"github.com/ybbus/jsonrpc"
	"math/big"
	"modernc.org/sortutil"
	"strings"
	//"modernc.org/sortutil"
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
	ArchiveNode		string		`json:"archive_node,omitempty"`
	//PeriodBlocks	uint64		`json:"period_blocks,omitempty"`
	EpochIndex		uint64  	`json:"epoch_index,omitempty"`
	ThisEpoch		uint64		`json:"this_epoch,omitempty"`
	LastEpoch		uint64		`json:"last_epoch,omitempty"`
}

type ValRewardsInfo struct {
	ValAddr		string			`json:"val_addr"`
	Rewards		*big.Int		`json:"rewards"`
	EpochIndex	uint64  		`json:"epoch_index"`
}

type ValidatorInfo struct {
	FeeAddr		string
	Status      string
	Coins 		string
	HBIncoming  string
}

func GetRewards(params *CallParams) (*big.Int, error){
	RewardsInfos, err := FetchTotalRewardsEPs(context.TODO(), params.ThisEpoch, params.LastEpoch)
	//valsRewardsInfos, err := GetDistributionPerEpoch(params.ArchiveNode, params.EpochIndex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	return RewardsInfos, nil

}


//GetDistributionPerEpoch to get distribution per epoch
func GetDistributionPerEpoch(archiveNode string, epochIndex uint64) ([]*ValRewardsInfo, error) {
	//use the block scraper method to get block fees during one epoch
	totalRewards := GetBlockEpochRewards(archiveNode, epochIndex)
	//totalRewards := big.NewInt(10000000000000)
	valRewInfo, err := calcuDistInEpoch(epochIndex, totalRewards, archiveNode)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	return valRewInfo, nil
}

//GetRewardsAtBlock return the total rewards at the specific block number
//func GetRewardsAtBlock(archiveNode string, blkNum uint64) (*big.Int, error) {
//	//get all the validators
//	blkhex := utils.EncodeUint64(blkNum)
//
//	//get all active validators
//	allVals, err := jsonrpcEthCallGetActVals(archiveNode, blkhex)
//	if err != nil {
//		return nil,errors.BadRequestError(errors.EthCallError, err)
//	}
//
//	//fetch the values from current block number
//	sumRewardsAtBlk := &big.Int{}
//	for _, val := range allVals {
//		valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, blkhex, val)
//		if err != nil {
//			return nil,errors.BadRequestError(errors.EthCallError, err)
//		}
//		//remove zero before next operation
//		valReward := valInfo.HBIncoming
//		valReward = removeConZero(valReward)
//		valReward = "0x" + valReward
//		rewardsInBig, err := utils.DecodeBig(valReward)
//		if err != nil {
//			return nil,errors.BadRequestError(errors.EthCallError, err)
//		}
//		sumRewardsAtBlk = sumRewardsAtBlk.Add(sumRewardsAtBlk, rewardsInBig)
//	}
//	return sumRewardsAtBlk, nil
//}

/*
during the epoch means the block numbers between ecpochIndex * EP, (epochIndex + 1) * EP - 1

 */

//getDeltaRewards return the rewards during the epoch number of blocks
//func getDeltaRewards(epochIndex uint64, archiveNode string) (*big.Int, error) {
//	epochStartNum := epochIndex * EP
//	epochEndNum := (epochIndex + 1) * EP - 1
//
//	rewardsAtStart, err := GetRewardsAtBlock(archiveNode, epochStartNum)
//	if err != nil {
//		return nil,errors.BadRequestError(errors.EthCallError, err)
//	}
//
//	rewardsAtEnd, err := GetRewardsAtBlock(archiveNode, epochEndNum)
//	if err != nil {
//		return nil,errors.BadRequestError(errors.EthCallError, err)
//	}
//
//	deltaRewards := &big.Int{}
//
//	deltaRewards.Sub(rewardsAtEnd, rewardsAtStart)
//
//	return deltaRewards, nil
//}

//calDistr: 50% per NumberOfActiveVal, 40% per Staking Coins, 10% per stakingOfCoins
func calcuDistInEpoch(epochIndex uint64, rewards *big.Int, archiveNode string) (valsInfo []*ValRewardsInfo, err error) {
	epochEndNum := (epochIndex + 1) * EP -1
	//vals is the pool Length, fetch all the pool info with number iteration
	epochEndNumHex := hexutil.EncodeUint64(epochEndNum)
	//make distribution of sumRewards
	rewardsPerActNums := new(big.Int)
	rewardsPerActNums.Div(rewards, new(big.Int).SetInt64(int64(2)))
	valnum, err := jsonrpcEthCallGetActVals(archiveNode, epochEndNumHex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}

	valMapCoins := make(map[string]*big.Int)
	for i := uint64(0); i < valnum; i ++ {
		valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, epochEndNumHex, i)
		if valInfo.Status == fmt.Sprintf("%064s", "0") {
			continue
		}
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}
		coinsBig := new(big.Int)
		valInfo.Coins = removeConZero(valInfo.Coins)
		if valInfo.Coins == fmt.Sprintf("%064s", "0") {
			coinsBig = big.NewInt(0)
		} else {
			coinsBig, err = hexutil.DecodeBig("0x"+valInfo.Coins)
			if err != nil {
				return nil,errors.BadRequestError(errors.EthCallError, err)
			}
		}

		valMapCoins[valInfo.FeeAddr] = coinsBig
	}
	var bigSort sortutil.BigIntSlice
	for _, v := range valMapCoins {
		bigSort = append(bigSort, v)
	}
	//sort the big numbers ASC
	bigSort.Sort()
	//if len(bigSort) <= 11 {
	//	return nil, errors.BadRequestErrorf(errors.EthCallError, "not enough validators in the slice!")
	//}

	//todo check with the PM on all active nodes allocation
	//act nodes 11 + 10(own nodes)
	ActCoinsArray := bigSort[len(bigSort)-6:]
	totalActCoins := sum(ActCoinsArray)
	fmt.Println(totalActCoins.String())
	//Here the ActNum should be 21
	ActNum := 21
	perActReward := new(big.Int)
	perActReward.Div(rewardsPerActNums, new(big.Int).SetInt64(int64(ActNum)))

	rewardsPerStakingCoins := new(big.Int)
	rewardsDouble :=new(big.Int)
	rewardsDouble.Mul(rewards, new(big.Int).SetInt64(int64(2)))
	rewardsPerStakingCoins.Div(rewardsDouble, new(big.Int).SetInt64(int64(5)))

	totalCoinsInEpoch := totalActCoins
	//sharePerCoin
	sharePerCoin := new(big.Int)
	sharePerCoin.Div(rewardsPerStakingCoins, totalCoinsInEpoch)

	//fetch standby vals
	vals := []string{}
	for k := range valMapCoins{
		vals = append(vals, k)
	}

	//actValSet to fetch the active val set
	actValSet := []string{}
	CoinsMapActAddr := make(map[*big.Int]string)
	for _, cv := range ActCoinsArray {
		for _, v := range vals {
			if valMapCoins[v] == cv {
				CoinsMapActAddr[cv] = v
			}
		}
	}

	for _, v := range CoinsMapActAddr{
		actValSet = append(actValSet, v)
	}


	//actValSet should be aligned to the len of it
	for _, actVal := range actValSet {
		perCoinsReward := new(big.Int)
		perCoinsReward.Mul(sharePerCoin, valMapCoins[actVal])

		actValRewards := new(big.Int)
		actValRewards.Add(perActReward, perCoinsReward)

		valInfo := &ValRewardsInfo{
			ValAddr: actVal,
			Rewards: actValRewards,
			EpochIndex: epochIndex,
		}
		valsInfo = append(valsInfo, valInfo)
	}


	valsbs := utils.StringArrayDiff(vals, actValSet)
	if len(valsbs) == 0 {
		return nil, errors.BadRequestErrorf(errors.EthCallError, "No standby val")
	}
	//sbValNums := len(valsbs)
	rewardsPerStandbyCoins := new(big.Int)
	rewardsPerStandbyCoins.Div(rewards, new(big.Int).SetInt64(int64(10)))

	//fetch all the rewards perStaking
	//standby nodes 11
	SBCoinsArray := bigSort[:5]
	totalSBCoins := sum(SBCoinsArray)

	//select the standby nodes
	CoinsMapSBAddr := make(map[*big.Int]string)
	for _, cv := range SBCoinsArray {
		for _, v := range valsbs {
			if valMapCoins[v] == cv {
				CoinsMapSBAddr[cv] = v
			}
		}
	}

	var sbValSet []string
	for _, v := range CoinsMapSBAddr{
		sbValSet = append(sbValSet, v)
	}

	sharePerSBCoin := new(big.Int)
	//check if all the sb coins equals zero
	if totalSBCoins.CmpAbs(big.NewInt(0)) == 0 {
		sharePerSBCoin = big.NewInt(0)
	} else {
		sharePerSBCoin.Div(rewardsPerStandbyCoins, totalSBCoins)
	}



	for _, sbv := range sbValSet{
		valInfo := &ValRewardsInfo{
			ValAddr: sbv,
			Rewards: new(big.Int).Mul(sharePerSBCoin, valMapCoins[sbv]),
			EpochIndex: epochIndex,
		}
		valsInfo = append(valsInfo, valInfo)
	}

	//remaining is the remaining of rewards - perActNums - perStakingCoins - perStandbyNums
	//remainingRewards := new(big.Int)
	//remainingRewards.Sub(rewards, rewardsPerActNums)
	//remainingRewards.Sub(remainingRewards, rewardsPerStakingCoins)
	//remainingRewards.Sub(remainingRewards, rewardsPerStandbyNums)


	//mock response
	return valsInfo, nil
}

func mockCalcDisInEpoch(epochIndex uint64, rewards *big.Int) (valsInfo []*ValRewardsInfo, err error) {
	valReMap := make(map[string]*big.Int)
	valReMap["0x1"] = big.NewInt(100)
	valReMap["0x2"] = big.NewInt(200)
	valReMap["0x3"] = big.NewInt(300)
	valReMap["0x4"] = big.NewInt(400)
	valReMap["0x5"] = big.NewInt(500)

	for i:= range valReMap {
		valsInfo = append(valsInfo, &ValRewardsInfo{
				ValAddr: i,
				Rewards: valReMap[i],
				EpochIndex: epochIndex,
		})
	}

	return valsInfo, nil
}


//todo: check with the node_voting phase III contract api
//jsonrpcEthCallGetValInfo used to eth_call validator info, the contract Addr is the proxy contract with abi getPoolWithStatus
func jsonrpcEthCallGetValInfo(archNode, blkNumHex string, poolId uint64) (*ValidatorInfo, error){
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//use the json_rpc api, e.g.{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"0x000000000000000000000000000000000000f000", "data":"0x8a11d7c9000000000000000000000000086119bd018ed4940e7427b9373c014f7b754ad5"}, "latest"],"id":1}
	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x7Ce9A4f22FB3B3e2d91cC895bb082d7BD6F08525"
	//fn getPoolWithStatus signature in smart contract
	getValInfoPrefix := "0x22fe6c24"

	//use the poolId as input
	hexutil.EncodeUint64(poolId)
	pid := strings.TrimPrefix(hexutil.EncodeUint64(poolId), "0x")
	pidpad := fmt.Sprintf("%064s", pid)
	dataOb := getValInfoPrefix + pidpad

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

//jsonrpcEthCallGetActVals to fetch pool length with all vals use abi getPoolLength()
func jsonrpcEthCallGetActVals(archNode, blkNumHex string) (uint64, error) {
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x7Ce9A4f22FB3B3e2d91cC895bb082d7BD6F08525"
	//fn getPoolLength() signature in smart contract
	getValsPrefix := "0xb3944d52"

	resp, err := client.Call("eth_call", map[string]interface{}{
		"to": validatorContractAddr,
		"data":getValsPrefix,
	}, blkNumHex)
	if err != nil {
		return 0,errors.BadRequestError(errors.EthCallError, err)
	}

	vals, err := resp.GetString()
	if err != nil {
		return 0,errors.BadRequestError(errors.EthCallError, err)
	}

	vals_num, err := splitVals(vals)
	if err != nil {
		return 0,errors.BadRequestError(errors.EthCallError, err)
	}

	return vals_num, nil

}

//jsonrpcEthCallNotifyAmount call notifyRewardAmount(address[] calldata _validators, uint256[] calldata _rewardAmounts)
func jsonrpcEthCallNotifyAmount(archNode string, valMapDist map[string]*big.Int) (error) {
	//init a new json rpc client
	client := jsonrpc.NewClient(archNode)

	//to assemble the data string structure with fn prefix, addr with left padding
	validatorContractAddr := "0x5CaeF96c490b5c357847214395Ca384dC3d3b85e"
	//fn notifyRewardAmount() signature in smart contract
	notifyRewardAmountPrefix := "0xf141d389"

	sliceLength := len(valMapDist)
	lengthHex := hexutil.EncodeUint64(uint64(sliceLength))
	lengthHex = strings.TrimPrefix(lengthHex, "0x")
	lengthPad := fmt.Sprintf("%064s", lengthHex)
	//to assemble the original data
	addrPrefix := "000000000000000000000000"
	var valaddrs string
	for k := range valMapDist {
		valkey := addrPrefix + k
		valaddrs = valaddrs + valkey
	}
	//address[] calldata
	addrCalldata := lengthPad + valaddrs

	var valValues string
	for _, v := range valMapDist {
		dist := hexutil.EncodeBig(v)
		dist = strings.TrimPrefix(dist, "0x")
		distpad := fmt.Sprintf("%064s", dist)
		valValues  = valValues + distpad
	}
	distcalldata := lengthPad + valValues

	dataOb := valaddrs + addrCalldata + distcalldata

	resp, err := client.Call("eth_call", map[string]interface{}{
		"to": validatorContractAddr,
		"data":notifyRewardAmountPrefix + dataOb,
	})

	if err != nil || resp.Error != nil{
		distributionlogger.Errorf("call notifyReward contract error %v", err)
		return err
	}

	return nil

}

//NotifyAmount
func getNotifyAmountData(valMapDist map[string]*big.Int) string {
	//to assemble the data string structure with fn prefix, addr with left padding
	//validatorContractAddr := "0x5CaeF96c490b5c357847214395Ca384dC3d3b85e"
	//fn notifyRewardAmount() signature in smart contract
	notifyRewardAmountPrefix := "0xf141d389"

	sliceLength := len(valMapDist)
	lengthHex := hexutil.EncodeUint64(uint64(sliceLength))
	lengthHex = strings.TrimPrefix(lengthHex, "0x")
	lengthPad := fmt.Sprintf("%064s", lengthHex)
	//to assemble the original data
	addrPrefix := "000000000000000000000000"
	var valaddrs string
	for k := range valMapDist {
		valkey := addrPrefix + k
		valaddrs = valaddrs + valkey
	}
	//address[] calldata
	addrCalldata := lengthPad + valaddrs

	var valValues string
	for _, v := range valMapDist {
		dist := hexutil.EncodeBig(v)
		dist = strings.TrimPrefix(dist, "0x")
		distpad := fmt.Sprintf("%064s", dist)
		valValues  = valValues + distpad
	}
	distcalldata := lengthPad + valValues

	dataOb := valaddrs + addrCalldata + distcalldata

	dataStr := notifyRewardAmountPrefix + dataOb
	return dataStr
}


//The corresponding archive node should open the rpc "congress" api in addition to other normal apis.
func rpcCongressGetAllVals(epochIndex uint64, archiveNode string) ([]string, error) {
	epochEndNum := (epochIndex + 1) * EP -1
	epochEndNumHex := hexutil.EncodeUint64(epochEndNum)

	valnum, err := jsonrpcEthCallGetActVals(archiveNode, epochEndNumHex)
	if err != nil {
		return nil,errors.BadRequestError(errors.EthCallError, err)
	}
	vals := []string{}
	//valnum is the pool Length, fetch all the pool info with number iteration

	for i := uint64(0); i < valnum; i ++ {
		valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, epochEndNumHex, i)
		if err != nil {
			return nil,errors.BadRequestError(errors.EthCallError, err)
		}
		vals = append(vals, valInfo.FeeAddr)
	}

	return vals, nil

}

//splitValInfo try to split the string into corresponding field according to voting smart contract
func splitValInfo(valInfo string) (*ValidatorInfo, error) {
	if len(valInfo) == 0 {
		return nil, errors.BadRequestErrorf(errors.EthCallError, "The valInfo is nil")
	}
	valInfo = strings.TrimPrefix(valInfo, "0x")
	return &ValidatorInfo{
		FeeAddr: valInfo[:64],
		Status: valInfo[704:768],
		Coins: valInfo[384:448],
		//HBIncoming: valInfo[192:256],
	}, nil
}
//splitVals try to split the string into corresponding field val address according to validators smart contract
func splitVals(vals string) (uint64, error) {
	if len(vals) == 0 {
		return 0, errors.BadRequestError(errors.EthCallError, "The vals is nil")
	}

	vals = strings.TrimPrefix(vals, "0x")
	//remove all the zeros in length
	strlen := vals[:64]
	strlen = removeConZero(strlen)
	length := "0x" + strlen
	nLen, err := utils.DecodeUint64(length)
	if err != nil {
		return 0, errors.BadRequestError(errors.EthCallError, "decode hexstring error")
	}

	if nLen == 0 {
		return 0, errors.BadRequestError(errors.EthCallError, "The length is zero")
	}

	return nLen, nil
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