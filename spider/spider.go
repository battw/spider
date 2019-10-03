package spider

import (
	"github.com/battw/spider/leg"
	"github.com/gorilla/websocket"
	"log"
)

type id int

type Spider struct {
	in        chan<- []byte
	addLeg    chan<- *leg.Leg
	removeLeg chan<- int
	legs      map[int]*leg.Leg
	brain     func(map[int]*leg.Leg, []byte)
}

func Hatch() *Spider {
	in := make(chan []byte)
	legs := make(map[int]*leg.Leg)
	addLeg := make(chan *leg.Leg)
	removeLeg := make(chan int)
	s := &Spider{in, addLeg, removeLeg, legs, Broadcast}
	go func() {
		for {
			select {
			case msg := <-in:
				s.brain(legs, msg)
			case leg := <-addLeg:
				legs[leg.Id()] = leg
				go leg.ListenToClient(in, removeLeg)
			case id := <-removeLeg:
				delete(legs, id)
			}
		}
	}()
	return s
}

func (s *Spider) GrowLeg(conn *websocket.Conn) {
	s.log("adding leg")
	l := leg.NewLeg(conn)
	s.addLeg <- l
}

func (s *Spider) log(text interface{}) {
	log.Printf("Spider: %v\n", text)
}

func Broadcast(legs map[int]*leg.Leg, msg []byte) {
	for _, l := range legs {
		l.SendMsg(msg)
	}
}
