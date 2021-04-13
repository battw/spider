package socket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Socket is a wrapper around a websocket connection to a client.
type Socket struct {
	id   int
	conn *websocket.Conn
}

// UserMsg contains a message as raw bytes and the socket ID of the sender.
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

// New constructs a socket object with a unique ID.
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

// ID is the identification number for this socket.
func (sock *Socket) ID() int {
	return sock.id
}

// ListenToClient forwards any messages received from the client into the "destination" channel.
func (sock *Socket) ListenToClient(destination chan<- *UserMsg, removeSocket chan<- int) {
	for {
		// ReadMessage returns test or binary messages only.
		// close, ping and pong are handled elsewhere.
		_, msg, err := sock.conn.ReadMessage()
		sock.logMsgReceived()
		if err != nil {
			sock.logWebsocketError(err)
			removeSocket <- sock.id
			break
		}
		destination <- &UserMsg{sock.id, msg}
	}
}

func (sock *Socket) logMsgReceived() {
	sock.log("received message from client")
}

func (sock *Socket) logWebsocketError(err error) {
	sock.log("message read from websocket returned error: " + err.Error())
}

func (sock *Socket) SendMsg(msg []byte) {
	sock.logMsgSend(msg)
	sock.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (sock *Socket) logMsgSend(msg []byte) {
	sock.log(fmt.Sprintf("sending message to client: %v", string(msg)))
}

func (sock *Socket) log(text interface{}) {
	log.Printf("socket %v: %v\n", sock.id, text)
}
