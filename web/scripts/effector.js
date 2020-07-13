import { createStore, createEffect, createEvent } from 'effector'

const makeLogger = text => data => { console.log(`${text}: ${data}`)}

const wsAddr = 'ws://127.0.0.1:8080/websock';

const onMessage = createEvent('new message')

const sock = new WebSocket(wsAddr);
sock.onopen = () => {
    sock.onmessage = msg => onMessage(msg)
    sock.onclose = () => { console.log("closing") }
}

const onMessageParse = onMessage.map(msg => JSON.parse(msg.data))

onMessageParse.watch(makeLogger("Message from server"))

const $msgs = createStore([])

$msgs.on(onMessage, (msgs, newMsg) => [...msgs, newMsg])

const $msgCount = $msgs.map(msgs => msgs.length)

$msgCount.watch(makeLogger("Msg count"))

const sendText = createEvent('send')

const sendSerialized = sendText.map(text => JSON.stringify({type: "PUBLIC", payload: text}))

sendSerialized.watch(makeLogger("Sending"))