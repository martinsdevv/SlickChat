package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	EventVersion int             `json:"event_version"`
	Timestamp    time.Time       `json:"timestamp"`
	PartitionKey string          `json:"partition_key"`
	Payload      json.RawMessage `json:"payload"`
}

func NewEvent(eventType string, partitionKey string, payload interface{}) (Event, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Event{}, err
	}

	return Event{
		EventID:      uuid.New().String(),
		EventType:    eventType,
		EventVersion: 1,
		Timestamp:    time.Now().UTC(),
		PartitionKey: partitionKey,
		Payload:      payloadBytes,
	}, nil
}
