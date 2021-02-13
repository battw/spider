package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"spider/leg"

	"github.com/gorilla/websocket"
)

type Hub struct {
	in        chan<- *leg.Msg
	addLeg    chan<- *leg.Leg
	removeLeg chan<- int
	legs      map[int]*leg.Leg
	router    Router
}

type Router func(*Hub, *leg.Msg)

func New(router Router) *Hub {
	in := make(chan *leg.Msg)
	legs := make(map[int]*leg.Leg)
	addLeg := make(chan *leg.Leg)
	removeLeg := make(chan int)
	hub := &Hub{in, addLeg, removeLeg, legs, router}
	go func() {
		for {
			select {
			case msg := <-in:
				hub.router(hub, msg)
			case leg := <-addLeg:
				legs[leg.Id()] = leg
				go leg.ListenToClient(in, removeLeg)
			case id := <-removeLeg:
				delete(legs, id)
			}
		}
	}()
	return hub
}

func (s *Hub) GrowLeg(conn *websocket.Conn) {
	s.log("adding leg")
	l := leg.NewLeg(conn)
	s.addLeg <- l
}

func (s *Hub) log(text interface{}) {
	log.Printf("hub: %v\n", text)
}

// #### ROUTERS #### //

func Broadcast(s *Hub, msg *leg.Msg) {
	for _, l := range s.legs {
		l.SendMsg(msg.Msg)
	}
}

const (
	mSend = iota + 1
	mBroadcast
	mIds
	mError
)

type mailMsg struct {
	Type    int
	To      int
	From    int
	Ids     []int
	Payload interface{}
}

func MailMsg(s *Hub, msg *leg.Msg) {
	in := &mailMsg{}
	if err := json.Unmarshal(msg.Msg, in); err != nil {
		s.log(fmt.Sprintf("cannot unmarshal json as MailMsg:\n\t %v\n", err))
		return
	}
	in.From = msg.From
	s.log(fmt.Sprintf("received message from %v\n", in.From))

	switch in.Type {
	case mBroadcast:
		s.log("broadcasting message")
		for k, l := range s.legs {
			if out, err := jsonMsg(
				in.Type, k, in.From, nil, in.Payload); err != nil {
				s.log(err)
			} else {
				l.SendMsg(out)
			}
		}
	case mSend:
		if l := s.legs[in.To]; l != nil {
			if out, err := jsonMsg(
				in.Type, in.To, in.From, nil, in.Payload); err != nil {
				s.log(err)
			} else {
				s.log(fmt.Sprintf("sending message to %v\n", in.To))
				l.SendMsg(out)
			}
		} else {
			errMsg := "Failed to send message: no client with id " + fmt.Sprint(in.To)
			if out, err := jsonMsg(
				in.Type, in.To, in.From, nil, errMsg); err != nil {
				s.log(fmt.Sprintf("Failed to encode message: %v\n", errMsg))
			} else {
				s.legs[in.From].SendMsg([]byte(out))
			}
		}
	case mIds:
		ids := make([]int, 0, len(s.legs))
		for k, _ := range s.legs {
			if k == in.From {
				continue
			}
			ids = append(ids, k)
		}
		sort.Ints(ids)

		if out, err := jsonMsg(mIds, msg.From, msg.From, ids, msg.Msg); err != nil {
			s.log(fmt.Sprintf("Failed to encode id message: %v\n", err))
		} else {
			s.log(fmt.Sprintf("Sending ids to %v: %v\n", in.From, ids))
			s.legs[in.From].SendMsg(out)
		}
	}
}

func jsonMsg(typ, to, from int, ids []int, payload interface{}) ([]byte, error) {
	msg := &mailMsg{typ, to, from, ids, payload}
	jmsg, err := json.Marshal(msg)
	return jmsg, err
}
