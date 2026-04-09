CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    owner_id UUID,
    ttl INT DEFAULT 0,
    paranoid_mode BOOL DEFAULT FALSE,
    zero_logging BOOL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);

CREATE INDEX idx_rooms_expires_at ON rooms(expires_at);

CREATE TABLE room_members (
    room_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX idx_room_members_user_id ON room_members(user_id);
