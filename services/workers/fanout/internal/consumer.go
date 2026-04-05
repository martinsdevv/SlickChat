package fanout

import (
	"context"
	"encoding/json"

	"github.com/martinsdevv/slickchat/infrastructure/log"

	"github.com/martinsdevv/slickchat/core/events"
	"github.com/segmentio/kafka-go"
)

func StartConsumer(broker string, handler func(events.Event)) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "message-events",
		GroupID: "fanout-group",
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Logger.Error("Erro ao iniciar consumer: ", err)
			continue
		}

		var event events.Event
		json.Unmarshal(msg.Value, &event)

		handler(event)
	}
}
