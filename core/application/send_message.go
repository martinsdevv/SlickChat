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

	now := time.Now().UTC()

	// valida domínio
	if err := room.CanUserSendMessage(userID, membership.Role, now); err != nil {
		return "", err
	}

	messageID := uuid.New()

	payload := events.MessageSent{
		MessageID:        messageID.String(),
		RoomID:           room.ID.String(),
		SenderID:         userID.String(),
		Content:          content,
		MessageType:      "TEXT",
		IsZeroLogging:    !room.CanPersistMessages(),
		TTL:              room.TTL,
		DestroyAfterRead: false,
		ExpiresAt:        room.ExpiresAt,
		SentAt:           now,
	}

	event, err := events.NewEvent(
		events.EventTypeMessageSent,
		room.ID.String(),
		payload,
	)
	if err != nil {
		return "", err
	}

	if err := producer.Publish(context.Background(), event); err != nil {
		return messageID.String(), err
	}

	return messageID.String(), nil
}
