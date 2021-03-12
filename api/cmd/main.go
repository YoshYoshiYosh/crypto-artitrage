package main

import (
	"crypto-artitrage/api/exchanges/binance"
	"crypto-artitrage/api/exchanges/coincheck"
	"crypto-artitrage/api/exchanges/currency_convert"
	"crypto-artitrage/api/exchanges/poloniex"
	"crypto-artitrage/api/utils"
	"fmt"
	"sync"
	"time"
)

type BestRate struct {
	Exchange string
	Rate     float64
}

// 1ドル何円で買えるか？ の値
var yenPricePerDoller float64

// 投入する最大金額
var maxDeposit float64 = 150000

// 期待する最低売買利益
var expectMinimumProfit = 1000

var CoincheckApiClient = coincheck.NewCoincheckApiClient()
var coincheckApiType = coincheck.ApiType
var coincheckRate float64

var BinanceApiClient = binance.NewBinanceApiClient()
var binanceApiType = binance.ApiType
var binanceRate float64

var PoloniexApiClient = poloniex.NewPoloniexApiClient()
var poloniexApiType = poloniex.ApiType
var poloniexRate float64

func main() {
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
		// TODO： 第2引数にはrateOfExchangeListを渡すように修正
		bestBuyExchange := pickBestRate("buy", rates)
		bestSellExchange := pickBestRate("sell", rates)

		// もっとも良い買いレートと売りレートから、アービトラージ取引した場合の利益を計算する
		willGetProfit := calcProfit(bestBuyExchange.Rate, bestSellExchange.Rate, maxDeposit)

		// 今回のAPIコールの結果を簡易ログファイルに出力
		utils.OutputAllRatesToLog(rateOfExchangeList, willGetProfit)

		// BitBayが安いかも
		fmt.Printf("購入取引所： %s, 購入価格： %f\n", bestBuyExchange.Exchange, bestBuyExchange.Rate)
		fmt.Printf("売却取引所： %s, 売却価格： %f\n", bestSellExchange.Exchange, bestSellExchange.Rate)

		// 「今回のレートによる予想売買利益」 と 「期待する最低売買利益」 を比較し、前者が大きい場合は処理を進める。
		// falseの場合は、再度リクエストを飛ばす
		if willGetProfit > expectMinimumProfit {
			fmt.Println("取引続けます！")
			// bestBuyExchange  の取引所に対して購入APIを投げる

			// bestSellExchange の取引所に 「ウォレット内の該当通貨残高」を取得するAPIを投げる（その1）

			// bestSellExchange の取引所ウォレットに送金する

			// bestSellExchange の取引所に 「ウォレット内の該当通貨残高」を取得するAPIを投げる（その2）
			// → その1の結果と比較して、残高が増えたタイミングで次の処理に進む（残高が増えたことがわかるまで、1秒スパンでAPIコール）

			// bestSellExchange の取引所に対して売却APIを投げる
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

func makeRatesList(rateOfExchangeList []map[string]float64) []float64 {
	rates := []float64{}
	for _, info := range rateOfExchangeList {
		for _, rate := range info {
			rates = append(rates, rate)
		}
	}
	return rates
}

func calcProfit(bestBuyRate, bestSellRate, maxDeposit float64) int {
	willGetProfit := (maxDeposit/bestBuyRate)*bestSellRate - maxDeposit
	return int(willGetProfit)
}
