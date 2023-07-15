package hackernews

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type HNItem struct {
	ID        int    `json:"id"`
	Parent    int    `json:"parent"`
	Kids      []int  `json:"kids"`
	Parts     []int  `json:"parts"`
	Score     int    `json:"score"`
	Timestamp int    `json:"time"`
	By        string `json:"by"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	URL       string `json:"url"`
	Dead      bool   `json:"dead"`
	Deleted   bool   `json:"deleted"`
}

type HNClient struct {
	BaseUrl string
}

func NewHNClient() *HNClient {
	var c HNClient
	c.BaseUrl = "https://hacker-news.firebaseio.com/v0/"
	return &c
}

func (c *HNClient) GetItemIDs() []int {
	var top500 []int
	url := c.BaseUrl + "topstories.json?print=pretty"

	response, err := c.makeRequest(url)
	if err != nil {
		return []int{}
	}
	json.Unmarshal(response, &top500)

	return top500
}

func (c *HNClient) GetItem(itemID int) (*HNItem, error) {
	var item HNItem
	log.Printf("Getting item %v\n", itemID)
	url := c.BaseUrl + "item/" + strconv.Itoa(itemID) + ".json?print=pretty"

	response, err := c.makeRequest(url)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(response, &item)
	return &item, nil
}

func (c *HNClient) makeRequest(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err
}
