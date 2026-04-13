-- Guarantees room owners are always ADMIN members in their own rooms.
INSERT INTO room_members (room_id, user_id, role, created_at)
SELECT r.id, r.owner_id, 'ADMIN', NOW()
FROM rooms r
WHERE r.owner_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM room_members rm
    WHERE rm.room_id = r.id
      AND rm.user_id = r.owner_id
  );

UPDATE room_members rm
SET role = 'ADMIN'
FROM rooms r
WHERE rm.room_id = r.id
  AND rm.user_id = r.owner_id
  AND rm.role <> 'ADMIN';
