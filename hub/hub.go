package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"spider/leg"

	"github.com/gorilla/websocket"
)

// TODO - Consider having a single control channel rather than removeLeg, addLeg etc.

type Hub struct {
	in        chan *leg.Msg
	addLeg    chan *leg.Leg
	removeLeg chan int
	legs      map[int]*leg.Leg
	router    Router
}

type Router func(*Hub, *leg.Msg)

func New(router Router) *Hub {

	hub := &Hub{
		in:        make(chan *leg.Msg),
		addLeg:    make(chan *leg.Leg),
		removeLeg: make(chan int),
		legs:      make(map[int]*leg.Leg),
		router:    router,
	}

	go hub.listen()

	return hub
}

func (hub *Hub) listen() {

	for {
		select {
		case msg := <-hub.in:
			hub.handleUserMsg(msg)
		case leg := <-hub.addLeg:
			// TODO - Extract method.
			hub.legs[leg.Id()] = leg
			go leg.ListenToClient(hub.in, hub.removeLeg)
		case id := <-hub.removeLeg:
			delete(hub.legs, id)
		}
	}
}

func (hub *Hub) handleUserMsg(userMsg *leg.Msg) {
	hub.router(hub, userMsg)
}

// TODO - should hub depend on websocket as well as Leg?
func (hub *Hub) GrowLeg(conn *websocket.Conn) {
	hub.log("adding leg")
	l := leg.NewLeg(conn)
	hub.addLeg <- l
}

func (hub *Hub) log(text interface{}) {
	log.Printf("hub: %v\n", text)
}

// #### ROUTERS #### //

func Broadcast(hub *Hub, msg *leg.Msg) {
	for _, l := range hub.legs {
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

func MailMsg(hub *Hub, msg *leg.Msg) {
	in := &mailMsg{}
	if err := json.Unmarshal(msg.Msg, in); err != nil {
		hub.log(fmt.Sprintf("cannot unmarshal json as MailMsg:\n\t %v\n", err))
		return
	}
	in.From = msg.From
	hub.log(fmt.Sprintf("received message from %v\n", in.From))

	switch in.Type {
	case mBroadcast:
		hub.log("broadcasting message")
		for k, l := range hub.legs {
			if out, err := jsonMsg(
				in.Type, k, in.From, nil, in.Payload); err != nil {
				hub.log(err)
			} else {
				l.SendMsg(out)
			}
		}
	case mSend:
		if l := hub.legs[in.To]; l != nil {
			if out, err := jsonMsg(
				in.Type, in.To, in.From, nil, in.Payload); err != nil {
				hub.log(err)
			} else {
				hub.log(fmt.Sprintf("sending message to %v\n", in.To))
				l.SendMsg(out)
			}
		} else {
			errMsg := "Failed to send message: no client with id " + fmt.Sprint(in.To)
			if out, err := jsonMsg(
				in.Type, in.To, in.From, nil, errMsg); err != nil {
				hub.log(fmt.Sprintf("Failed to encode message: %v\n", errMsg))
			} else {
				hub.legs[in.From].SendMsg([]byte(out))
			}
		}
	case mIds:
		ids := make([]int, 0, len(hub.legs))
		for k, _ := range hub.legs {
			if k == in.From {
				continue
			}
			ids = append(ids, k)
		}
		sort.Ints(ids)

		if out, err := jsonMsg(mIds, msg.From, msg.From, ids, msg.Msg); err != nil {
			hub.log(fmt.Sprintf("Failed to encode id message: %v\n", err))
		} else {
			hub.log(fmt.Sprintf("Sending ids to %v: %v\n", in.From, ids))
			hub.legs[in.From].SendMsg(out)
		}
	}
}

func jsonMsg(typ, to, from int, ids []int, payload interface{}) ([]byte, error) {
	msg := &mailMsg{typ, to, from, ids, payload}
	jmsg, err := json.Marshal(msg)
	return jmsg, err
}
