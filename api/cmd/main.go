package main

import (
	"bytes"
	"crypto-artitrage/api/exchanges/binance"
	"crypto-artitrage/api/exchanges/coincheck"
	"crypto-artitrage/api/exchanges/currency_convert"
	"crypto-artitrage/api/exchanges/poloniex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BestRate struct {
	Exchange string
	Rate     float64
}

var yenPricePerDoller float64

var CoincheckApiClient = coincheck.NewCoincheckApiClient()
var coincheckApiType = coincheck.ApiType
var coincheckRate float64

var BinanceApiClient = binance.NewBinanceApiClient()
var binanceApiType = binance.ApiType
var binanceRate float64

var PoloniexApiClient = poloniex.NewPoloniexApiClient()
var poloniexApiType = poloniex.ApiType
var poloniexRate float64

var expectMinimumProfit = 1000

func main() {
	// スクレイピング結果を取得するときの呼び方例
	// result := utils.Scraping()
	// fmt.Println(result)
	for {
		start := time.Now()
		wg := &sync.WaitGroup{}
		wg.Add(4)
		go func() {
			yenPricePerDoller = currency_convert.GetYenPricePerDoller(wg)
		}()
		go func() {
			coincheckRate = CoincheckApiClient.CallApi(coincheckApiType["storeRate"], wg)
		}()
		go func() {
			binanceRate = BinanceApiClient.CallApi(binanceApiType["checkPrice"], wg)
		}()
		go func() {
			poloniexRate = PoloniexApiClient.CallApi(poloniexApiType["checkPrice"], wg)
		}()
		wg.Wait()

		// 取引所の名前と、その取引所のレートをマップ形式で格納
		rateOfExchangeList := []map[string]float64{
			{"Coincheck": coincheckRate},
			{"Binance": UsdToYen(binanceRate, yenPricePerDoller)},
			{"Poloniex": UsdToYen(poloniexRate, yenPricePerDoller)},
		}

		// 各取引所のレートだけを配列に格納
		rates := makeRatesList(rateOfExchangeList)

		// 各取引所のレートを比較し、もっとも良い買いレート、売りレートを計算
		bestBuyExchange := pickBestRate("buy", rates)
		bestSellExchange := pickBestRate("sell", rates)

		// もっとも良い買いレートと売りレートから、アービトラージ取引した場合の利益を計算する
		willGetProfit := calcProfit(bestBuyExchange.Rate, bestSellExchange.Rate)

		// 今回のAPIコールの結果を簡易ログファイルに出力
		outputAllRatesToLog(rateOfExchangeList, willGetProfit)

		// BitBayが安いかも
		fmt.Printf("購入取引所： %s, 購入価格： %f\n", bestBuyExchange.Exchange, bestBuyExchange.Rate)
		fmt.Printf("売却取引所： %s, 売却価格： %f\n", bestSellExchange.Exchange, bestSellExchange.Rate)

		// 「今回のレートによる予想売買利益」 と 「期待する最低売買利益」 を比較し、前者が大きい場合は処理を進める。
		// falseの場合は、再度リクエストを飛ばす
		if willGetProfit > expectMinimumProfit {
			fmt.Println("取引続けます！")
		} else {
			fmt.Println("利益少ないんで無理っす..")
		}

		end := time.Now()
		fmt.Printf("%f秒\n", (end.Sub(start)).Seconds())
		time.Sleep(3 * time.Second)
	}
}

func pickBestRate(selectType string, rates []float64) BestRate {
	bestRate := selectRate(selectType, rates)
	var bestRateExchange string

	switch bestRate {
	case coincheckRate:
		bestRateExchange = "coincheck"
	case UsdToYen(binanceRate, yenPricePerDoller):
		bestRateExchange = "binance"
	case UsdToYen(poloniexRate, yenPricePerDoller):
		bestRateExchange = "poloniex"
	}

	return BestRate{bestRateExchange, bestRate}
}

func UsdToYen(usd, rate float64) float64 {
	return usd * rate
}

func selectRate(selectType string, rates []float64) float64 {
	var bestRate float64

	if selectType == "buy" {
		for i := 0; i < len(rates); i++ {
			if i == 0 {
				bestRate = rates[0]
				continue
			}
			if rates[i] < rates[i-1] {
				bestRate = rates[i]
			}
		}
	} else {
		for i := 0; i < len(rates); i++ {
			if i == 0 {
				bestRate = rates[0]
				continue
			}
			if rates[i] > rates[i-1] {
				bestRate = rates[i]
			}
		}
	}
	return bestRate
}

func outputAllRatesToLog(rateOfExchangeList []map[string]float64, expectProfit int) {
	logFile, err := os.OpenFile("./api/logs/log.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()

	logHeader := []byte(fmt.Sprintf("%s取引所%sレート", strings.Repeat(" ", 3), strings.Repeat(" ", 10)))
	_, err = logFile.Write([]byte(string(logHeader) + string("\n")))

	for _, info := range rateOfExchangeList {
		for exchangeName, rate := range info {
			rateToString := strconv.FormatFloat(rate, 'f', -1, 64)
			inputToByte := []byte(fmt.Sprintf("%-15s", exchangeName+":") + rateToString + "\n")
			_, err = logFile.Write([]byte(string(inputToByte)))

			if err != nil {
				log.Fatal(err)
			}
		}
	}
	aboutProfit := []byte(fmt.Sprintf("予想売却益： %d円", expectProfit))
	_, err = logFile.Write([]byte("\n" + string(aboutProfit) + string("\n")))

	afterEqual := bytes.Repeat([]byte("-"), 30)
	_, err = logFile.Write([]byte(string(afterEqual) + string("\n")))

}

func makeRatesList(rateOfExchangeList []map[string]float64) []float64 {
	rates := []float64{}
	for _, info := range rateOfExchangeList {
		for _, rate := range info {
			rates = append(rates, rate)
		}
	}
	return rates
}

func calcProfit(bestBuyRate, bestSellRate float64) int {
	maxDeposit := float64(150000)

	willGetProfit := (maxDeposit/bestBuyRate)*bestSellRate - maxDeposit
	return int(willGetProfit)
}
