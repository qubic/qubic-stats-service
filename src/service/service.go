package service

import (
	"fmt"
	"io"
	"net/http"
)

type Service struct {
	CoinGeckoToken string
	WebPort        string

	Database           string
	SpectrumCollection string
}

func (s *Service) runService() {

}

func FetchCoinGeckoPrice() {

	url := "https://api.coingecko.com/api/v3/simple/price?ids=qubic-network&vs_currencies=usd"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", "TOKEN")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

}
