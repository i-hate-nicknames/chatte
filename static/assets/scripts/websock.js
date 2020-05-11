let sock = new WebSocket('ws://localhost:8080/websock');
sock.onopen = function (event) {
    sock.send("hello server");
}