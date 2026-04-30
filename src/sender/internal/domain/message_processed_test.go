package domain

import (
	"encoding/json"
	"testing"
)

func TestMessageProcessed_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	orig := MessageProcessed{
		ExternalID: "e1",
		Channel:    "email",
		Recipient:  "a@b.c",
		Payload:    json.RawMessage(`{"a":1}`),
	}
	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	var out MessageProcessed
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.ExternalID != orig.ExternalID {
		t.Fatal(out)
	}
}
