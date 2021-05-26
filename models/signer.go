package models

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sha1sum/aws_signing_client"
	"github.com/starslabhq/rewards-collection/utils"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
)

type Payload struct {
	Addrs  				[]string  `json:"addrs"`
	Data 				ReqData    `json:"data"`
	Chain    			string 	  `json:"chain"`
	EncryptParams  		EncParams    `json:"encrypt_params"`
}

type ReqData struct {
	//ToTag is the input data for contract revoking params
	ToTag		string			`json:"to_tag"`
	Asset		string			`json:"asset"`
	Decimal     int				`json:"decimal"`
	Platform	string			`json:"platform"`
	Nonce       int				`json:"nonce"`
	From   		string			`json:"from"`
	//To is the contract Addr
	To			string			`json:"to"`
	//GasLimit here
	FeeStep		string			`json:"fee_step"`
	//GasPrice here
	FeePrice    string			`json:"fee_price"`
	FeeAsset	string			`json:"fee_asset"`
}

type EncParams struct {
	Tasks      []Task		`json:"tasks"`
	TxType     string		`json:"tx_type"`
}

type Task struct {
	TaskId 		string			`json:"task_id"`
	UserId 		string			`json:"user_id"`
	OriginAddr	string			`json:"origin_addr"`
	TaskType	string			`json:"task_type"`
}

type Response struct {
	Result 		bool	`json:"result"`
	Data        RespData `json:"data"`
}

type RespData struct {
	EncryptData   string `json:"encrypt_data"`
	Extra         RespEx  `json:"extra"`
}

type RespEx struct {
	Cipher		string 		`json:"cipher"`
	TxHash      string		`json:"txhash"`
}

func fetchNonce(archnode, addr string) (int, error) {
	client, err := ethclient.Dial(archnode)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	//addr in hex string
	commonAddr := utils.HexToAddress(addr)
	nonce, err := client.NonceAt(context.TODO(), common.Address(commonAddr),nil)
	if err != nil {
		return 0, err
	}
	return int(nonce), nil
}

func signGateway(archNode, sysAddr string, valMapDist map[string]*big.Int)  {
	//var credentials *credentials.Credentials
	signer := v4.NewSigner(credentials.AnonymousCredentials)

	//var myClient *http.Client
	var (
		serverCrt = "/Users/wangming/data/gateway_service/server.cer.pem"
		clientCrt = "/Users/wangming/data/gateway_service/client.cer.pem"
		clientKey = "/Users/wangming/data/gateway_service/client.key.pem"
	)

	myClient := TwoWaySSlWithClient(serverCrt, clientCrt, clientKey)
	awsClient, err := aws_signing_client.New(signer, myClient, "signer", "blockchain")
	if err != nil {
		return
	}
	//testing url
	Url := "https://172.18.23.38:21000/gateway/sign"

	//fetch the contract data
	dataStr := getNotifyAmountData(valMapDist)

	//fetch toaddr nonce
	nonce, err := fetchNonce(archNode, sysAddr)
	if err != nil {
		return
	}

	contractAddr := "0x5CaeF96c490b5c357847214395Ca384dC3d3b85e"

	//assemble the data field for sending transaction
	reqData := ReqData{
		To: contractAddr,
		ToTag: dataStr,
		Nonce: nonce,
		Asset: "ht",
		Decimal: 18,
		Platform: "starlabsne3",
		From: sysAddr,
		FeeStep: "200000",
		FeePrice: "40000000000",
		FeeAsset: "ht",
	}

	//reqDataByte, err := json.Marshal(reqData)
	//if err != nil {
	//	return
	//}
	//string(reqDataByte)

	encPara := EncParams{
		Tasks: []Task{
			{TaskId: "0",
			TaskType: "",
			UserId: "",
			OriginAddr: "",
			},
		},
		TxType: "transfer",
	}
	//encParaByte, err := json.Marshal(encPara)
	//if err != nil {
	//	return
	//}
	//string(encParaByte)

	data := &Payload{
		Addrs: []string{sysAddr},
		Chain: "ht2",
		Data: reqData,
		EncryptParams: encPara,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	body := bytes.NewReader(payloadBytes)

	resp, err := awsClient.Post(Url, "application/json", body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println(string(respBody))
	//take some action to parse the response body

}

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}


func addTrust(pool *x509.CertPool, path string) {
	caCrt, err := ioutil.ReadFile(path)
	if err!= nil {
		fmt.Println("ReadFile err:",err)
		return
	}
	pool.AppendCertsFromPEM(caCrt)
}

func TwoWaySSlWithClient(serverCrt, clientCrt, clientKey string) *http.Client {
	//The sslfile dir is the directory for store some files after decryption
	pool := x509.NewCertPool()
	// This loads the certificate provided by the server to verify the data returned by the server.
	addTrust(pool,serverCrt)
	//Here to load the client's own certificate, to be consistent with the certificate provided to the server, otherwise the server verification will not pass
	cliCrt, err := tls.LoadX509KeyPair(clientCrt, clientKey)
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		return nil
	}

	//use the transport for the ssl config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cliCrt},
		},
	}
	client := &http.Client{Transport: tr, Timeout: 123 * time.Second}
	return client
}

