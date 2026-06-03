-- Avatar e banner persistem com a sala (branding).
ALTER TABLE rooms
    ADD COLUMN IF NOT EXISTS avatar_object_key TEXT,
    ADD COLUMN IF NOT EXISTS banner_object_key TEXT;

-- Anexos de mensagem: apenas quando a mensagem é persistível.
CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL,
    room_id UUID NOT NULL,
    object_key TEXT NOT NULL,
    media_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_attachments_message_id ON attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_attachments_room_id ON attachments(room_id);
