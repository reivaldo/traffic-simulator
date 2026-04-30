CREATE TABLE IF NOT EXISTS message_events (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(id),
    event_type VARCHAR(64) NOT NULL,
    event_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
