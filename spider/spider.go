package spider

import (
	"github.com/battw/spider/leg"
	"github.com/gorilla/websocket"
	"log"
)

type id int

type Spider struct {
	in     chan<- leg.LegMsg
	addLeg chan<- *leg.Leg
}

func Hatch() *Spider {
	in := make(chan leg.LegMsg)
	legs := make(map[int]*leg.Leg)
	addLeg := make(chan *leg.Leg)

	go func() {
		for {
			select {
			case msg := <-in:
				for _, l := range legs {
					l.SendMsg(msg)
				}
			case leg := <-addLeg:
				legs[leg.Id()] = leg
				go leg.ListenToClient(in)
			}
		}
	}()
	return &Spider{in, addLeg}
}

func (s *Spider) GrowLeg(conn *websocket.Conn) {
	s.log("adding leg")
	l := leg.NewLeg(conn)
	s.addLeg <- l
}

func (s *Spider) log(text interface{}) {
	log.Printf("Spider: %v\n", text)
}
