window.onload = function () {
    if (!window["WebSocket"]) {
        console.error("Browser doesn't support websockets. Exiting script");
        return;
    }

    const output = document.getElementById("output");
    const send = document.getElementById("send");
    const addresses = document.getElementById("addresses");
    const text = document.getElementById("text");
    const conn = new WebSocket("ws://" + document.location.host + "/ws");

    function writeToDiv(div, text) {
        const textDiv = document.createElement("div");
        textDiv.textContent = text;
        div.appendChild(textDiv);
    }

    writeToDiv(output, "initiating mail test");


    //On receiving a message over the WebSocket, display it
    conn.onmessage = event => {
        const data = JSON.parse(event.data);
        console.log("message received: " + String(data));
        writeToDiv(output, data.payload);
        writeToDiv(output, data.ids);
    };
    
    conn.onopen = () => {
        writeToDiv(output, "opened websocket");
    };

    conn.onclose = () => {
        writeToDiv(output, "closed websocket");
    };

    send.onclick = () => {
        console.log("sending message: " + text.value);
        const msg = {to: -2, payload: text.value};
        conn.send(JSON.stringify(msg));
    };
        
};

















