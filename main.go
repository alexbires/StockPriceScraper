package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	t := handleFlags()
	if verifyTicker(t) {
		fmt.Println("ticker valid")
	}
	fmt.Println(t)

	pageContents := getPriceInformation(t)
	price := findPriceFromHTML(pageContents, t)
	fmt.Printf("The price of %s is: %f", t, price)
}

func verifyTicker(ticker string) bool {
	/*
		We need to ensure that people are giving us a
		valid ticker (based on format)
	*/
	match, _ := regexp.Match(`\w{2,6}`, []byte(ticker))
	if match {
		return true
	}
	return false
}

func handleFlags() string {
	/*
		We need to get information from the user about which stock
		they want to scrape
	*/
	ticker := flag.String("ticker", "", "The ticker to look up")
	flag.Parse()
	return *ticker
}

func setupTLS() *tls.Config {
	/*
		Sets up TLS with reasonably strong parameters
	*/

	config := tls.Config{
		InsecureSkipVerify: false,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		MinVersion: tls.VersionTLS12,
	}
	return &config
}

func getPriceInformation(ticker string) string {
	/*
		Retrieves the price information from yahoo

		Returns:
			String of the html response
	*/
	templateUrl := "https://finance.yahoo.com/quote/%s?p=%s&.tsrc=fin-srch"
	finishedUrl := fmt.Sprintf(templateUrl, ticker, ticker)

	transport := &http.Transport{
		TLSClientConfig: setupTLS(),
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	response, err := httpClient.Get(finishedUrl)
	if err != nil {
		fmt.Printf("There was an error: %s", err)
	}

	defer response.Body.Close()

	content, _ := ioutil.ReadAll(response.Body)
	return string(content)
}

func findPriceFromHTML(response string, ticker string) float64 {
	/*
		Processes the HTML and returns the price information

		Return:
			float64 the price of the stock
	*/
	price := 0.0
	doc, err := goquery.NewDocumentFromReader(strings.NewReader((response)))
	if err != nil {
		fmt.Printf("Error creating the document %s", err)
	}
	selectorTemplate := "fin-streamer[data-symbol='%s']"
	selector := fmt.Sprintf(selectorTemplate, ticker)
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		value, _ := s.Attr("data-field")
		if value == "regularMarketPrice" {
			priceString, _ := s.Attr("value")
			price, _ = strconv.ParseFloat(priceString, 64)
			fmt.Println(s.Attr("value"))
		}

	})
	return price
}
