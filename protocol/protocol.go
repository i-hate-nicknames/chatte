package protocol

import "encoding/json"

// todo: sum types? How do I make a heterogeneous
// client message

type Message struct {
	Type    string
	Payload interface{}
}

func Unmarshal(data []byte) (*Message, error) {
	var result Message
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
