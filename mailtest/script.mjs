import { SpiderClient } from "../spiderclient.mjs";

window.onload = function () {
    const spider = SpiderClient();

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

    spider.regMsgHandler(
        obj => {
            console.log("received mail");
            displayMsg(obj);
        }
    );

    spider.regIdHandler(
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
        spider.send(text.value, to);
    };

    broadcastButton.onclick = () => {
        console.log("broadcasting message: " + text.value);
        spider.broadcast(text.value);
    };

    idButton.onclick = () => {
        console.log("fetching ids");
        spider.requestIds();
    };
};

















