package bitly

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Bitlink struct {
	ID       string `json:"id"`
	Link     string `json:"link"`
	LongUrl  string `json:"long_link"`
	Archived bool   `json:"deleted"`
}

const URL = "https://api-ssl.bitly.com/v4/bitlinks"

type BitlyClient struct {
	ApiToken string
}

func NewClient(apiToken string) *BitlyClient {
	var client BitlyClient
	client.ApiToken = apiToken
	return &client
}

func (b *BitlyClient) ShortenUrl(longUrl string) string {
	// TODO: url encode longUrl before adding to bit.ly
	payload := strings.NewReader(`{"long_url": "` + longUrl + `"}`)

	req, err := http.NewRequest("POST", URL, payload)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Add("Authorization", `Bearer `+b.ApiToken)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Panic(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var bitlink Bitlink

	json.Unmarshal(body, &bitlink)

	return bitlink.Link
}
