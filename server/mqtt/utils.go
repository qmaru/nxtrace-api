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

// decodeMessage decode to go map
// func decodeMessage(message []byte) (map[string]any, error) {
// 	var messageJson map[string]any
// 	err := json.Unmarshal(message, &messageJson)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return messageJson, nil
// }

// decodeTaskMessage decode to TaskMessage
func decodeTaskMessage(message []byte) (*TaskMessage, error) {
	var t TaskMessage
	if err := json.Unmarshal(message, &t); err == nil {
		return &t, nil
	}

	// 兼容旧格式（params 可能为 []any）
	var raw map[string]any
	if err := json.Unmarshal(message, &raw); err != nil {
		return nil, err
	}

	t.Region, _ = raw["region"].(string)
	t.Target, _ = raw["target"].(string)
	t.SourceName, _ = raw["source_name"].(string)
	t.SourceID, _ = raw["source_id"].(string)

	if p, ok := raw["params"].([]any); ok {
		t.Params = make([]string, 0, len(p))
		for _, v := range p {
			if s, ok := v.(string); ok {
				t.Params = append(t.Params, s)
			}
		}
	}

	return &t, nil
}
