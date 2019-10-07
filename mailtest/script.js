window.onload = function () {
    if (!window["WebSocket"]) {
        console.error("Browser doesn't support websockets. Exiting script");
        return;
    }

    const output = document.getElementById("output");
    const sendButton = document.getElementById("sendbutton");
    const addresses = document.getElementById("addresses");
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
        console.log("message received");
        writeToDiv(output, data.payload);
    };
    
    conn.onopen = () => {
        writeToDiv(output, "opened websocket");
        const msg = {to: 1, payload: "hello"};
        conn.send(JSON.stringify(msg));
    };

    conn.onclose = () => {
        writeToDiv(output, "closed websocket");
    };
};

















