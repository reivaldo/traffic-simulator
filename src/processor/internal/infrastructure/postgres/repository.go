package postgresinfra

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/traffic-simulator/processor/internal/domain"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SaveProcessed(ctx context.Context, msg domain.MessageEnvelope) error {
	payload := map[string]any{}
	if len(msg.Payload) > 0 {
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			return err
		}
	}
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO messages (external_id, channel, recipient, payload, status)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (external_id) DO UPDATE SET
		     channel = EXCLUDED.channel,
		     recipient = EXCLUDED.recipient,
		     payload = EXCLUDED.payload,
		     status = EXCLUDED.status,
		     updated_at = now()`,
		msg.ExternalID,
		msg.Channel,
		msg.Recipient,
		payload,
		"processed_pending_publish",
	)
	return err
}

func (r *MessageRepository) MarkPublished(ctx context.Context, externalID string) error {
	_, err := r.db.Exec(
		ctx,
		"UPDATE messages SET status = $1, updated_at = now() WHERE external_id = $2",
		"processed",
		externalID,
	)
	return err
}
