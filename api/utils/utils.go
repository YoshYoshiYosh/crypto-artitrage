package utils

import (
	"os/exec"
)

func Scraping() string {
	// 吉田に書いてもらう
	scrapingRateResultByte, _ := exec.Command("ruby", "./api/utils/scraping.rb").Output()

	rate := string(scrapingRateResultByte)
	return rate
}
