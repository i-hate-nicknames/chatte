const wsAddr = 'ws://localhost:8080/websock';
const MSG_QUIT = "QUIT";
const MSG_PUBLIC = "PUBLIC";
const MSG_PING = "PING";
const MSG_PRIVATE = "PRIVATE";
const PING_TIMEOUT = 120000;

// wrapper object around chat state and chat dom elements
let state = {
    sock: null,
    sendBtn: null,
    messageLog: null,
    userInputArea: null,
    running: false,
};

function makeMessage(type, payload) {
    return {
        "type": type,
        "payload": payload ?? {}
    }
}

function initSocket(state, onOpen) {
    let sock = new WebSocket(wsAddr);
    sock.onopen = () => {
        state.running = true;
        sock.onmessage = (event) => {
            addMessage(event.data)
        }
        sock.onclose = () => {
            state.running = false;
            console.log("Server closed connection");
            state.sendBtn.setAttribute("disabled", "disabled");
        }
        onOpen();
    }
    
    state.sock = sock;
}

function initDom() {
    state.sendBtn = document.getElementById("send");
    state.messageLog = document.getElementById("message-log");
    state.userInputArea = document.getElementById("user-input");
    state.sendBtn.removeAttribute("disabled");
    state.sendBtn.onclick = sendBroadcastMessage;
}

function addMessage(text) {
    let msgElem = document.createElement('p');
    msgElem.innerHTML = text;
    state.messageLog.appendChild(msgElem);
}

function sendBroadcastMessage() {
    let text = state.userInputArea.value;
    if (!state.running || text === "") {
        return;
    }
    sendMessage(MSG_PUBLIC, text);
    state.userInputArea.value = "";
}

function sendMessage(type, payload) {
    const message = JSON.stringify(makeMessage(type, payload));
    console.log("sending " + message);
    state.sock.send(message);
}

function startPinger() {
    pinger = () => {
        if (!state.running) {
            return;
        }
        sendMessage(MSG_PING);
        setTimeout(pinger, PING_TIMEOUT);
    };
    pinger();
}

function StartApp() {
    initDom();
    initSocket(state, startPinger);
}

StartApp();
