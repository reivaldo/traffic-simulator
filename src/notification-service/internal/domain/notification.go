package domain

import "encoding/json"

type Notification struct {
	ExternalID string          `json:"external_id"`
	Channel    string          `json:"channel"`
	Recipient  string          `json:"recipient"`
	Payload    json.RawMessage `json:"payload"`
}

type ProviderResult struct {
	Provider string
	Status   string
	Error    error
	Duration float64 // milliseconds
}
