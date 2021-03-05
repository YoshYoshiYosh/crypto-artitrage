package main

import (
	"crypto-artitrage/api/exchanges/coincheck"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	ini "gopkg.in/ini.v1"
)

type ApiClient struct {
	key        string
	secret     string
	httpClient *http.Client
}

type ConfigList struct {
	accessKey string
	secretKey string
}

type BalanceInfo struct {
	Success      bool   `json:"success"`
	Jpy          string `json:"jpy"`
	Btc          string `json:"btc"`
	Iost         string `json:"iost"`
	Xem          string `json:"xem"`
	Xlm          string `json:"xlm"`
	Qtum         string `json:"qtum"`
	Lsk          string `json:"lsk"`
	JpyReserved  string `json:"jpy_reserved"`
	BtcReserved  string `json:"btc_reserved"`
	JpyLendInUse string `json:"jpy_lend_in_use"`
	BtcLendInUse string `json:"btc_lend_in_use"`
	JpyLent      string `json:"jpy_lent"`
	BtcLent      string `json:"btc_lent"`
	JpyDebt      string `json:"jpy_debt"`
	BtcDebt      string `json:"btc_debt"`
}

type RateRequestJson struct {
	OrderType string  `json:"order_type"`
	Pair      string  `json:"pair"`
	Amount    float64 `json:"amount"`
}

var Config ConfigList

var CoincheckApiClient = coincheck.NewCoincheckApiClient()

func init() {
	cfg, _ := ini.Load("./api/config/config.ini") // main.goの格納ディレクトリからの相対パスでなく、「go runコマンドの実行ディレクトリ」からの相対パスを引数として渡す

	Config = ConfigList{
		accessKey: cfg.Section("coincheck").Key("access_key").String(),
		secretKey: cfg.Section("coincheck").Key("secret_key").String(),
	}
}

const (
	base_url = "https://coincheck.com"
)

var orderType = map[string]string{
	"sell": "sell",
	"buy":  "buy",
}

var rate = map[string]string{
	"btc":  "btc_jpy",
	"etc":  "etc_jpy",
	"fct":  "fct_jpy",
	"mona": "mona_jpy",
}

var storeRate = map[string]string{
	"btc":  "btc_jpy",
	"eth":  "eth_jpy",
	"etc":  "etc_jpy",
	"lsk":  "lsk_jpy",
	"fct":  "fct_jpy",
	"xrp":  "xrp_jpy",
	"xem":  "xem_jpy",
	"ltc":  "ltc_jpy",
	"bch":  "bch_jpy",
	"mona": "mona_jpy",
	"xlm":  "xlm_jpy",
	"qtum": "qtum_jpy",
	"bat":  "bat_jpy",
	"iost": "iost_jpy",
	"enj":  "enj_jpy",
}

var apiType = map[string]string{
	"accountBalance":            "/api/accounts/balance",
	"tradeRate":                 "/api/exchange/orders/rate",
	"storeRate":                 "/api/rate",
	"getTransactions":           "/api/exchange/orders/transactions",
	"getTransactionsPagination": "/api/exchange/orders/transactions_pagination",
	"newOrder":                  "/api/exchange/orders",
}

func getRateRequestBody(orderType, pair string) []byte {
	body := RateRequestJson{
		OrderType: orderType,
		Pair:      pair,
		Amount:    100,
	}
	json, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}
	return json
}

func setQueryStringOfRate(url, orderType, pair string, amount float64) string {
	return fmt.Sprintf("%s?order_type=%s&pair=%s&amount=%s", url, orderType, pair, strconv.FormatFloat(float64(amount), 'f', -1, 64))
}

func addPathOfStoreRate(url, pair string) string {
	return fmt.Sprintf("%s/%s", url, pair)
}

func setRequestHeader(req *http.Request) {
	fmt.Println("req.Body")
	fmt.Println(req.Body)
	// nonce（リクエストごとに増加する数字として現在時刻のUNIXタイム）を設定
	nonce := strconv.FormatInt(time.Now().Unix(), 10)

	// nonce, リクエストURL, リクエストボディを連結
	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println("body")
	fmt.Println(string(body))
	message := nonce + req.URL.String() + string(body)
	fmt.Println("message")
	fmt.Println(message)

	// 署名してSignatureを作成
	hmac := hmac.New(sha256.New, []byte(Config.secretKey))
	hmac.Write([]byte(message))
	sign := hex.EncodeToString(hmac.Sum(nil))

	// ヘッダーをセットする用のmapを作成
	header := map[string]string{
		"ACCESS-KEY":       Config.accessKey,
		"ACCESS-NONCE":     nonce,
		"ACCESS-SIGNATURE": sign,
	}

	// requestにヘッダーを追加
	for key, value := range header {
		req.Header.Set(key, value)
	}
	// if string(body) != "" {
	req.Header.Set("Content-Type", "application/json")
	// }
}

func balanceLog(responseBody []byte) {
	var info BalanceInfo
	json.Unmarshal(responseBody, &info)

	fmt.Println("--------------------")
	fmt.Println("BTC")
	fmt.Println("  保有数：", info.Btc)
	fmt.Println("--------------------")
	fmt.Println("IOST")
	fmt.Println("  保有数：", info.Iost)
	fmt.Println("--------------------")
	fmt.Println("XEM")
	fmt.Println("--------------------")
	fmt.Println("XLM")
	fmt.Println("  保有数：", info.Xlm)
	fmt.Println("  保有数：", info.Xem)
	fmt.Println("--------------------")
	fmt.Println("QTUM")
	fmt.Println("  保有数：", info.Qtum)
	fmt.Println("--------------------")
	fmt.Println("LSK")
	fmt.Println("  保有数：", info.Lsk)
	fmt.Println("--------------------")
}

func main() {
	path := apiType["storeRate"]
	CoincheckApiClient.CallApi(path)
}

// func main() {

// 	coincheck := coincheck.NewCoincheckApiClient()
// 	fmt.Println(CoincheckApiClient.HttpRequest.URL)
// 	// path := apiType["accountBalance"]
// 	// path := apiType["tradeRate"]
// 	path := apiType["storeRate"]
// 	// path := apiType["getTransactions"]
// 	// path := apiType["getTransactionsPagination"]

// 	requestURL, err := url.Parse(base_url + path)

// 	var data []byte

// 	switch path {
// 	case apiType["accountBalance"]:
// 	case apiType["getTransactions"]:
// 		data = []byte{}
// 	case apiType["tradeRate"]:
// 		requestURL, _ = url.Parse(setQueryStringOfRate(base_url+path, orderType["buy"], rate["mona"], 1))
// 	case apiType["storeRate"]:
// 		requestURL, _ = url.Parse(addPathOfStoreRate(base_url+path, storeRate["xrp"]))
// 		// requestURL, _ = url.Parse(addPathOfStoreRate(base_url+path, storeRate["iost"]))
// 	default:
// 		data = []byte{}
// 	}

// 	// 2つ以上のAPIを呼び出すときは、goroutineでリクエストを飛ばす（=リクエストを関数化する）

// 	request, _ := http.NewRequest("GET", requestURL.String(), bytes.NewBuffer(data))
// 	setRequestHeader(request)

// 	// CoincheckApiClient.HttpRequest, _ = http.NewRequest("GET", requestURL.String(), bytes.NewBuffer(data))
// 	// fmt.Println(CoincheckApiClient.HttpRequest.URL)

// 	httpClient := new(http.Client)

// 	response, err := httpClient.Do(request)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	body, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		log.Fatal()
// 	}

// 	switch path {
// 	case apiType["accountBalance"]:
// 		balanceLog(body)
// 	case apiType["tradeRate"]:

// 	case apiType["storeRate"]:
// 		fmt.Println(string(body))
// 	case apiType["getTransactions"]:
// 	default:
// 	}
// }
