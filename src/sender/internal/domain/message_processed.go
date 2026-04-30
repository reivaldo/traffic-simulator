package domain

import "encoding/json"

type MessageProcessed struct {
	ExternalID string          `json:"external_id"`
	Channel    string          `json:"channel"`
	Recipient  string          `json:"recipient"`
	Payload    json.RawMessage `json:"payload"`
}
