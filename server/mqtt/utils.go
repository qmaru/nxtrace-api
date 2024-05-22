package mqtt

import (
	"encoding/json"
)

// encodeMessage encode to json data
func encodeMessage(message any) ([]byte, error) {
	messageJson, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return messageJson, nil
}

// decodeMessage decode to go map
func decodeMessage(message []byte) (map[string]any, error) {
	var messageJson map[string]any
	err := json.Unmarshal(message, &messageJson)
	if err != nil {
		return nil, err
	}
	return messageJson, nil
}
