/**
  Creates a Foot object, which is the client end-point for a browser websocket 
  network. 
  Send messages with foot.send()
  Get a list of the ids of the other connected clients with getids.
  You will need to register a handler for the response using regIdhandler(f).
  Register a handling function for incoming messages with regmsghandler(f)
**/
export function Foot() {
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
        const data = JSON.parse(event.data);
        switch (data.Type) {
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
                idHandler(data.Ids);
            } else {
                console.error("ids received but idHandler hasn't been registered");
            }
            break;
        }
    };


    /** send a message to the client whose id is "to". 
        "msg" is whatever you want **/
    function send(msg, to) {
        const msgObj = {Type: msgType.send, To: to, Payload: msg};
        conn.send(JSON.stringify(msgObj));
    }

    /** send "msg" to all clients connected to the network. 
        "msg" is whatever you want **/
    function broadcast(msg) {
        const msgObj = {Type: msgType.broadcast, Payload: msg};
        conn.send(JSON.stringify(msgObj));
    }

    /** request a list of the current client ids from the server. 
        to handle to reply call regIdhandler(f) with an appropriate function. **/
    function requestIds() {
        const msg = {Type: msgType.getIds};
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













