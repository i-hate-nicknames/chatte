const wsAddr = 'ws://localhost:8080/websock';

let state = {
    sock: null,
    sendBtn: null,
    messageLog: null,
    userInputArea: null,
    initialized: false,
};

function InitSocket(state) {
    let sock = new WebSocket(wsAddr);
    sock.onopen = function (event) {
        state.initialized = true;
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

function sendMessage() {
    let text = state.userInputArea.value;
    if (!state.initialized || text === "") {
        return;
    }
    console.log("sending " + text);
    state.sock.send(text);
    state.userInputArea.value = "";
}

function StartApp() {
    initDom();
    InitSocket(state)
}

StartApp();
