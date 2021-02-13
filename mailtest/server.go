package main

import (
	"log"
	"net/http"
	"spider/hub"

	"github.com/gorilla/websocket"
)

// main sets up a server providing a web based chat webSocketAdapter on "/"
func main() {
	s := hub.New(hub.MailMsg)
	http.HandleFunc("/", servePage)
	http.HandleFunc("/script.mjs", serveScript)
	http.HandleFunc("/foot.mjs", serveFoot)
	http.HandleFunc("/ws", wsConnect(s))
	http.ListenAndServe(":5000", nil)
}

// wsConnect returns an http request handler which upgrades the connection to a
// websocket and adds it to the hub.
func wsConnect(s *hub.Hub) func(http.ResponseWriter, *http.Request) {
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

func serveScript(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "mailtest/script.mjs")
}

func serveFoot(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "client/client.mjs")
}

// servePage is an http request handler which serves the webchat web page to a
// webSocketAdapter.
func servePage(writer http.ResponseWriter, request *http.Request) {
	http.ServeFile(writer, request, "mailtest/client.html")
}
