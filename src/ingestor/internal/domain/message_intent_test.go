package domain

import "testing"

func TestNewMessageIntent_ValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	_, err := NewMessageIntent("", "email", "user@example.com", map[string]any{"k": "v"})
	if err == nil {
		t.Fatal("expected validation error for empty external_id")
	}
}

func TestNewMessageIntent_DefaultsPayloadToEmptyMap(t *testing.T) {
	t.Parallel()

	intent, err := NewMessageIntent("id-1", "email", "user@example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent.Payload == nil {
		t.Fatal("expected payload to be initialized")
	}
}
