package protocol

import (
	"encoding/json"
	"fmt"
)

type MessageType string

const (
	TypeQuit    MessageType = "QUIT"
	TypePing    MessageType = "PING"
	TypePulic   MessageType = "PUBLIC"
	TypePrivate MessageType = "PRIVATE"
)

func Unmarshal(data []byte) (Message, error) {
	var intermediate struct{ Type MessageType }
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return nil, err
	}
	switch intermediate.Type {
	case TypeQuit:
		return quitMessage, nil
	case TypePing:
		return pingMessage, nil
	case TypePulic:
		var msg PublicMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	case TypePrivate:
		var msg PrivateMessage
		err = json.Unmarshal(data, &msg)
		if err != nil {
			return nil, err
		}
		return msg, nil
	default:
		return nil, fmt.Errorf("Unknown type")
	}
}

type Message interface {
	GetType() MessageType
}

type QuitMessage struct{}

var quitMessage = QuitMessage{}

func (m QuitMessage) GetType() MessageType {
	return TypeQuit
}

type PrivateMessage struct {
	Recepient string
	Text      string
}

func (m PrivateMessage) GetType() MessageType {
	return TypePrivate
}

type PublicMessage struct {
	Text string
}

func (m PublicMessage) GetType() MessageType {
	return TypePulic
}

type PingMessage struct{}

var pingMessage = PingMessage{}

func (m PingMessage) GetType() MessageType {
	return TypePing
}
