package protocol

import "encoding/json"

type ClientMessage struct {
	Type    string
	Payload interface{}
}

func Unmarshal(data []byte) (*ClientMessage, error) {
	var result ClientMessage
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
