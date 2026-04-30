package natsinfra

import (
	"encoding/json"
	"testing"

	"github.com/yourusername/traffic-simulator/ingestor/internal/domain"
)

func TestPublishJSONShape_MatchesMessageIntent(t *testing.T) {
	t.Parallel()
	intent, err := domain.NewMessageIntent("id-1", "email", "a@b.c", map[string]any{"k": 1})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{
		"external_id": intent.ExternalID,
		"channel":     intent.Channel,
		"recipient":   intent.Recipient,
		"payload":     intent.Payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["channel"] != "email" {
		t.Fatal(m)
	}
}
