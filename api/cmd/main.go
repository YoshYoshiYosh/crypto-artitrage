package main

import (
	"crypto-artitrage/api/exchanges/binance"
	"crypto-artitrage/api/exchanges/coincheck"
	"crypto-artitrage/api/exchanges/currency_convert"
	"fmt"
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

func main() {
	start := time.Now()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		yenPricePerDoller = currency_convert.GetYenPricePerDoller(wg)
	}()
	go func() {
		coincheckRate = CoincheckApiClient.CallApi(coincheckApiType["storeRate"], wg)
	}()
	go func() {
		binanceRate = BinanceApiClient.CallApi(binanceApiType["checkPrice"], wg)
	}()
	wg.Wait()

	rates := []float64{coincheckRate, UsdToYen(binanceRate, yenPricePerDoller)}
	bestBuyExchange := pickBestRate("buy", rates)
	bestSellExchange := pickBestRate("sell", rates)

	// Expected profit
	// BitBayが安いかも
	fmt.Println(bestBuyExchange.Rate)
	fmt.Println(bestSellExchange.Rate)
	end := time.Now()
	fmt.Printf("%f秒\n", (end.Sub(start)).Seconds())
}

func pickBestRate(selectType string, rates []float64) BestRate {
	bestRate := selectRate(selectType, rates)
	var bestRateExchange string

	switch bestRate {
	case coincheckRate:
		bestRateExchange = "coincheck"
	case UsdToYen(binanceRate, yenPricePerDoller):
		bestRateExchange = "binance"
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
