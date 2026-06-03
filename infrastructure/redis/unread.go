package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

func unreadKey(userID, roomID string) string {
	return fmt.Sprintf("user:%s:room:%s:unread", userID, roomID)
}

func parseUnreadKey(key, userID string) (roomID string, ok bool) {
	prefix := "user:" + userID + ":room:"
	suffix := ":unread"
	if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
		return "", false
	}
	roomID = strings.TrimPrefix(key, prefix)
	roomID = strings.TrimSuffix(roomID, suffix)
	if roomID == "" {
		return "", false
	}
	return roomID, true
}

// ListUnreadByUser returns room_id -> unread count for keys > 0.
func ListUnreadByUser(ctx context.Context, rdb *redis.Client, userID string) (map[string]int64, error) {
	pattern := fmt.Sprintf("user:%s:room:*:unread", userID)
	result := make(map[string]int64)

	iter := rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		roomID, ok := parseUnreadKey(key, userID)
		if !ok {
			continue
		}
		n, err := rdb.Get(ctx, key).Int64()
		if err != nil {
			continue
		}
		if n > 0 {
			result[roomID] = n
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// ClearRoomUnread sets unread count to zero for a user in a room.
func ClearRoomUnread(ctx context.Context, rdb *redis.Client, userID, roomID string) error {
	return rdb.Set(ctx, unreadKey(userID, roomID), 0, 0).Err()
}

// GetRoomUnread reads the current unread count.
func GetRoomUnread(ctx context.Context, rdb *redis.Client, userID, roomID string) (int64, error) {
	val, err := rdb.Get(ctx, unreadKey(userID, roomID)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, nil
	}
	if n < 0 {
		return 0, nil
	}
	return n, nil
}
