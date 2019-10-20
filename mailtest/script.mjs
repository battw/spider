import { Foot } from "../foot.mjs";

window.onload = function () {
    const foot = Foot();

    const output = document.getElementById("output");
    const sendButton = document.getElementById("sendButton");
    const broadcastButton = document.getElementById("broadcastButton");
    const idButton = document.getElementById("idButton");
    const addresses = document.getElementById("addresses");
    const text = document.getElementById("text");

    function displayMsg(msg) {
        const msgDiv = document.createElement("div");
        msgDiv.textContent = JSON.stringify(msg);
        output.append(msgDiv);
    }

    displayMsg("initiating mail test");

    foot.regMsgHandler(
        obj => {
            console.log("received mail");
            displayMsg(obj);
        }
    );

    foot.regIdHandler(
        ids => {
            console.log("received ids");

            // remove current id options 
            while (addresses.lastChild) {
                addresses.lastChild.remove();
            }

            // create new id options
            ids.forEach(
                (id) => {
                    const opt = document.createElement("option");
                    opt.value = id;
                    opt.textContent = id;
                    addresses.append(opt);
                }
            );
        }
    );
    

    sendButton.onclick = () => {
        console.log("sending message: " + text.value);
        const to = Number(addresses.children[addresses.selectedIndex].value);
        foot.send(text.value, to);
    };

    broadcastButton.onclick = () => {
        console.log("broadcasting message: " + text.value);
        foot.broadcast(text.value);
    };

    idButton.onclick = () => {
        console.log("fetching ids");
        foot.requestIds();
    };
};

















