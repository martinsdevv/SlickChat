package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/domain"
	"github.com/martinsdevv/slickchat/core/events"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
)

func SendMessage(
	producer *kafkainfra.Producer,
	room *domain.Room,
	membership *domain.RoomMembership,
	userID uuid.UUID,
	content string,
) (string, error) {
	id := uuid.New()
	if err := SendMessageWithID(producer, room, membership, userID, id, content); err != nil {
		return "", err
	}
	return id.String(), nil
}

// SendMessageWithID publica MessageSent com o id fornecido (ex.: após gravar sender_id no Redis no gateway).
func SendMessageWithID(
	producer *kafkainfra.Producer,
	room *domain.Room,
	membership *domain.RoomMembership,
	userID uuid.UUID,
	messageID uuid.UUID,
	content string,
) error {

	now := time.Now().UTC()

	if err := room.CanUserSendMessage(userID, membership.Role, now); err != nil {
		return err
	}

	var expiresAt *time.Time
	if room.TTL > 0 {
		t := now.Add(time.Duration(room.TTL) * time.Second)
		expiresAt = &t
	}

	payload := events.MessageSent{
		MessageID:        messageID.String(),
		RoomID:           room.ID.String(),
		SenderID:         userID.String(),
		Content:          content,
		MessageType:      "TEXT",
		IsZeroLogging:    !room.CanPersistMessages(),
		TTL:              room.TTL,
		DestroyAfterRead: false,
		ExpiresAt:        expiresAt,
		SentAt:           now,
	}

	event, err := events.NewEvent(
		events.EventTypeMessageSent,
		room.ID.String(),
		payload,
	)
	if err != nil {
		return err
	}

	return producer.Publish(context.Background(), event)
}
