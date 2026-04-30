CREATE TABLE IF NOT EXISTS failures (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
