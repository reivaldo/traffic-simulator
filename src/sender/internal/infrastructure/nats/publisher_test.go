package natsinfra

import (
	"encoding/json"
	"testing"

	"github.com/yourusername/traffic-simulator/sender/internal/domain"
)

func TestMessageProcessed_Marshal_ForNATS(t *testing.T) {
	t.Parallel()

	msg := domain.MessageProcessed{ExternalID: "x", Channel: "c", Recipient: "r", Payload: json.RawMessage(`{}`)}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 10 {
		t.Fatal("unexpected payload", string(b))
	}
}
