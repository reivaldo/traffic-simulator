package natsinfra

import (
	"encoding/json"
	"testing"

	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

func TestProcessedPayload_JSONShape(t *testing.T) {
	t.Parallel()

	msg := domain.MessageEnvelope{
		ExternalID: "e1",
		Channel:    "c",
		Recipient:  "r",
		Payload:    json.RawMessage(`{}`),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["external_id"] != "e1" {
		t.Fatal(m)
	}
}
