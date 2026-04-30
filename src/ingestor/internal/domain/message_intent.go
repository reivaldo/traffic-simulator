package domain

import (
	"errors"
	"strings"
)

type MessageIntent struct {
	ExternalID string
	Channel    string
	Recipient  string
	Payload    map[string]any
}

func NewMessageIntent(externalID, channel, recipient string, payload map[string]any) (MessageIntent, error) {
	intent := MessageIntent{
		ExternalID: strings.TrimSpace(externalID),
		Channel:    strings.TrimSpace(channel),
		Recipient:  strings.TrimSpace(recipient),
		Payload:    payload,
	}
	if intent.Payload == nil {
		intent.Payload = map[string]any{}
	}
	if err := intent.Validate(); err != nil {
		return MessageIntent{}, err
	}
	return intent, nil
}

func (m MessageIntent) Validate() error {
	if m.ExternalID == "" {
		return errors.New("external_id is required")
	}
	if m.Channel == "" {
		return errors.New("channel is required")
	}
	if m.Recipient == "" {
		return errors.New("recipient is required")
	}
	return nil
}
