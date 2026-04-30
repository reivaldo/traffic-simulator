package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/traffic-simulator/notification-service/internal/domain"
)

// EmailGateway simulates delivery to an Email provider via HTTP
type EmailGateway struct {
	apiURL string
	client *http.Client
}

func NewEmailGateway(apiURL string) *EmailGateway {
	return &EmailGateway{
		apiURL: apiURL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (g *EmailGateway) Send(ctx context.Context, notification domain.Notification) (status string, err error) {
	payload := map[string]string{
		"external_id": notification.ExternalID,
		"to":          notification.Recipient,
		"channel":     notification.Channel,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, g.apiURL+"/send", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "failed", fmt.Errorf("email http call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "failed", fmt.Errorf("email provider returned status code %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "failed", fmt.Errorf("email response parse failed: %w", err)
	}

	if statusVal, ok := result["status"].(string); ok {
		statusVal = strings.ToLower(strings.TrimSpace(statusVal))
		if statusVal == "success" || statusVal == "ok" || statusVal == "delivered" {
			return statusVal, nil
		}
		return statusVal, fmt.Errorf("email provider returned non-success status: %s", statusVal)
	}

	return "failed", fmt.Errorf("email response missing status field")
}
