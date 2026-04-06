package kafka

import (
	"context"
	"encoding/json"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/martinsdevv/slickchat/infrastructure/log"
	"github.com/segmentio/kafka-go"
)

type Handler func(events.Event)

func StartConsumer(broker, topic, groupID string, handler Handler) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Logger.Error("erro ao ler mensagem", "error", err)
			continue
		}

		var event events.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Logger.Error("erro ao deserializar evento", "error", err)
			continue
		}

		handler(event)
	}
}
