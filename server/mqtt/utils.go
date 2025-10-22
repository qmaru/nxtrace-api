package mqtt

import (
	"encoding/json"
)

type TaskMessage struct {
	Region     string   `json:"region"`
	Target     string   `json:"target"`
	Params     []string `json:"params"`
	SourceName string   `json:"source_name"`
	SourceID   string   `json:"source_id"`
}

// encodeMessage encode to json data
func encodeMessage(message any) ([]byte, error) {
	messageJson, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return messageJson, nil
}

// decodeTaskMessage decode to TaskMessage
func decodeTaskMessage(message []byte) (*TaskMessage, error) {
	var t TaskMessage
	if err := json.Unmarshal(message, &t); err == nil {
		return &t, nil
	}
	return &t, nil
}
