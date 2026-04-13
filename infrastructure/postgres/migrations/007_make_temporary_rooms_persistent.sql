-- Temporary rooms keep existing; only messages should expire.
UPDATE rooms
SET expires_at = NULL
WHERE type = 'TEMPORARY';
