package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"spider/socket"

	"github.com/gorilla/websocket"
)

// TODO - Consider having a single control channel rather than removeSocketChan, addSocketChan etc.

type Hub struct {
	userMsgChan      chan *socket.Msg
	addSocketChan    chan *socket.Socket
	removeSocketChan chan int
	sockets          map[int]*socket.Socket
	routeUserMsg     Router
}

type Router func(*Hub, *socket.Msg)

func New(router Router) *Hub {

	hub := &Hub{
		userMsgChan:      make(chan *socket.Msg),
		addSocketChan:    make(chan *socket.Socket),
		removeSocketChan: make(chan int),
		sockets:          make(map[int]*socket.Socket),
		routeUserMsg:     router,
	}

	go hub.handleIncoming()

	return hub
}

func (hub *Hub) handleIncoming() {

	handleUserMsg := func(userMsg *socket.Msg) {
		hub.routeUserMsg(hub, userMsg)
	}

	addSocket := func(socket *socket.Socket) {
		hub.sockets[socket.Id()] = socket
		go socket.ListenToClient(hub.userMsgChan, hub.removeSocketChan)
	}

	deleteSocket := func(id int) {
		delete(hub.sockets, id)
	}

	for {
		select {
		case msg := <-hub.userMsgChan:
			handleUserMsg(msg)
		case socket := <-hub.addSocketChan:
			addSocket(socket)
		case id := <-hub.removeSocketChan:
			deleteSocket(id)
		}
	}
}

// TODO - Extract this into external API.
func (hub *Hub) AddSocket(conn *websocket.Conn) {
	hub.log("adding leg")
	l := socket.NewLeg(conn)
	hub.addSocketChan <- l
}

func (hub *Hub) log(text interface{}) {
	log.Printf("hub: %v\n", text)
}

// TODO move routers somewhere else.
// #### ROUTERS #### //

func Broadcast(hub *Hub, msg *socket.Msg) {
	for _, l := range hub.sockets {
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

// TODO - WTF is this horrible mess
func MailMsg(hub *Hub, msg *socket.Msg) {
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
		for k, l := range hub.sockets {
			if out, err := jsonMsg(
				in.Type, k, in.From, nil, in.Payload); err != nil {
				hub.log(err)
			} else {
				l.SendMsg(out)
			}
		}
	case mSend:
		if l := hub.sockets[in.To]; l != nil {
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
				hub.sockets[in.From].SendMsg([]byte(out))
			}
		}
	case mIds:
		ids := make([]int, 0, len(hub.sockets))
		for k, _ := range hub.sockets {
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
			hub.sockets[in.From].SendMsg(out)
		}
	}
}

func jsonMsg(typ, to, from int, ids []int, payload interface{}) ([]byte, error) {
	msg := &mailMsg{typ, to, from, ids, payload}
	jmsg, err := json.Marshal(msg)
	return jmsg, err
}
