const wsAddr = 'ws://localhost:8080/websock';
const MSG_QUIT = "QUIT";
const MSG_PUBLIC = "PUBLIC";

// wrapper object around chat state and chat dom elements
let state = {
    sock: null,
    sendBtn: null,
    messageLog: null,
    userInputArea: null,
    initialized: false,
};

function makeMessage(type, payload) {
    return {
        "type": type,
        "payload": payload ?? {}
    }
}

function initSocket(state) {
    let sock = new WebSocket(wsAddr);
    sock.onopen = () => {
        state.initialized = true;
        sock.onmessage = (event) => {
            addMessage(event.data)
        }
        sock.onclose = () => {
            console.log("Server closed connection");
            state.sendBtn.setAttribute("disabled", "disabled");
        }
    }
    
    state.sock = sock;
}

function initDom() {
    state.sendBtn = document.getElementById("send");
    state.messageLog = document.getElementById("message-log");
    state.userInputArea = document.getElementById("user-input");
    state.sendBtn.removeAttribute("disabled");
    state.sendBtn.onclick = sendMessage;
}

function addMessage(text) {
    let msgElem = document.createElement('p');
    msgElem.innerHTML = text;
    state.messageLog.appendChild(msgElem);
}

function sendMessage() {
    let text = state.userInputArea.value;
    if (!state.initialized || text === "") {
        return;
    }
    const message = JSON.stringify(makeMessage(MSG_PUBLIC, text));
    console.log("sending " + message);
    state.sock.send(message);
    state.userInputArea.value = "";
}

function StartApp() {
    initDom();
    initSocket(state);
}

StartApp();
