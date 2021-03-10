package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Scraping() float64 {
	// 吉田に書いてもらう
	scrapingRateResultByte, _ := exec.Command("ruby", "./api/utils/scraping.rb").Output()

	rate, _ := strconv.ParseFloat(string(scrapingRateResultByte), 64)
	return rate
}

func OutputAllRatesToLog(rateOfExchangeList []map[string]float64, expectProfit int) {
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
