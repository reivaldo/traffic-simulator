package domain

import (
	"encoding/json"
	"testing"
)

func TestMessageEnvelope_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	orig := MessageEnvelope{
		ExternalID: "e1",
		Channel:    "email",
		Recipient:  "a@b.c",
		Payload:    json.RawMessage(`{"k":"v"}`),
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	var out MessageEnvelope
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.ExternalID != orig.ExternalID || string(out.Payload) != string(orig.Payload) {
		t.Fatalf("mismatch: %+v vs %+v", out, orig)
	}
}
