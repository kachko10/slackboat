package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"golang.org/x/net/websocket"
	"sync/atomic"
)

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	Id string `json:"id"`
}

type Message struct {
	Id      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

var counter uint64

func getMessage(ws *websocket.Conn) (m Message, err error) {
	err = websocket.JSON.Receive(ws, &m)
	return
}

func postMessage(ws *websocket.Conn, m Message) error {

	m.Id = atomic.AddUint64(&counter, 1)
	return websocket.JSON.Send(ws, m)
}


func slackConnect(token string) (*websocket.Conn, string) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)
	resp, err := http.Get(url)
	if err != nil {
		println(err)
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println(err)
		panic(err)
	}
	var respObj responseRtmStart
	err = json.Unmarshal(body, &respObj)

	id := respObj.Self.Id
	ws, err := websocket.Dial(respObj.Url, "", "https://api.slack.com/")
	if err != nil {
		println(err)
		panic(err)
	}
	return ws, id

}
