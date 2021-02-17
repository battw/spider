package main

import (
	"spider/hub"
	"spider/server"
)

// main sets up a server providing a web based chat webSocketAdapter on "/"
func main() {
	server := server.New(hub.BroadcastMsg)
	server.RegisterRoute("/", "broadcasttest/client.html")
	server.Run(5000)
}
