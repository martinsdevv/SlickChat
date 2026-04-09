package ttl

import (
	"context"
	"time"

	"github.com/martinsdevv/slickchat/core/application"
	"github.com/martinsdevv/slickchat/core/contracts"
	kafkainfra "github.com/martinsdevv/slickchat/infrastructure/kafka"
	"github.com/martinsdevv/slickchat/infrastructure/log"
)

func Run(ctx context.Context, repo contracts.MessageRepository, producer *kafkainfra.Producer, pollEvery time.Duration, batchSize int) {
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	scan := func() {
		now := time.Now().UTC()
		msgs, err := repo.ListExpired(ctx, now, batchSize)
		if err != nil {
			log.Logger.Error("list expired messages", err)
			return
		}

		for _, m := range msgs {
			if m.ExpiresAt == nil {
				continue
			}
			if err := application.PublishMessageExpired(producer, m.ID, m.RoomID, *m.ExpiresAt); err != nil {
				log.Logger.Error("publish message expired", err, "message_id", m.ID)
			}
		}
	}

	scan()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scan()
		}
	}
}
