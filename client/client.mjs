// TODO - CLEAN ME
// TODO - improve logging

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
        console.error("Browser doesn't support websockets. Exiting script");
        return null;
    }

    const msgType = {
        send: 1,
        broadcast: 2,
        getIds: 3
    };

    let conn = new WebSocket("ws://" + document.location.host + "/ws");
    let idHandler = null;
    let msgHandler = null; 

    conn.onopen = () => {
        console.log("opened websocket");
    };

    conn.onclose = () => {
        console.log("closed websocket");
    };

    conn.onmessage = event => {
        console.log("incoming message")
        const data = JSON.parse(event.data);
        switch (data.MsgType) {
        case msgType.send:
        case msgType.broadcast:
            if (msgHandler !== null) {
                msgHandler(data);
            } else {
                console.error("message received but msgHandler hasn't been registered");
            }
            break;
        case msgType.getIds:
            if (idHandler !== null) {
                idHandler(data.IDs);
            } else {
                console.error("ids received but idHandler hasn't been registered");
            }
            break;
        }
    };


    /** send a message to the client whose id is "to". 
        "msg" is whatever you want **/
    function send(msg, to) {
        const msgObj = {MsgType: msgType.send, DestinationID: to, Payload: msg};
        conn.send(JSON.stringify(msgObj));
    }

    /** send "msg" to all clients connected to the network. 
        "msg" is whatever you want **/
    function broadcast(msg) {
        const msgObj = {MsgType: msgType.broadcast, Payload: msg};
        conn.send(JSON.stringify(msgObj));
    }

    /** request a list of the current client ids from the server. 
        to handle to reply call regIdhandler(f) with an appropriate function. **/
    function requestIds() {
        const msg = {MsgType: msgType.getIds};
        conn.send(JSON.stringify(msg));
    }

    /** "handler" should be a function which accepts an array of numbers **/
    function regIdHandler(handler) {
        idHandler = handler;
    }

    /** "handler" should be a function which accepts an object (the object 
        type is the same as the messages you are sending with foot.send(msg)) **/
    function regMsgHandler(handler) {
        msgHandler = handler;
    }

    return {
        send: send,
        broadcast: broadcast,
        requestIds: requestIds,
        regIdHandler: regIdHandler,
        regMsgHandler: regMsgHandler
    };
}














