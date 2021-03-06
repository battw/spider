package hub

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/battw/spider/socket"
)

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
		hub.sockets[socket.ID()] = socket
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
	hub.logSocketAdd(socket)
	hub.addSocketChan <- socket
}

func (hub *Hub) logSocketAdd(socket *socket.Socket) {
	hub.log("adding socket " + strconv.Itoa(socket.ID()))
}

func (hub *Hub) GetSocket(id int) (*socket.Socket, error) {

	socket := hub.sockets[id]

	var err error = nil
	if socket == nil {
		err = hub.newInvalidIDError(id)
		hub.log(err.Error())
	}

	return socket, err
}

func (hub *Hub) newInvalidIDError(socketID int) error {
	return fmt.Errorf("socket %v does not exist", socketID)
}

func (hub *Hub) getSocketIDs() []int {

	IDs := make([]int, 0, len(hub.sockets))

	for ID := range hub.sockets {
		IDs = append(IDs, ID)
	}

	sort.Ints(IDs)

	hub.logGetIDs(IDs)

	return IDs
}

func (hub *Hub) logGetIDs(IDs []int) {
	hub.log(fmt.Sprintf("getting socket IDs: %v", IDs))
}

func (hub *Hub) log(text interface{}) {
	log.Printf("hub: %v\n", text)
}

// TODO move router somewhere else by creating a proper API for Hub.

type msgType int

const (
	sendType msgType = iota + 1
	broadcastType
	fetchIdsType
	errorType
)

type mailMsg struct {
	MsgType       msgType
	DestinationID int
	SenderID      int
	IDs           []int
	Payload       interface{}
}

type handleFuncType = func(*Hub, *mailMsg)

var handleFuncMap = map[msgType]handleFuncType{
	sendType:      handleSend,
	broadcastType: handleBroadcast,
	fetchIdsType:  handleIDs,
}

func (msg *mailMsg) String() string {
	return fmt.Sprintf("mailMsg {MsgType: %v, DestinationID: %v, SenderID: %v, Ids: %v, Payload: %v}",
		msg.MsgType, msg.DestinationID, msg.SenderID, msg.IDs, msg.Payload)
}

func HandleMailMsg(hub *Hub, incomingMsg *socket.UserMsg) {

	msg, err := unpackMailMsg(hub, incomingMsg)
	if err != nil {
		sendErrorMsg(hub, msg.SenderID, "failed to unpack message.")
		return
	}

	routeMailMsg(hub, msg)
}

func routeMailMsg(hub *Hub, msg *mailMsg) {
	logMsgReceived(hub, msg)
	handleFuncMap[msg.MsgType](hub, msg)
}

func logMsgReceived(hub *Hub, msg *mailMsg) {
	hub.log(fmt.Sprintf("received message from %v: %v", msg.SenderID, msg))
}

func unpackMailMsg(hub *Hub, msg *socket.UserMsg) (*mailMsg, error) {

	unpackedMsg := &mailMsg{}
	unpackedMsg.SenderID = msg.SenderID

	/* TODO - message might not return an error but be wrong e.g. have 0 == senderID == destinationID.
	This should be checked */
	err := json.Unmarshal(msg.Msg, unpackedMsg)
	if err != nil {
		logUnpackError(hub, err)
	}

	return unpackedMsg, err
}

func logUnpackError(hub *Hub, err error) {
	hub.log(fmt.Sprintf("cannot unmarshal json as MailMsg: %v", err))
}

func handleBroadcast(hub *Hub, msg *mailMsg) {

	logMsgBroadcast(hub, msg)

	for _, ID := range hub.getSocketIDs() {

		msg := &mailMsg{
			MsgType:       broadcastType,
			DestinationID: ID,
			SenderID:      msg.SenderID,
			Payload:       msg.Payload,
		}

		sendMsg(hub, msg)
	}
}

func logMsgBroadcast(hub *Hub, msg *mailMsg) {
	hub.log(fmt.Sprintf("broadcasting message: %v", msg))
}

func handleSend(hub *Hub, msg *mailMsg) {
	err := sendMsg(hub, msg)
	if err != nil {
		sendErrorMsg(hub, msg.SenderID, "failed to send message.")
	}
}

func sendErrorMsg(hub *Hub, destinationID int, errorString string) {
	errorMsg := newErrorMsg(destinationID, errorString)
	sendMsg(hub, errorMsg)
}

func newErrorMsg(destinationID int, errorString string) *mailMsg {
	return &mailMsg{
		MsgType:       errorType,
		DestinationID: destinationID,
		Payload:       errorString,
	}
}

func sendMsg(hub *Hub, msg *mailMsg) error {

	packedMsg, err := packMailMsg(msg)
	if err != nil {
		return err
	}

	socket, err := hub.GetSocket(msg.DestinationID)
	if err != nil {
		return err
	}

	socket.SendMsg(packedMsg)
	return nil
}

func packMailMsg(msg *mailMsg) ([]byte, error) {

	json, err := json.Marshal(msg)
	if err != nil {
		logPackingError(msg, err)
	}

	return json, err
}

func logPackingError(msg *mailMsg, err error) {
	log.Printf("failed to convert mailMsg to JSON: (%v), (%v)", msg, err)
}

func handleIDs(hub *Hub, msg *mailMsg) {

	var IDs []int = hub.getSocketIDs()

	IDs = removeX(IDs, msg.SenderID)

	IDMsg := &mailMsg{
		MsgType:       fetchIdsType,
		DestinationID: msg.SenderID,
		SenderID:      msg.SenderID,
		IDs:           IDs,
	}

	sendMsg(hub, IDMsg)
}

func removeX(slice []int, x int) []int {

	var xIndex = -1

	for i := range slice {
		if slice[i] == x {
			xIndex = i
		}
	}

	if xIndex != -1 {
		slice = append(slice[:xIndex], slice[xIndex+1:]...)
	}

	return slice
}
