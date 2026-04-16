package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/segmentio/kafka-go"
)

type Handler func(events.Event)

func StartConsumer(broker, topic, groupID string, handler Handler) {
	for {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker},
			Topic:   topic,
			GroupID: groupID,
		})

		log.Logger.Info("consumer conectado", "topic", topic, "group_id", groupID)

		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Logger.Error("erro ao ler mensagem", "error", err, "topic", topic, "group_id", groupID)
				_ = reader.Close()
				time.Sleep(1200 * time.Millisecond)
				break
			}

			var event events.Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Logger.Error("erro ao deserializar evento", "error", err)
				continue
			}

			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						log.Logger.Error("panic no handler do consumer", "panic", recovered, "topic", topic)
					}
				}()
				handler(event)
			}()
		}
	}
}
