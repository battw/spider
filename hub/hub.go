package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"spider/socket"
)

// TODO - Consider having a single control channel rather than removeSocketChan, addSocketChan etc.

type Hub struct {
	userMsgChan      chan *socket.UserMsg
	addSocketChan    chan *socket.Socket
	removeSocketChan chan int
	sockets          map[int]*socket.Socket
	routeUserMsg     Router
}

type Router func(*Hub, *socket.UserMsg)

func New(router Router) *Hub {

	hub := &Hub{
		userMsgChan:      make(chan *socket.UserMsg),
		addSocketChan:    make(chan *socket.Socket),
		removeSocketChan: make(chan int),
		sockets:          make(map[int]*socket.Socket),
		routeUserMsg:     router,
	}

	go hub.handleIncoming()

	return hub
}

func (hub *Hub) handleIncoming() {

	handleUserMsg := func(userMsg *socket.UserMsg) {
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

func (hub *Hub) AddSocket(socket *socket.Socket) {
	hub.log("adding socket " + strconv.Itoa(socket.Id()))
	hub.addSocketChan <- socket
}

func (hub *Hub) log(text interface{}) {
	log.Printf("hub: %v\n", text)
}

// TODO move routers somewhere else.
// #### ROUTERS #### //
func BroadcastMsg(hub *Hub, msg *socket.UserMsg) {
	for _, l := range hub.sockets {
		l.SendMsg(msg.Msg)
	}
}

type msgType int

const (
	send msgType = iota + 1
	broadcast
	fetchIds
)

type mailMsg struct {
	MsgType       msgType
	DestinationID int
	SenderID      int
	Ids           []int
	Payload       interface{}
}

type handleFuncType = func(*Hub, *mailMsg)

var handleFuncMap = map[msgType]handleFuncType{
	send:      sendMsg,
	broadcast: broadcastMsg,
	fetchIds:  sendIds,
}

func (msg *mailMsg) String() string {
	return fmt.Sprintf("mailMsg {MsgType: %v, DestinationID: %v, SenderID: %v, Ids: %v, Payload: %v}",
		msg.MsgType, msg.DestinationID, msg.SenderID, msg.Ids, msg.Payload)
}

//TODO - Should the broadcast message a special case of send? Should send take a list of destinationIds instead?
func RouteMailMsg(hub *Hub, msg *socket.UserMsg) {
	incomingMsg := &mailMsg{}
	if err := json.Unmarshal(msg.Msg, incomingMsg); err != nil {
		hub.log(fmt.Sprintf("cannot unmarshal json as MailMsg: %v", err))
		return
	}
	incomingMsg.SenderID = msg.SenderID
	hub.log(fmt.Sprintf("received message from %v", incomingMsg.SenderID))
	hub.log(fmt.Sprintf("message as string: %v", string(msg.Msg)))
	hub.log(fmt.Sprintf("message as mailMsg: %v", incomingMsg))

	handleFuncMap[incomingMsg.MsgType](hub, incomingMsg)
}

func broadcastMsg(hub *Hub, msg *mailMsg) {
	hub.log("broadcasting message")
	for k, l := range hub.sockets {
		if out, err := jsonMsg(
			msg.MsgType, k, msg.SenderID, nil, msg.Payload); err != nil {
			hub.log(err)
		} else {
			l.SendMsg(out)
		}
	}
}

func sendMsg(hub *Hub, msg *mailMsg) {
	if recieverSocket := hub.sockets[msg.DestinationID]; recieverSocket != nil {
		if out, err := jsonMsg(
			msg.MsgType, msg.DestinationID, msg.SenderID, nil, msg.Payload); err != nil {
			hub.log(err)
		} else {
			hub.log(fmt.Sprintf("sending message to %v", msg.DestinationID))
			recieverSocket.SendMsg(out)
		}
	} else {
		errMsg := "Failed to send message: no client with id " + fmt.Sprint(msg.DestinationID)
		if out, err := jsonMsg(
			msg.MsgType, msg.DestinationID, msg.SenderID, nil, errMsg); err != nil {
			hub.log(fmt.Sprintf("Failed to encode message: %v", errMsg))
		} else {
			hub.sockets[msg.SenderID].SendMsg([]byte(out))
		}
	}
}

func sendIds(hub *Hub, msg *mailMsg) {
	ids := make([]int, 0, len(hub.sockets))
	for k := range hub.sockets {
		if k == msg.SenderID {
			continue
		}
		ids = append(ids, k)
	}
	sort.Ints(ids)

	if out, err := jsonMsg(fetchIds, msg.SenderID, msg.SenderID, ids, nil); err != nil {
		hub.log(fmt.Sprintf("Failed to encode id message: %v", err))
	} else {
		hub.log(fmt.Sprintf("Sending ids to %v: %v", msg.SenderID, ids))
		hub.sockets[msg.SenderID].SendMsg(out)
	}
}

func jsonMsg(typ msgType, to, from int, ids []int, payload interface{}) ([]byte, error) {
	msg := &mailMsg{typ, to, from, ids, payload}
	jmsg, err := json.Marshal(msg)
	return jmsg, err
}
