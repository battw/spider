package socket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Socket struct {
	id   int
	conn *websocket.Conn
}

type UserMsg struct {
	SenderID int
	Msg      []byte
}

// idChan outputs unique ids in a concurrency safe fashion
var idChan <-chan int = func() <-chan int {
	ch := make(chan int)
	go func() {
		// id starts from 1 so sockets constructed using New can be
		// distinguished from uninitialised ones (which would have id=0)
		for id := 1; ; id++ {
			ch <- id
		}
	}()
	return ch
}()

func New(writer http.ResponseWriter, request *http.Request) (*Socket, error) {

	conn, err := upgradeToWebsocket(writer, request)
	id := <-idChan

	return &Socket{id, conn}, err
}

func upgradeToWebsocket(writer http.ResponseWriter, request *http.Request) (*websocket.Conn, error) {
	// TODO - Move the upgrader config somewhere settable
	var upgrader = websocket.Upgrader{
		HandshakeTimeout:  0,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		WriteBufferPool:   nil,
		Subprotocols:      nil,
		Error:             nil,
		CheckOrigin:       nil,
		EnableCompression: false,
	}

	return upgrader.Upgrade(writer, request, nil)
}

func (sock *Socket) Id() int {
	return sock.id
}

// listenToClient forwards any messages received from the client into "dest"
func (sock *Socket) ListenToClient(destination chan<- *UserMsg, removeLeg chan<- int) {
	for {
		// ReadMessage returns test or binary messages only.
		// close, ping and pong are handled elsewhere.
		_, msg, err := sock.conn.ReadMessage()
		sock.log("received message from client")
		if err != nil {
			sock.log("message read from websocket returned error: " + err.Error())
			removeLeg <- sock.id
			break
		}
		destination <- &UserMsg{sock.id, msg}
	}
}

// sendMsg converts msg to json and sends it down the leg.
func (sock *Socket) SendMsg(m []byte) {
	sock.log("sending message to client")
	sock.conn.WriteMessage(websocket.TextMessage, []byte(m))
}

func (sock *Socket) log(text interface{}) {
	log.Printf("socket %v: %v\n", sock.id, text)
}
