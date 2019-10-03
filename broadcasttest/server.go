package main

import (
	"github.com/battw/spider/spider"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// main sets up a server providing a web based chat webSocketAdapter on "/"
func main() {
	s := spider.Hatch(spider.Broadcast)
	http.HandleFunc("/", servePage)
	http.HandleFunc("/ws", wsConnect(s))
	http.ListenAndServe(":5000", nil)
}

// wsConnect returns an http request handler which upgrades the connection to a
// websocket and adds it to the spider.
func wsConnect(s *spider.Spider) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		//upgrade connection
		var upgrader = websocket.Upgrader{
			HandshakeTimeout:  0,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			Subprotocols:      nil,
			Error:             nil,
			CheckOrigin:       nil,
			EnableCompression: false,
		}

		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		s.GrowLeg(conn)
	}
}

// servePage is an http request handler which serves the webchat web page to a
// webSocketAdapter.
func servePage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "broadcasttest/client.html")
}
