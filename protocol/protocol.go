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

type Message struct {
	Discriminator MessageType
	Private       *PrivateMessage
	Public        *PublicMessage
}

type PrivateMessage struct {
	Recipient string
	Text      string
}

type PublicMessage struct {
	Text string
}

func Unmarshal(data []byte) (*Message, error) {
	var intermediate struct{ Type MessageType }
	err := json.Unmarshal(data, &intermediate)
	if err != nil {
		return nil, err
	}
	message := &Message{Discriminator: intermediate.Type}
	switch intermediate.Type {
	case TypePulic:
		var inner PublicMessage
		err = json.Unmarshal(data, &inner)
		if err != nil {
			return nil, err
		}
		message.Public = &inner
	case TypePrivate:
		var inner PrivateMessage
		err = json.Unmarshal(data, &inner)
		if err != nil {
			return nil, err
		}
		message.Private = &inner
	default:
		return nil, fmt.Errorf("Unknown type")
	}
	return message, nil
}
