package leg

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type Leg struct {
	id   int
	conn *websocket.Conn
}

// idChan outputs unique ids in a concurrency safe fashion
var idChan <-chan int = func() <-chan int {
	ch := make(chan int)
	go func() {
		// id starts from 1 so legs constructed using NewLeg can be
		// distinguished from uninitialised ones (which would have id=0)
		for id := 1; ; id++ {
			ch <- id
		}
	}()
	return ch
}()

// NOTE if toFoot is not read from, it will block its input!
func NewLeg(conn *websocket.Conn) *Leg {
	id := <-idChan
	log.Printf("newLeg: id = %v\n", id)
	return &Leg{id, conn}
}

func (l *Leg) Id() int {
	return l.id
}

// listenToClient forwards any messages received from the client into "dest"
func (l *Leg) ListenToClient(dest chan<- []byte, removeLeg chan<- int) {
	// REMOVE LEG FROM SPIDER
	for {
		// ReadMessage returns test or binary messages only.
		// close, ping and pong are handled elsewhere.
		_, msg, err := l.conn.ReadMessage()
		l.log("received message from client")
		if err != nil {
			l.log(fmt.Sprintf(
				"message read from websocket returned error: %v\n",
				err))
			removeLeg <- l.id
			break
		}
		dest <- msg
	}
}

// sendMsg converts msg to json and sends it down the leg.
func (l *Leg) SendMsg(m []byte) {
	l.log("sending message to client")
	l.conn.WriteMessage(websocket.TextMessage, []byte(m))
}

func (l *Leg) log(text interface{}) {
	log.Printf("leg %v: %v\n", l.id, text)
}
