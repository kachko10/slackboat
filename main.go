package main

import (
	"os"
	"strings"
	"time"
	"fmt"
	"net/http"
	"golang.org/x/net/websocket"
	"encoding/json"
)

type DataSet struct {
	Data StockData `json:"dataset"`
}

type StockData struct {
	ID                  int      `json:"id"`
	DatasetCode         string   `json:"dataset_code"`
	DatabaseCode        string   `json:"database_code"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	RefreshedAt         string   `json:"refreshed_at"`
	NewestAvailableDate string   `json:"newest_available_date"`
	OldestAvailableDate string   `json:"oldest_available_date"`
	ColumnNames         []string `json:"column_names"`
	Frequency           string   `json:"frequency"`
	Type                string   `json:"type"`
	Premium             bool     `json:"premium"`
	Limit               string   `json:"limit"`
	Transform           string   `json:"transform"`
	ColumnIndex         int      `json:"column_index"`
	StartDate           string   `json:"start_date"`
	EndDate             string   `json:"end_date"`
	Changes             []Data   `json:"data"`
	Collapse            string   `json:"collapse"`
	Order               string   `json:"order"`
	DatabaseId          int      `json:"database_id"`
}

type Data struct {
	Date        string
	PercentDiff float64
}

const STOCK_PREFIX = "stock:"

func (d *Data) MarshalJSON() ([]byte, error) {
	arr := []interface{}{d.Date, d.PercentDiff}
	return json.Marshal(arr)
}

func (d *Data) UnmarshalJSON(bs []byte) error {
	arr := []interface{}{}
	err := json.Unmarshal(bs, &arr)
	if err != nil {
		println(err)
		panic(err)
	}
	d.Date = arr[0].(string)
	d.PercentDiff = arr[1].(float64)
	return nil
}

func main() {

	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		panic("no slack slackToken supplied")
	}

	quandlToken:=os.Getenv("QUANDL_TOKEN")
	if quandlToken == "" {
		panic("no quandlToken  supplied")
	}

	ws, slackBootId := slackConnect(slackToken)
	for {
		msg, err := getMessage(ws)

		if err != nil {
			println(err)
		}

		replay(msg, slackBootId,quandlToken, ws,)
	}
}

func replay(m Message, slackBootId string,quandlToken string, ws *websocket.Conn) {

	if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+slackBootId+">") {
		//stock quote should be prefixed stock:"
		if strings.Contains(m.Text, STOCK_PREFIX) {
			go func(m Message, ws *websocket.Conn) {
				getStockData(m, ws,quandlToken)
			}(m,ws)
		} else {
			m.Text = "Dude i can only answer about stocks "
			postMessage(ws, m)
		}

	}
}

//https://www.quandl.com/api/v3/datasets/WIKI/FB.json?column_index=4&start_date=2014-01-01&end_date=2014-12-31&collapse=monthly&transform=rdiff&api_key=some_key
func getStockData(m Message, ws *websocket.Conn,quandlToken string) {

	stockQuote := strings.SplitAfter(m.Text, STOCK_PREFIX)[1]
	stockQuote = strings.TrimSpace(stockQuote)
	current_time := time.Now().Local()
	last_year := current_time.AddDate(0, -1, 0)

	url := fmt.Sprintf("https://www.quandl.com/api/v3/datasets/WIKI/%s.json?column_index=4&start_date=%s&end_date=%s&collapse=monthly&transform=none&api_key=%s",
		stockQuote, last_year.Format("2006-01-02"),
		current_time.Format("2006-01-02"), quandlToken)

	resp, err := http.Get(url)
	if err != nil {
		println(err)
		panic(err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotModified {
		var stackData DataSet
		err = json.NewDecoder(resp.Body).Decode(&stackData)
		if err != nil {
			println(err)
			panic(err)
		}
		m.Text = fmt.Sprintf("Current Price %.3f Price for last month :%s  was : %.3f",
			stackData.Data.Changes[0].PercentDiff, stackData.Data.Changes[1].Date,
			stackData.Data.Changes[1].PercentDiff)
	} else {
		m.Text = fmt.Sprintf("Quote %s does not exist", stockQuote)
	}
	defer resp.Body.Close()
	postMessage(ws, m)
}
