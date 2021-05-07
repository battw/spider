package server

// TODO - CLEAN ME

import (
	"log"
	"net/http"
	"spider/hub"
	"spider/socket"
	"strconv"
)

type Server struct{}

func New(userMsgRouter hub.Router) *Server {
	theHub := hub.New(userMsgRouter)
	http.HandleFunc("/ws", wsConnect(theHub))
	return &Server{}
}

// RegisterRoute takes a URL extension (including leading /)
// and sets a handler to serve the file specified by filepath (without a leading /).
func (server *Server) RegisterRoute(URLExtension string, filepath string) {

	handler := func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, filepath)
	}

	http.HandleFunc(URLExtension, handler)
}

func (server *Server) Run(port int) {
	portString := ":" + strconv.Itoa(port)
	http.ListenAndServe(portString, nil)
}

// wsConnect returns an http request handler which upgrades the connection to a
// websocket and adds it to the hub.
func wsConnect(theHub *hub.Hub) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		socket, err := socket.New(writer, request)

		if err != nil {
			log.Println("Failed to create web socket: " + err.Error())
			return
		}

		log.Printf("new socket: id = %v\n", socket.ID())
		theHub.AddSocket(socket)

	}
}
