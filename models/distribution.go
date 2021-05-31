package models

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"github.com/starslabhq/rewards-collection/errors"
	"github.com/starslabhq/rewards-collection/utils"
	"gorm.io/gorm"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	distributionlogger = logging.MustGetLogger("rewards.distribution.models")
	EPDuration = int64(3)
	//sysAddr should be provided by gateway service side
	sysAddr = "0xe2cdcf16d70084ac2a9ce3323c5ad3fa44cddbda"
	//default40GWei = int64(40000000000)

	//todo integration with validator
	validatorUrl = "http://huobichain-dev-02.sinnet.huobiidc.com:5005/validate/cross/check"
	validatorAccessKey = Key{
		AccessKey: AccessKey,
		SecretKey: SecretKey,
	}
	//todo archNode candidates connection before online
	archNodes = []string{
		"47.243.52.187:8545",
		"47.242.228.39:8545",
	}
)

type ValDist struct {
	ValAddr			string
	Distribution 	*big.Int
	ThisEpoch		int64
	LastEpoch		int64
}

type ValMapRewards struct {
	ValAddr			string
	Rewards 		*big.Int
}

type sendHelper struct {
	ArchNode string
	EpochEnd int64
	RawTx    string
	TxHash   string
	valMap   map[string]*big.Int
	mu    sync.RWMutex
}

type DecParams struct {
	Tasks      []Task		`json:"tasks"`
	TxType     string		`json:"tx_type"`
	RawTx      string		`json:"raw_tx"`
}

func newSendHelper() *sendHelper {
	archNode := viper.GetString("server.archiveNodeUrl")
	if len(archNode) == 0 {
		blockslogger.Errorf("No archNode config!")
		return nil
	}

	return &sendHelper{
		ArchNode: archNode,
		//GasPrice: default40GWei,
	}
}

//todo kibana log

//ProcessSend is the entrypoint of send module
func ProcessSend(ctx context.Context) error{
	helper := newSendHelper()
	err := helper.DoSend(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (helper *sendHelper) DoSend(ctx context.Context) error {
	//fetch the latest epoch info
	laInfo := ScramChainInfo(helper.ArchNode)
	epEnd := laInfo.EpochIndex
	epStart := epEnd - uint64(EPDuration) + 1

	//1. begin pre send process
	preSendBool, err := helper.PreSend(ctx, epStart, epEnd, helper.ArchNode)
	if preSendBool && (err != nil) {
		if len(helper.RawTx) > 0 && len(helper.TxHash) > 0 {
			sendBool, err2 := helper.SendDistribution(helper.RawTx, helper.TxHash, helper.ArchNode)
			//send check success
			if sendBool && (err2 !=nil){
				var vals []*ValDist
				for v := range helper.valMap{
					val := &ValDist{
						ValAddr: v,
						ThisEpoch: int64(epStart),
						LastEpoch: int64(epEnd),
					}
					vals = append(vals, val)
				}
				sr := &SendRecord{
					RawTx: helper.RawTx,
				}
				//update the database when successful
				err3 := PostSend(ctx, vals, sr)
				if err3 != nil {
					return err3
				}
			}
		}
	}
	return err
}


//fetchRawTx
func (helper *sendHelper)fetchRawTx(ctx context.Context, epStart, epEnd uint64, archiveNode string) (string, string, error) {
	valmap, err := PumpDistInfo(ctx, epStart, epEnd, helper.ArchNode)
	if err != nil {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return "", "", err
	}
	if len(valmap) == 0 {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return "", "", err
	}

	//get the gateway encrypted data
	encData, err := signGateway(archiveNode, sysAddr, valmap)
	if err != nil {
		distributionlogger.Errorf("Fetch enc data from gateway service error %v", err)
		return "", "", err
	}

	validaReq := ValidatorReq{
		EncryptData: encData.Data.EncryptData,
		Cipher: encData.Data.Extra.Cipher,
	}

	rawTx, _ := ValidateEnc(validaReq, validatorUrl, validatorAccessKey)

	if len(rawTx) == 0 {
		return "", "", errors.BadRequestErrorf(errors.EthCallError, "The rawTx is empty")
	}
	return rawTx, encData.Data.Extra.TxHash, nil
}

//PreSend to pump distribution from database, then take some check before sending
func (helper *sendHelper)PreSend(ctx context.Context, epStart, epEnd uint64, archiveNode string) (bool, error){
	valmap, err := PumpDistInfo(ctx, epStart, epEnd, archiveNode)
	if err != nil {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return false, err
	}
	if len(valmap) == 0 {
		distributionlogger.Errorf("Fetch validator distribution error %v", err)
		return false, err
	}

	rawTx, txHash, err := helper.fetchRawTx(ctx, epStart, epEnd, archiveNode)
	if err != nil {
		return false, err
	}

	//fetch the pending nonce for sending transaction
	nonce, err := fetchPendingNonce(archiveNode, sysAddr)
	if err != nil {
		return false, err
	}

	sr := &SendRecord{
		RawTx: rawTx,
		Nonce: int64(nonce),
		ThisEpoch: int64(epStart),
		LastEpoch: int64(epEnd),
		//GasPrice: default40GWei,
		TxHash: txHash,
		Stat: RecordCreated,
	}

	//save the send record
	distributionlogger.Infof("Beigin to save the send record")
	err = SaveSendRecord(context.TODO(), sr)
	if err != nil {
		return false, err
	}

	//update the field in sender helper
	helper.RawTx = rawTx
	helper.TxHash = txHash
	helper.valMap = valmap


	distributionlogger.Infof("Prepare to send from epStart %d and epEnd %d with result %v", epStart, epEnd, valmap)
	return true, nil
}

type ValidatorResp struct {
	Data 	  DecParams		`json:"data"`
	//RawTx     string       `json:"raw_tx"`
	OK 		  bool		   `json:"ok"`
}

type ValidatorReq struct {
	EncryptData  string		`json:"encrypt_data"`
	Cipher		 string     `json:"cipher"`
}

func ValidateEnc(encData ValidatorReq, targetUrl string, accessKey Key) (rawTx string, ok bool) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

	payloadBytes, err := json.Marshal(&encData)
	if err != nil {
		return
	}
	body := bytes.NewReader(payloadBytes)
	//set the request header according to aws v4 signature
	req1, err := http.NewRequest("POST", targetUrl, body)
	req1.Header.Set("content-type", "application/json")
	req1.Header.Set("Host", "signer.blockchain.amazonaws.com")
	req1.Host = AwsV4SigHeader
	_, err = SignRequestWithAwsV4UseQueryString(req1,&accessKey,"blockchain","signer")

	//Post the response
	resp, err := myclient.Do(req1)
	if err != nil {
		distributionlogger.Errorf("Validator service check failed")
		return "", false
	}
	defer resp.Body.Close()

	//unmarshall the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}

	var DecData ValidatorResp
	err = json.Unmarshal(respBody, &DecData)
	if err != nil {
		return "", false
	}

	return DecData.Data.RawTx, DecData.OK

}

func (helper *sendHelper)SendDistribution(rawTx, txHash, archNode string) (bool, error)  {
	//1. dial the node to check the connection
	targetNodes := []string{}
	targetNodes = append(targetNodes, archNode, archNodes[0], archNodes[1])
	for _, v := range targetNodes{
		rpcClient, _ := rpc.Dial(v)
		_ = rpcClient.CallContext(context.Background(),nil,"eth_sendRawTransaction", rawTx)
	}

	//wait 30s for on-chain
	time.Sleep(30 * time.Second)
	//get the nonce after time waiting
	nonceAt, _ := fetchNonce(archNode, sysAddr)
	//nonceDB
	var sr SendRecord
	MDB(context.TODO()).Where("raw_tx = ?", rawTx).First(&sr)
	nonceDB := sr.Nonce

	////catch the receipt status
	client, err := ethclient.Dial(archNode)
	if err != nil {
		return false, err
	}
	defer client.Close()
	receipt, err := client.TransactionReceipt(context.TODO(), common.Hash(utils.HexToHash(txHash)))
	if err != nil{
		return false, err
	}
	//1. 没有上链   SR-Status：pending
	//下一次发送，check pending txhash, 如果发送成功，更新状态-->2 or 3


	//2. 上链失败	  SR-Status：fail
	//3. 上链成功	  SR-Status：success

	//use the selection case to verify the success of the tx
	if receipt.Status == uint64(0) || (int64(nonceAt) == nonceDB){
		return false, nil
	}

	return true, nil
}


func (helper *sendHelper)ResendRawTx(rawTx string) (bool, error)  {
	archNode := BestArchNode(archNodes)
	client, _ := rpc.Dial(archNode)
	err := client.CallContext(context.Background(),nil,"eth_sendRawTransaction", rawTx)
	if err != nil{
		return false, err
	}

	return true, nil
}


//ContractEventListening to trace the log of event NotifyRewardSummary after the contract notifyReward
func ContractEventListening(archnode, txhash string) (uint64, uint64, error){
	//use archnode instead for active tracing
	client, err := ethclient.Dial(archnode)
	if err != nil {
		return 0, 0, err
	}
	defer client.Close()
	receipt, err := client.TransactionReceipt(context.TODO(), common.Hash(utils.HexToHash(txhash)))
	if err != nil{
		return 0,0, err
	}

	//catch the receipt status
	if receipt.Status == uint64(1){
		distributionlogger.Debugf("The transaction is success!")
	}
	//take action to handle the receipt logs
	//event NotifyRewardSummary(uint256 inputLength, uint256 okLength), there is no indexed, only in data field
	datastr := hex.EncodeToString(receipt.Logs[0].Data)

	//split the datastr
	datastr = strings.TrimPrefix(datastr, "0x")
	inputLenStr := datastr[:64]
	inputLenStr = removeConZero(inputLenStr)
	inLen, err := utils.DecodeUint64("0x" + inputLenStr)
	if err != nil {
		return 0, 0, nil
	}

	okLenStr := datastr[64:128]
	okLenStr = removeConZero(okLenStr)
	okLen, err := utils.DecodeUint64("0x" + okLenStr)
	if err != nil {
		return 0, 0, nil
	}

	//done no need to do extra operation
	distributionlogger.Infof("The input pools number is %d, and the success execution in contract is %d", inLen, okLen)
	if inLen != okLen {
		distributionlogger.Errorf("There have been data mismatch during execution in contract!")
	}

	return inLen, okLen, nil
}


//PostSend
func PostSend(ctx context.Context, vals []*ValDist, sr *SendRecord) error {
	//normal process
	//vals := []*ValDist{}
	for _, valD := range vals {
		affectedRows, err := updateDisInDB(ctx, valD)
		if affectedRows == int64(0) || err != nil {
			distributionlogger.Errorf("The updating distributed flag error with val addr %s", valD.ValAddr)
			continue
		}
	}

	//when the Send bool is true, update the status to success
	//var sr *SendRecord
	err := UpdateSendRecord(ctx, sr)
	if err != nil {
		distributionlogger.Errorf("Updating the send record table failed")
	}

	distributionlogger.Infof("Sending from %d to %d finished", vals[0].ThisEpoch, vals[0].LastEpoch)
	return nil

}

//updateDisInDB to update distribution in Database
func updateDisInDB(ctx context.Context, valD *ValDist) (int64, error) {
	rw := Reward{}
	eplist := []int64{}
	deltaEP := valD.LastEpoch - valD.ThisEpoch + 1
	for i := valD.ThisEpoch; i <= valD.LastEpoch; i ++ {
		eplist = append(eplist, i)
	}
	db := MDB(ctx).Model(&rw).Where("validator_addr = ? and epoch_index IN ?", valD.ValAddr, eplist).Updates(map[string]interface{}{"distributed": true})
	if db.RowsAffected != deltaEP || db.Error != nil {
		distributionlogger.Errorf("Update distribution in db error")
		return 0, errors.BadRequestError(errors.DatabaseError, "Update distribution in db error")
	}
	return db.RowsAffected, nil
}

//PumpDistInfo to pump the distribution info from database
func PumpDistInfo(ctx context.Context, epStart, epEnd uint64, archiveNode string) (map[string]*big.Int, error) {
	valMapDist := make(map[string]*big.Int)
	//get the vals at the end of this period
	vals, err  := rpcCongressGetAllVals(epEnd, archiveNode)
	if err != nil {
		return nil, err
	}
	for _, val := range vals{
		valdis, err := fetchValToDisWithinEP(ctx,val,epStart,epEnd)
		if err != nil {
			return nil, err
		}
		//filter the value of zero distribution
		if valdis.Distribution.Cmp(big.NewInt(0)) == 0 {
			continue
		}
		valMapDist[val] = valdis.Distribution
	}

	return valMapDist, nil
}


//fetchValDist to fetch all the distribution rewards through some EPs
func fetchValDist(ctx context.Context, valAddr string) (*ValDist, error) {
	valds := []*big.Int{}
	rws := []Reward{}
	//fetch the undis reward
	edrw, err := fetchValEdDis(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	//fetch the todis reward
	torw, err := fetchValToDis(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	rw := MDB(ctx).Where("distributed = ? and validator_addr = ? and epoch_index > ? and epoch_index <= ?", 0, valAddr, edrw.EpochIndex, torw.EpochIndex).FindInBatches(&rws, 36, func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})
	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)

	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
		ThisEpoch: edrw.EpochIndex + int64(1),
		LastEpoch: torw.EpochIndex,
	}, nil
}

//fetchValToDisWithinEP to fetch val epoch rewards during a epoch range
func fetchValToDisWithinEP(ctx context.Context, valAddr string, epStart, epEnd uint64) (*ValDist, error) {
	rws := []Reward{}
	valds := []*big.Int{}
	eplist := []uint64{}
	for i := epStart; i <= epEnd; i ++ {
		eplist = append(eplist, i)
	}
	deltaEP := epEnd - epStart + 1
	//db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	rw := MDB(ctx).Where("validator_addr = ? and epoch_index IN ? and distributed = ?", valAddr, eplist, false).FindInBatches(&rws, int(deltaEP), func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})

	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)
	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
		ThisEpoch: int64(epStart),
		LastEpoch: int64(epEnd),
	}, nil
}

func FetchTotalRewardsEPs(ctx context.Context, epStart, epEnd uint64) (*big.Int, error) {
	eps := []Epoch{}
	total := []*big.Int{}
	eplist := []uint64{}
	for i := epStart; i <= epEnd; i ++ {
		eplist = append(eplist, i)
	}
	deltaEP := epEnd - epStart + 1

	ep := MDB(ctx).Where("epoch_index IN ?", eplist).FindInBatches(&eps, int(deltaEP), func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, ep := range eps{
			rwbig, ok := new(big.Int).SetString(ep.TotalFees, 10)
			if ok{
				total = append(total, rwbig)
			}
		}
		return nil
	})
	distributionlogger.Debugf("The rows affected should be %d", ep.RowsAffected)

	totald := sum(total)
	return totald, nil
}


//just for UT
func fetchValDistForUT(ctx context.Context, EPs int, valAddr string, db *gorm.DB) (*ValDist, error) {
	valds := []*big.Int{}
	rws := []Reward{}
	//fetch the distributed already reward
	edrw, err := fetchValEdDisUT(ctx, valAddr, db)
	if err != nil {
		return nil, err
	}

	//fetch the todis reward
	torw, err := fetchValToDisUT(ctx, valAddr, db)
	if err != nil {
		return nil, err
	}

	rw := db.Where("distributed = ? and validator_addr = ? and epoch_index > ? and epoch_index <= ?", 0, valAddr, edrw.EpochIndex, torw.EpochIndex).FindInBatches(&rws, EPs, func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})
	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)

	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
	}, nil
}


func fetchValEdDis(ctx context.Context, valAddr string) (*Reward, error) {
	rw := &Reward{}
	MDB(ctx).Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 1, valAddr).First(&rw)
	return rw, nil
}

func fetchValToDis(ctx context.Context, valAddr string) (*Reward, error) {
	rw := &Reward{}
	MDB(ctx).Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	return rw, nil
}

//just for UT
func fetchValEdDisUT(ctx context.Context, valAddr string, db *gorm.DB) (*Reward, error) {
	rw := &Reward{}
	db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 1, valAddr).First(&rw)
	return rw, nil
}

//just for UT
func fetchValToDisUT(ctx context.Context, valAddr string, db *gorm.DB) (*Reward, error) {
	rw := &Reward{}
	db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	return rw, nil
}

//jsut for UT within EP range
func fetchValToDisWithinEPUT(ctx context.Context, valAddr string, db *gorm.DB, epStart, epEnd int64) (*ValDist, error) {
	rws := []Reward{}
	valds := []*big.Int{}
	eplist := []int64{}
	for i := epStart; i <= epEnd; i ++ {
		eplist = append(eplist, i)
	}
	deltaEP := epEnd - epStart + 1
	//db.Order("epoch_index DESC").Where("distributed= ? and validator_addr = ?", 0, valAddr).First(&rw)
	rw := db.Where("validator_addr = ? and epoch_index IN ?", valAddr, eplist).FindInBatches(&rws, int(deltaEP), func(tx *gorm.DB, batch int) error {
		//batch processing the results
		for _, rw := range rws{
			rwbig, ok := new(big.Int).SetString(rw.Rewards, 10)
			if ok{
				valds = append(valds, rwbig)
			}
		}
		return nil
	})

	distributionlogger.Debugf("The rows affected should be %d", rw.RowsAffected)
	//get the total distribution
	totald := sum(valds)

	return &ValDist{
		ValAddr: valAddr,
		Distribution: totald,
		ThisEpoch: epStart,
		LastEpoch: epEnd,
	}, nil
}

//just for UT
func updateDisInDBUT(ctx context.Context, valD *ValDist, tx *gorm.DB) (int64, error) {
	rw := Reward{}
	eplist := []int64{}
	deltaEP := valD.LastEpoch - valD.ThisEpoch + 1
	for i := valD.ThisEpoch; i <= valD.LastEpoch; i ++ {
		eplist = append(eplist, i)
	}

	db := tx.Model(&rw).Where("validator_addr = ? and epoch_index IN ?", valD.ValAddr, eplist).Updates(map[string]interface{}{"distributed": true})

	if db.RowsAffected != deltaEP || db.Error != nil {
		distributionlogger.Errorf("Update distribution in db error")
		return 0, errors.BadRequestError(errors.DatabaseError, "Update distribution in db error")
	}
	return db.RowsAffected, nil
}