package main

import (
	"crypto-artitrage/api/exchanges/coincheck"
)

var CoincheckApiClient = coincheck.NewCoincheckApiClient()

var coincheckApiType = map[string]string{
	"accountBalance":            "/api/accounts/balance",
	"tradeRate":                 "/api/exchange/orders/rate",
	"storeRate":                 "/api/rate",
	"getTransactions":           "/api/exchange/orders/transactions",
	"getTransactionsPagination": "/api/exchange/orders/transactions_pagination",
	"newOrder":                  "/api/exchange/orders",
}

func main() {
	coincheckRequestPath := coincheckApiType["accountBalance"]
	CoincheckApiClient.CallApi(coincheckRequestPath)
}
