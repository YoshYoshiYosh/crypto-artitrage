package coincheck

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"crypto-artitrage/api/exchanges/common"
)

type CoincheckApiClient common.ApiClient

type ApiPathAndMethod common.ApiPathAndMethod

var cfg = common.Config

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

type RateResponseJson struct {
	Rate string `json:"rate"`
}

func NewCoincheckApiClient() *CoincheckApiClient {
	return &CoincheckApiClient{
		Key:         cfg.Section("coincheck").Key("access_key").String(),
		Secret:      cfg.Section("coincheck").Key("secret_key").String(),
		HttpClient:  new(http.Client),
		HttpRequest: new(http.Request),
	}
}

func (client *CoincheckApiClient) SetRequestHeader() {
	// nonce（リクエストごとに増加する数字として現在時刻のUNIXタイム）を設定
	nonce := strconv.FormatInt(time.Now().Unix(), 10)

	// nonce, リクエストURL, リクエストボディを連結
	body, _ := ioutil.ReadAll(client.HttpRequest.Body)
	message := nonce + client.HttpRequest.URL.String() + string(body)

	// 署名してSignatureを作成
	hmac := hmac.New(sha256.New, []byte(client.Secret))
	hmac.Write([]byte(message))
	sign := hex.EncodeToString(hmac.Sum(nil))

	//リクエストヘッダに初期値を代入しないとエラーになる
	client.HttpRequest.Header = make(http.Header)
	// ヘッダーをセットする用のmapを作成
	header := map[string]string{
		"ACCESS-KEY":       client.Key,
		"ACCESS-NONCE":     nonce,
		"ACCESS-SIGNATURE": sign,
	}

	// requestにヘッダーを追加
	for key, value := range header {
		client.HttpRequest.Header.Set(key, value)
	}

	if string(body) != "" {
		client.HttpRequest.Header.Set("Content-Type", "application/json")
	}
}

// https://teratail.com/questions/116661
func (client *CoincheckApiClient) CallApi(pathAndMethod ApiPathAndMethod, wg *sync.WaitGroup) float64 {
	defer wg.Done()

	path, httpMethod := pathAndMethod.Path, pathAndMethod.Method

	requestURL, _ := url.Parse(baseUrl + path)

	var data []byte

	switch path {
	case ApiType["tradeRate"].Path:
		requestURL, _ = url.Parse(setQueryStringOfRate(baseUrl+path, orderType["buy"], rate["mona"], 1))
	case ApiType["storeRate"].Path:
		requestURL, _ = url.Parse(addPathOfStoreRate(baseUrl+path, storeRate["xrp"]))
	default:
		data = []byte{}
	}

	client.HttpRequest, _ = http.NewRequest(httpMethod, requestURL.String(), bytes.NewBuffer(data))
	client.SetRequestHeader()

	response, err := client.HttpClient.Do(client.HttpRequest)
	if err != nil {
		log.Fatal(err)
	} else if response.StatusCode != 200 {
		log.Fatal("Coincheck APIからのデータ取得に失敗しました")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// ここでchannelに値を渡す
	switch path {
	case ApiType["accountBalance"].Path:
		balanceLog(body)
	default:
		// fmt.Println(string(body))
	}

	var RateResponseJson RateResponseJson
	json.Unmarshal(body, &RateResponseJson)

	rate, _ := strconv.ParseFloat(RateResponseJson.Rate, 64)
	return rate
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
	fmt.Println("  保有数：", info.Xem)
	fmt.Println("--------------------")
	fmt.Println("XLM")
	fmt.Println("  保有数：", info.Xlm)
	fmt.Println("--------------------")
	fmt.Println("QTUM")
	fmt.Println("  保有数：", info.Qtum)
	fmt.Println("--------------------")
	fmt.Println("LSK")
	fmt.Println("  保有数：", info.Lsk)
	fmt.Println("--------------------")
}

const (
	baseUrl = "https://coincheck.com"
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

var ApiType = map[string]ApiPathAndMethod{
	"accountBalance":            {"/api/accounts/balance", "GET"},
	"tradeRate":                 {"/api/exchange/orders/rate", "GET"},
	"storeRate":                 {"/api/rate", "GET"},
	"getTransactions":           {"/api/exchange/orders/transactions", "GET"},
	"getTransactionsPagination": {"/api/exchange/orders/transactions_pagination", "GET"},
	"newOrder":                  {"/api/exchange/orders", "POST"},
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
