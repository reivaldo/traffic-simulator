package domain

import (
	"encoding/json"
	"testing"
)

func TestNotification_JSON(t *testing.T) {
	t.Parallel()

	raw := `{"external_id":"e1","channel":"sms","recipient":"+1","payload":{"a":1}}`
	var n Notification
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatal(err)
	}
	if n.ExternalID != "e1" || n.Channel != "sms" {
		t.Fatalf("unexpected: %+v", n)
	}
}
