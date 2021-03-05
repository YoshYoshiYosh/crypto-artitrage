package main

import (
	"crypto-artitrage/api/exchanges/coincheck"
)

var CoincheckApiClient = coincheck.NewCoincheckApiClient()
var coincheckApiType = coincheck.ApiType

func main() {
	// goroutineで動かしたい
	CoincheckApiClient.CallApi(coincheckApiType["accountBalance"])
	CoincheckApiClient.CallApi(coincheckApiType["storeRate"])
}
