// TODO - user should register a handleBroadcastMsg function
// TODO - only the payload of a message should be passed to the user callbacks

/**
  Creates a SpiderClient object, which is the client end-point for an n-to-n browser messaging
  network. 
  
  How to Use
  1. Register handlers for incoming messages using the "register(ID|Peer|Broadcast|Error)+MsgHandler" functions.
  2. Request the IDs of the other clients on the network using "requestIds".
  3. Send a message using "send".
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

    // Register websocket callbacks:

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

    // Handlers for incoming messages:

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
    
    // User functions:

    /** send a message to the client with the given id. 
        "msg" is anything that can be serialised using JSON.stringify(). **/
    function send(msg, destinationID) {
        const packedMsg = { MsgType: msgType.send, DestinationID: destinationID, Payload: msg }
        conn.send(JSON.stringify(packedMsg))
    }

    /** send "msg" to all clients connected to the network. 
        "msg" is anything that can be serialised using JSON.stringify(). **/
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

    /** "handler" should be a function which accepts an array of numbers representing the
     *  IDs of the other clients on the network. **/
    function registerIDMsgHandler(handler) {
        IDMsgHandler = handler
    }

    /** "handler" should be a function which accepts an object (the object 
        type is the same as the messages you are sending with SpiderClient.send(msg)) **/
    function registerPeerMsgHandler(handler) {
        peerMsgHandler = handler
    }

    /** "handler" should be a function which accepts an object (the object 
        type is the same as the messages you are sending with SpiderClient.send(msg)) **/
    function registerBroadcastMsgHandler(handler) {
        broadcastMsgHandler = handler
    }

    /** "handler" should be a function which accepts a string. */
    function registerErrorMsgHandler(handler) {
        errorMsgHandler = handler
    }

    return {
        send: send,
        broadcast: broadcast,
        requestIDs: requestIDs,
        registerIDMsgHandler: registerIDMsgHandler,
        registerPeerMsgHandler: registerPeerMsgHandler,
        registerBroadcastMsgHandler: registerBroadcastMsgHandler,
        registerErrorMsgHandler: registerErrorMsgHandler,
    }
}














