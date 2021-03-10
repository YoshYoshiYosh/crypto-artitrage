package currency_convert

import (
	"bytes"
	"crypto-artitrage/api/exchanges/common"
	"crypto-artitrage/api/utils"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type yenPricePerDoller struct {
	Rate float64 `json:"USD_JPY"`
}

const (
	baseUrl = "https://free.currconv.com"
	apiPath = "/api/v7/convert?q=USD_JPY&compact=ultra&apiKey="
)

var apiKey = common.Config.Section("currency_convert").Key("access_key").String()

func GetYenPricePerDoller(wg *sync.WaitGroup) float64 {
	defer wg.Done()

	client := new(http.Client)
	requestUrl, _ := url.Parse(baseUrl + apiPath + apiKey)

	request, _ := http.NewRequest("GET", requestUrl.String(), bytes.NewBuffer([]byte{}))

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else if response.StatusCode != 200 {
		// ここにスクレイピングの処理入れる？
		log.Println("通貨レートの取得に失敗しました。")
		return utils.Scraping()
	}

	body, _ := ioutil.ReadAll(response.Body)

	var yenPricePerDoller yenPricePerDoller
	json.Unmarshal(body, &yenPricePerDoller)

	return yenPricePerDoller.Rate
}
