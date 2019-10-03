package spider

import (
	"encoding/json"
	"fmt"
	"github.com/battw/spider/leg"
	"github.com/gorilla/websocket"
	"log"
)

type id int

type Spider struct {
	in        chan<- *leg.Msg
	addLeg    chan<- *leg.Leg
	removeLeg chan<- int
	legs      map[int]*leg.Leg
	brain     Brain
}

type Brain func(*Spider, *leg.Msg)

func Hatch(b Brain) *Spider {
	in := make(chan *leg.Msg)
	legs := make(map[int]*leg.Leg)
	addLeg := make(chan *leg.Leg)
	removeLeg := make(chan int)
	s := &Spider{in, addLeg, removeLeg, legs, b}
	go func() {
		for {
			select {
			case msg := <-in:
				s.brain(s, msg)
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

// #### BRAINS #### //

func Broadcast(s *Spider, msg *leg.Msg) {
	for _, l := range s.legs {
		l.SendMsg(msg.Msg)
	}
}

type mailMsg struct {
	to      int
	from    int
	payload interface{}
}

func MailMsg(s *Spider, msg *leg.Msg) {
	var mm mailMsg
	if err := json.Unmarshal(msg.Msg, &mm); err != nil {
		s.log(fmt.Sprintf("cannot unmarshal json as MailMsg:\n\t %v\n", err))
		return
	}
	switch {
	case mm.to == 0: // broadcast message
		for _, l := range s.legs {
			l.SendMsg(msg.Msg)
		}
	case mm.to > 0: // send to addressee
		if l := s.legs[mm.to]; l == nil {
			l.SendMsg(msg.Msg)
		} else {
			s.log(fmt.Sprintf(
				"Tried to send message to non-existant address:\n\t %v\n",
				mm.to))
			return
		}
	}
}
