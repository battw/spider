// TODO - user should register a handleBroadcastMsg function
// TODO - only the payload of a message should be passed to the user callbacks

/**
  Creates a SpiderClient object, which is the client end-point for a browser websocket 
  network. 
  Send messages with client.send()
  Get a list of the ids of the other connected clients with getIds.
  You will need to register a handler for the response using regIdHandler(f).
  Register a handling function for incoming messages with regMsgHandler(f)
**/
export function SpiderClient() {
    if (!window["WebSocket"]) {
        console.error("Browser doesn't support websockets. Exiting Spider.")
        return null;
    }

    let conn = new WebSocket("ws://" + document.location.host + "/ws")
    // These are set by the user for a specific application
    let IDMsgHandler = null
    let peerMsgHandler = null
    let broadcastMsgHandler = null
    let errorMsgHandler = null


    const msgType = {
        send: 1,
        broadcast: 2,
        getIDs: 3,
        error: 4,
    }

    let msgHandler = new Map([
        [msgType.send, handlePeerMsg],
        [msgType.broadcast, handleBroadcastMsg],
        [msgType.getIDs, handleIDMsg],
        [msgType.error, handleErrorMsg],
    ])


    conn.onopen = () => {
        console.log("opened websocket")
    }

    conn.onclose = () => {
        console.log("closed websocket")
    }

    conn.onmessage = event => {
        const msg = JSON.parse(event.data)

        console.log(`message received: ${msg}`)

        msgHandler.get(msg.MsgType)(msg)
    }

    function handlePeerMsg(msg) {
        if (peerMsgHandler !== null) {
            peerMsgHandler(msg.Payload, msg.SenderID)
        } else {
            console.error("message received but msgHandler hasn't been registered")
        }
    }

    function handleBroadcastMsg(msg) {
        if (broadcastMsgHandler !== null) {
            broadcastMsgHandler(msg.Payload, msg.SenderID)
        } else {
            console.error("broadcast message received but broadcastMsgHandler hasn't been registered")
        }
    }

    function handleErrorMsg(msg) {
        if (errorMsgHandler !== null) {
            errorMsgHandler(msg.Payload)
        } else {
            console.error("error message received but errorMsgHandler hasn't been registered")
        }
    }

    function handleIDMsg(msg) {
        if (IDMsgHandler !== null) {
            IDMsgHandler(msg.IDs)
        } else {
            console.error("IDs received but IDHandler hasn't been registered")
        }
    }
    
    /** send a message to the client whose id is "to". 
        "msg" is whatever you want **/
    function send(msg, destinationID) {
        const packedMsg = { MsgType: msgType.send, DestinationID: destinationID, Payload: msg }
        conn.send(JSON.stringify(packedMsg))
    }

    /** send "msg" to all clients connected to the network. 
        "msg" is whatever you want **/
    function broadcast(msg) {
        const packedMsg = { MsgType: msgType.broadcast, Payload: msg }
        conn.send(JSON.stringify(packedMsg))
    }

    /** request a list of the current client ids from the server. 
        to handle to reply call regIdhandler(f) with an appropriate function. **/
    function requestIDs() {
        const msg = { MsgType: msgType.getIDs }
        conn.send(JSON.stringify(msg))
    }

    /** "handler" should be a function which accepts an array of numbers **/
    function regIDMsgHandler(handler) {
        IDMsgHandler = handler
    }

    /** "handler" should be a function which accepts an object (the object 
        type is the same as the messages you are sending with SpiderClient.send(msg)) **/
    function regPeerMsgHandler(handler) {
        peerMsgHandler = handler
    }

    function regBroadcastMsgHandler(handler) {
        broadcastMsgHandler = handler
    }

    function regErrorMsgHandler(handler) {
        errorMsgHandler = handler
    }

    return {
        send: send,
        broadcast: broadcast,
        requestIDs: requestIDs,
        regIDMsgHandler: regIDMsgHandler,
        regPeerMsgHandler: regPeerMsgHandler,
        regBroadcastMsgHandler: regBroadcastMsgHandler,
        regErrorMsgHandler: regErrorMsgHandler,
    }
}














