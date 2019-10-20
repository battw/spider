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


    function send(msg, to) {
        const msgObj = {Type: msgType.send, To: to, Payload: msg};
        conn.send(JSON.stringify(msgObj));
    }

    function broadcast(msg) {
        const msgObj = {Type: msgType.broadcast, Payload: msg};
        conn.send(JSON.stringify(msgObj));
    }

    function requestIds() {
        const msg = {Type: msgType.getIds};
        conn.send(JSON.stringify(msg));
    }

    /** "handler" should be a function which accepts an array of numbers **/
    function regIdHandler(handler) {
        idHandler = handler;
    }

    /** "handler" should be a function which accepts an object **/
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














