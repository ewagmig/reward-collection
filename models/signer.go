package models

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/starslabhq/rewards-collection/utils"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// head key, case insensitive
const (
	headKeyData          = "date"
	headKeyXAmzDate      = "X-Amz-Date"
	headKeyAuthorization = "authorization"
	headKeyHost          = "host"
	iSO8601BasicFormat = "20060102T150405Z"
	iSO8601BasicFormatShort = "20060102"
)
var lf = []byte{'\n'}

// url query params
const (
	queryKeySignature        = "X-Amz-Signature"
	queryKeyAlgorithm        = "X-Amz-Algorithm"
	queryKeyCredential       = "X-Amz-Credential"
	queryKeyDate             = "X-Amz-Date"
	queryKeySignatureHeaders = "X-Amz-SignedHeaders"
)

const (
	aws4HmacSha256Algorithm = "AWS4-HMAC-SHA256"
)

// Key holds a set of Amazon Security Credentials.
type Key struct {
	AccessKey string
	SecretKey string
}


type Payload struct {
	Addrs  				[]string  `json:"addrs"`
	Data 				string    `json:"data"`
	Chain    			string 	  `json:"chain"`
	EncryptParams  		string    `json:"encrypt_params"`
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	myclient := &http.Client{Transport: tr, Timeout: 123 * time.Second}

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
	reqData := &ReqData{
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



	reqDataByte, err := json.Marshal(reqData)
	if err != nil {
		return
	}


	encPara := &EncParams{
		Tasks: []Task{
			{TaskId: "0",
			TaskType: "",
			UserId: "",
			OriginAddr: "",
			},
		},
		TxType: "transfer",
	}
	encParaByte, err := json.Marshal(encPara)
	if err != nil {
		return
	}


	fmt.Println(string(reqDataByte), string(encParaByte))

	//testing data with template
	//testData := "{\\\"to_tag\\\":\\\"0x867904b4000000000000000000000000e8a3dcb34da80f2cc625671ab5d3228ef32b3b580000000000000000000000000000000000000000000000000000000124101100\\\",\\\"nonce\\\":225,\\\"decimal\\\":8,\\\"asset\\\":\\\"husd\\\",\\\"platform\\\":\\\"husd\\\",\\\"from\\\":\\\"e8a3dcb34da80f2cc625671ab5d3228ef32b3b58\\\",\\\"to\\\":\\\"83aa0b40e0cebc79b957b973769dbea55b2c485b\\\",\\\"fee_step\\\":90000,\\\"fee_price\\\":20000000000,\\\"fee_asset\\\":\\\"eth\\\",\\\"amount\\\":0,\\\"chain_id\\\":1337}"
	//enp := "{\\\"tasks\\\":[{\\\"extra\\\":\\\"{\\\\\\\"fee\\\\\\\":\\\\\\\"100000000\\\\\\\"}\\\",\\\"task_id\\\":\\\"190813173720232242\\\",\\\"user_id\\\":\\\"220\\\",\\\"task_type\\\":\\\"1\\\",\\\"origin_addr\\\":\\\"e8a3dcb34da80f2cc625671ab5d3228ef32b3b58\\\"}],\\\"tx_type\\\":\\\"contract,issue\\\"}"

	data := &Payload{
		Addrs: []string{sysAddr},
		Chain: "ht2",
		Data: "'" + string(reqDataByte) + "'",
		EncryptParams: "'" + string(encParaByte) + "'",
	}

	fmt.Println("The string concat is", data.Data, data.EncryptParams)

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	body := bytes.NewReader(payloadBytes)


	req1, err := http.NewRequest("POST", Url, body)
	req1.Header.Set("content-type", "application/json")
	req1.Header.Set("Host", "signer.blockchain.amazonaws.com")
	key := &Key{
		AccessKey: "gateway",
		SecretKey: "12345678",
	}
	//fmt.Println("The request Body is:", string(payloadBytes))

	req1.Host = "signer.blockchain.amazonaws.com"
	sp, err := SignRequestWithAwsV4UseQueryString(req1,key,"blockchain","signer")
	distributionlogger.Infof("the sp is %v", sp)
	resp, err := myclient.Do(req1)

	//resp, err := awsClient.Post(Url, "application/json", body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println(string(respBody))
	//todo take some action to parse the response body, get rawTransaction, security check PM

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
	//cliCrt, err := tls.LoadX509KeyPair(clientCrt, clientKey)
	//if err != nil {
	//	fmt.Println("Loadx509keypair err:", err)
	//	return nil
	//}

	//use the transport for the ssl config
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			//RootCAs:      pool,
			//Certificates: []tls.Certificate{cliCrt},
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr, Timeout: 123 * time.Second}
	return client
}
// Sign ...
func (k *Key) Sign(t time.Time, region, name string) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(region))
	h = ghmac(h, []byte(name))
	h = ghmac(h, []byte("aws4_request"))
	return h
}
func SignRequestWithAwsV4UseQueryString(req *http.Request, key *Key, region, name string) (sp *SignProcess, err error) {
	date := req.Header.Get(headKeyData)
	t := time.Now().UTC()
	if date != "" {
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return
		}
	}
	values := req.URL.Query()
	values.Set(headKeyXAmzDate, t.Format(iSO8601BasicFormat))

	//req.Header.Set(headKeyHost, req.Host)

	sp = new(SignProcess)
	sp.Key = key.Sign(t, region, name)

	values.Set(queryKeyAlgorithm, aws4HmacSha256Algorithm)
	values.Set(queryKeyCredential, key.AccessKey+"/"+creds(t, region, name))
	cc := bytes.NewBufferString("")
	writeHeaderList(req, nil, cc, false)
	values.Set(queryKeySignatureHeaders, cc.String())
	req.URL.RawQuery = values.Encode()

	writeStringToSign(t, req, nil, sp, false, region, name)
	values = req.URL.Query()
	values.Set(queryKeySignature, hex.EncodeToString(sp.AllSHA256))
	req.URL.RawQuery = values.Encode()

	return
}

func creds(t time.Time, region, name string) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + region + "/" + name + "/aws4_request"
}

func gsha256(data []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

type SignProcess struct {
	Key           []byte
	Body          []byte
	BodySHA256    []byte
	Request       []byte
	RequestSHA256 []byte
	All           []byte
	AllSHA256     []byte
}

func writeHeaderList(r *http.Request, signedHeadersMap map[string]bool, requestData io.Writer, isServer bool) {
	a := make([]string, 0)
	for k := range r.Header {
		if isServer {
			if _, ok := signedHeadersMap[strings.ToLower(k)]; !ok {
				continue
			}
		}
		a = append(a, strings.ToLower(k))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{';'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeStringToSign(
	t time.Time,
	r *http.Request,
	signedHeadersMap map[string]bool,
	sp *SignProcess,
	isServer bool,
	region, name string) {
	lastData := bytes.NewBufferString(aws4HmacSha256Algorithm)
	lastData.Write(lf)

	lastData.Write([]byte(t.Format(iSO8601BasicFormat)))
	lastData.Write(lf)

	lastData.Write([]byte(creds(t, region, name)))
	lastData.Write(lf)

	writeRequest(r, signedHeadersMap, sp, isServer)
	lastData.WriteString(hex.EncodeToString(sp.RequestSHA256))
	// fmt.Fprintf(lastData, "%x", sp.RequestSHA256)

	sp.All = lastData.Bytes()
	sp.AllSHA256 = ghmac(sp.Key, sp.All)
}

func writeRequest(r *http.Request, signedHeadersMap map[string]bool, sp *SignProcess, isServer bool) {
	requestData := bytes.NewBufferString("")
	//content := strings.Split(r.Host, ":")
	r.Header.Set(headKeyHost, "signer.blockchain.amazonaws.com")


	requestData.Write([]byte(r.Method))
	requestData.Write(lf)

	writeURI(r, requestData)
	requestData.Write(lf)

	writeQuery(r, requestData)
	requestData.Write(lf)

	writeHeader(r, signedHeadersMap, requestData, isServer)
	requestData.Write(lf)
	requestData.Write(lf)

	writeHeaderList(r, signedHeadersMap, requestData, isServer)
	requestData.Write(lf)

	writeBody(r, requestData, sp)

	sp.Request = requestData.Bytes()
	sp.RequestSHA256 = gsha256(sp.Request)
}

func writeURI(r *http.Request, requestData io.Writer) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	_, _ = requestData.Write([]byte(path))
}

func writeQuery(r *http.Request, requestData io.Writer) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		if strings.ToLower(k) == queryKeySignature {
			continue
		}
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write([]byte{'&'})
		}
		_, _ = requestData.Write([]byte(s))
	}
}

func writeHeader(r *http.Request, signedHeadersMap map[string]bool, requestData *bytes.Buffer, isServer bool) {
	a := make([]string, 0)
	for k, v := range r.Header {
		if isServer {
			if _, ok := signedHeadersMap[strings.ToLower(k)]; !ok {
				continue
			}
		}
		sort.Strings(v)
		a = append(a, strings.ToLower(k)+":"+strings.Join(v, ","))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			_, _ = requestData.Write(lf)
		}
		_, _ = requestData.WriteString(s)
	}
}

func writeBody(r *http.Request, requestData io.StringWriter, sp *SignProcess) {
	var b []byte
	// If the payload is empty, use the empty string as the input to the SHA256 function
	// http://docs.amazonwebservices.com/general/latest/gr/sigv4-create-canonical-request.html
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	sp.Body = b

	sp.BodySHA256 = gsha256(b)
	_, _ = requestData.WriteString(hex.EncodeToString(sp.BodySHA256))
}

func (p *SignProcess) String() string {
	result := new(strings.Builder)
	fmt.Fprintf(result, "key(hex): %s\n\n", hex.EncodeToString(p.Key))
	fmt.Fprintf(result, "body:\n%s\n", string(p.Body))
	fmt.Fprintf(result, "body sha256: %s\n\n", hex.EncodeToString(p.BodySHA256))
	fmt.Fprintf(result, "request:\n%s\n", string(p.Request))
	fmt.Fprintf(result, "request sha256: %s\n\n", hex.EncodeToString(p.RequestSHA256))
	fmt.Fprintf(result, "all:\n%s\n", string(p.All))
	fmt.Fprintf(result, "all sha256: %s\n", hex.EncodeToString(p.AllSHA256))
	return result.String()
}
func SignRequestWithAwsV4(req *http.Request, key *Key, region, name string) (sp *SignProcess, err error) {
	date := req.Header.Get(headKeyData)
	t := time.Now().UTC()
	if date != "" {
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return
		}
	}
	req.Header.Set(headKeyXAmzDate, t.Format(iSO8601BasicFormat))

	sp = new(SignProcess)
	sp.Key = key.Sign(t, region, name)
	writeStringToSign(t, req, nil, sp, false, region, name)

	auth := bytes.NewBufferString(aws4HmacSha256Algorithm + " ")
	auth.Write([]byte("Credential=" + key.AccessKey + "/" + creds(t, region, name)))
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("SignedHeaders="))
	writeHeaderList(req, nil, auth, false)
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("Signature=" + hex.EncodeToString(sp.AllSHA256)))

	req.Header.Set(headKeyAuthorization, auth.String())
	return
}