import { SpiderClient } from "../spiderclient.mjs";

window.onload = function () {
    const spider = SpiderClient();

    const output = document.getElementById("output");
    const sendButton = document.getElementById("sendButton");
    const broadcastButton = document.getElementById("broadcastButton");
    const IDButton = document.getElementById("IDButton");
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

    spider.regIDHandler(
        IDs => {
            console.log("received IDs");

            // remove current ID options 
            while (addresses.lastChild) {
                addresses.lastChild.remove();
            }

            // create new ID options
            IDs.forEach(
                (ID) => {
                    const opt = document.createElement("option");
                    opt.value = ID;
                    opt.textContent = ID;
                    addresses.append(opt);
                }
            );
        }
    );
    

    sendButton.onclick = () => {
        console.log("sending message: " + text.value);
        const destinationID = Number(addresses.children[addresses.selectedIndex].value);
        spider.send(text.value, destinationID);
    };

    broadcastButton.onclick = () => {
        console.log("broadcasting message: " + text.value);
        spider.broadcast(text.value);
    };

    IDButton.onclick = () => {
        console.log("fetching IDs");
        spider.requestIDs();
    };
};

















