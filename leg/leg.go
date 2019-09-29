package leg

import (
	"github.com/gorilla/websocket"
	"log"
)

type LegMsg = string

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
func (l *Leg) ListenToClient(dest chan<- LegMsg) {
	for {
		typ, msg, err := l.conn.ReadMessage()
		l.log("received message from client")
		if err != nil {
			log.Printf("failed message read from websocket: %v\n",
				err)
			continue
		}
		if typ != websocket.TextMessage {
			log.Printf(
				"message read from websocket is of wrong type: %v\n",
				typ)
			continue
		}
		dest <- string(msg)
	}
}

// sendMsg converts msg to json and sends it down the leg.
func (l *Leg) SendMsg(m LegMsg) {
	l.log("sending message")
	l.conn.WriteMessage(websocket.TextMessage, []byte(m))
}

func (l *Leg) log(text interface{}) {
	log.Printf("leg %v: %v\n", l.id, text)
}
