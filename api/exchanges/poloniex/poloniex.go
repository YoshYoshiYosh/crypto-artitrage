package poloniex

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

type PoloniexApiClient struct {
	Key         string
	Secret      string
	HttpClient  *http.Client
	HttpRequest *http.Request
}

var cfg, _ = ini.Load("./api/config/config.ini")

func NewPoloniexApiClient() *PoloniexApiClient {
	return &PoloniexApiClient{
		Key:         cfg.Section("poloniex").Key("access_key").String(),
		Secret:      cfg.Section("poloniex").Key("secret_key").String(),
		HttpClient:  new(http.Client),
		HttpRequest: new(http.Request),
	}
}

func (client *PoloniexApiClient) setRequestHeader(req *http.Request) {
	// nonce（リクエストごとに増加する数字として現在時刻のUNIXタイム）を設定
	nonce := strconv.FormatInt(time.Now().Unix(), 10)

	// nonce, リクエストURL, リクエストボディを連結
	body, _ := ioutil.ReadAll(req.Body)
	message := nonce + req.URL.String() + string(body)

	// 署名してSignatureを作成
	hmac := hmac.New(sha256.New, []byte(client.Secret))
	hmac.Write([]byte(message))
	sign := hex.EncodeToString(hmac.Sum(nil))

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
