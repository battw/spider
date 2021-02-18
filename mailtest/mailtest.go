package main

import (
	"spider/hub"
	"spider/server"
)

// main sets up a server providing a web based chat webSocketAdapter on "/"
func main() {
	server := server.New(hub.HandleMailMsg)
	server.RegisterRoute("/", "mailtest/client.html")
	server.RegisterRoute("/script.mjs", "mailtest/script.mjs")
	// TODO - Should this route be specified in the server?
	server.RegisterRoute("/spiderclient.mjs", "client/client.mjs")
	server.Run(5000)
}
