package binance

import (
	"bytes"
	"crypto-artitrage/api/exchanges/common"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type BinanceApiClient common.ApiClient

type ApiPathAndMethod common.ApiPathAndMethod

type RateResponseJson struct {
	Symbol string `json:"symbol"`
	Rate   string `json:"price"`
}

// https://api.binance.com/api/v3/ticker/price?symbol=XRPUSDT
// https://www.binance.com/ja/trade/XRP_USDT?layout=pro
var ApiType = map[string]ApiPathAndMethod{
	"checkPrice": {"/api/v3/ticker/price?", "GET"},
}

const (
	baseUrl = "https://api.binance.com"
)

var tickerRate = map[string]string{
	"xrp": "symbol=XRPUSDT",
}

var cfg = common.Config

func NewBinanceApiClient() *BinanceApiClient {
	return &BinanceApiClient{
		Key:         cfg.Section("binance").Key("access_key").String(),
		Secret:      cfg.Section("binance").Key("secret_key").String(),
		HttpClient:  new(http.Client),
		HttpRequest: new(http.Request),
	}
}

func (client *BinanceApiClient) SetRequestHeader() {
	// nonce（リクエストごとに増加する数字として現在時刻のUNIXタイム）を設定
	nonce := strconv.FormatInt(time.Now().Unix(), 10)

	// nonce, リクエストURL, リクエストボディを連結
	body, _ := ioutil.ReadAll(client.HttpRequest.Body)
	message := nonce + client.HttpRequest.URL.String() + string(body)

	// 署名してSignatureを作成
	hmac := hmac.New(sha256.New, []byte(client.Secret))
	hmac.Write([]byte(message))
	// sign := hex.EncodeToString(hmac.Sum(nil))

	//リクエストヘッダに初期値を代入しないとエラーになる
	client.HttpRequest.Header = make(http.Header)
	// ヘッダーをセットする用のmapを作成
	header := map[string]string{
		"X-MBX-APIKEY": client.Key,
	}
	// header := map[string]string{
	// 	"ACCESS-KEY":       client.Key,
	// 	"ACCESS-NONCE":     nonce,
	// 	"ACCESS-SIGNATURE": sign,
	// }

	// requestにヘッダーを追加
	for key, value := range header {
		client.HttpRequest.Header.Set(key, value)
	}

	if string(body) != "" {
		client.HttpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
}

func (client *BinanceApiClient) CallApi(pathAndMethod ApiPathAndMethod, wg *sync.WaitGroup) float64 {
	defer wg.Done()
	path, httpMethod := pathAndMethod.Path, pathAndMethod.Method

	requestURL, _ := url.Parse(baseUrl + path)

	var data []byte

	switch path {
	case ApiType["checkPrice"].Path:
		requestURL, _ = url.Parse(baseUrl + path + tickerRate["xrp"])
	default:
		data = []byte{}
	}

	client.HttpRequest, _ = http.NewRequest(httpMethod, requestURL.String(), bytes.NewBuffer(data))
	client.SetRequestHeader()

	response, err := client.HttpClient.Do(client.HttpRequest)
	if err != nil {
		log.Fatal(err)
	} else if response.StatusCode != 200 {
		log.Fatal("Binance APIからのデータ取得に失敗しました")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// ここでchannelに値を渡す
	switch path {
	case ApiType["accountBalance"].Path:
	default:
		// fmt.Println(string(body))
	}

	var RateResponseJson RateResponseJson
	json.Unmarshal(body, &RateResponseJson)

	rate, _ := strconv.ParseFloat(RateResponseJson.Rate, 64)
	return rate
}
