window.onload = function () {
    if (!window["WebSocket"]) {
        console.error("Browser doesn't support websockets. Exiting script");
        return;
    }

    const output = document.getElementById("output");
    const sendButton = document.getElementById("sendButton");
    const broadcastButton = document.getElementById("broadcastButton");
    const idButton = document.getElementById("idButton");
    const addresses = document.getElementById("addresses");
    const text = document.getElementById("text");
    const conn = new WebSocket("ws://" + document.location.host + "/ws");
    const msgType = {
        send: 1,
        broadcast: 2,
        getIds: 3
    };

    /** create a div containing "text" and add as a child of "div".
    **/
    function writeToDiv(div, text) {
        const textDiv = document.createElement("div");
        textDiv.textContent = text;
        div.appendChild(textDiv);
    }

    function displayMsg(msg) {
        const msgDiv = document.createElement("div");
        msgDiv.textContent = JSON.stringify(msg);
        output.append(msgDiv);
    }

    writeToDiv(output, "initiating mail test");


    //On receiving a message over the WebSocket, display it
    conn.onmessage = event => {
        const data = JSON.parse(event.data);
        console.log("message received: " + event.data);

        switch (data.Type) {
        case msgType.send:
        case msgType.broadcast:
            console.log("received mail");
            displayMsg(data);
            break;
        case msgType.getIds:
            console.log("received ids");
            while (addresses.lastChild) {
                addresses.lastChild.remove();
            }
            data.Ids.forEach(
                (id) => {
                    const opt = document.createElement("option");
                    opt.value = id;
                    opt.textContent = id;
                    addresses.append(opt);
                }
            );
            break;
        }
    };
    
    conn.onopen = () => {
        writeToDiv(output, "opened websocket");
    };

    conn.onclose = () => {
        writeToDiv(output, "closed websocket");
    };

    sendButton.onclick = () => {
        console.log("sending message: " + text.value);
        const to = Number(addresses.children[addresses.selectedIndex].value);
        const msg = {Type: msgType.send, To: to, Payload: text.value};
        conn.send(JSON.stringify(msg));
    };

    broadcastButton.onclick = () => {
        console.log("broadcasting message: " + text.value);
        const msg = {Type: msgType.broadcast, Payload: text.value};
        conn.send(JSON.stringify(msg));
    };

    idButton.onclick = () => {
        console.log("fetching ids");
        const msg = {type: msgType.getIds};
        conn.send(JSON.stringify(msg));
    };
        
};

















