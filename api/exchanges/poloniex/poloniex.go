package poloniex

import (
	"bytes"
	"crypto-artitrage/api/exchanges/common"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type PoloniexApiClient common.ApiClient

type ApiPathAndMethod common.ApiPathAndMethod

type RateResponseJson struct {
	USDC_XRP struct {
		// id            int64
		Rate string `json:"last"`
		// lowestAsk     float64
		// highestBid    float64
		// percentChange float64
		// baseVolume    float64
		// quoteVolume   float64
		// isFrozen      int64
		// high24hr      float64
		// low24hr       float64
	} `json:"USDC_XRP"`
}

var ApiType = map[string]ApiPathAndMethod{
	"checkPrice": {"?command=returnTicker", "GET"},
}

const (
	baseUrl = "https://poloniex.com/public"
)

var cfg = common.Config

func NewPoloniexApiClient() *PoloniexApiClient {
	return &PoloniexApiClient{
		Key:         cfg.Section("poloniex").Key("access_key").String(),
		Secret:      cfg.Section("poloniex").Key("secret_key").String(),
		HttpClient:  new(http.Client),
		HttpRequest: new(http.Request),
	}
}

func (client *PoloniexApiClient) SetRequestHeader() {
	// nonce（リクエストごとに増加する数字として現在時刻のUNIXタイム）を設定
	// nonce := strconv.FormatInt(time.Now().Unix(), 10)

	// nonce, リクエストURL, リクエストボディを連結
	body, _ := ioutil.ReadAll(client.HttpRequest.Body)
	// message := nonce + req.URL.String() + string(body)

	// 署名してSignatureを作成
	// hmac := hmac.New(sha256.New, []byte(client.Secret))
	// hmac.Write([]byte(message))
	// sign := hex.EncodeToString(hmac.Sum(nil))

	// ヘッダーをセットする用のmapを作成
	header := map[string]string{
		"Key": client.Key,
	}

	// requestにヘッダーを追加
	for key, value := range header {
		client.HttpRequest.Header.Set(key, value)
	}
	if string(body) != "" {
		client.HttpRequest.Header.Set("Content-Type", "application/json")
	}
}

func (client *PoloniexApiClient) CallApi(pathAndMethod ApiPathAndMethod, wg *sync.WaitGroup) float64 {
	defer wg.Done()
	path, httpMethod := pathAndMethod.Path, pathAndMethod.Method

	requestURL, _ := url.Parse(baseUrl + path)

	var data []byte

	// switch path {
	// case ApiType["checkPrice"].Path:
	// 	requestURL, _ = url.Parse(baseUrl + path + tickerRate["xrp"])
	// default:
	// 	data = []byte{}
	// }

	client.HttpRequest, _ = http.NewRequest(httpMethod, requestURL.String(), bytes.NewBuffer(data))
	client.SetRequestHeader()

	response, err := client.HttpClient.Do(client.HttpRequest)
	if err != nil {
		log.Fatal(err)
	} else if response.StatusCode != 200 {
		log.Fatal("Poloniex APIからのデータ取得に失敗しました")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// ここでchannelに値を渡す
	// switch path {
	// case ApiType["accountBalance"].Path:
	// default:
	// fmt.Println(string(body))
	// }

	var RateResponseJson RateResponseJson
	json.Unmarshal(body, &RateResponseJson)

	rate, _ := strconv.ParseFloat(RateResponseJson.USDC_XRP.Rate, 64)
	return rate
}
