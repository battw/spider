package parts

import (
	"github.com/gorilla/websocket"
	"log"
)

type LegMsg map[string]interface{}

type leg struct{}

func Leg(ws *websocket.Conn) chan<- LegMsg {
	legChan := make(chan<- LegMsg)
	return legChan
}

// sendMsg converts msg to json and sends it down the leg.
func (l *leg) sendMsg(msg interface{}) {
}

func (l *leg) log(text interface{}) {
	log.Printf("leg: %v\n", text)
}
