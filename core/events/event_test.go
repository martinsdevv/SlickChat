package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	ev, err := NewEvent(EventTypeMessageSent, "room-key", map[string]string{"k": "v"})
	require.NoError(t, err)
	assert.NotEmpty(t, ev.EventID)
	assert.Equal(t, EventTypeMessageSent, ev.EventType)
	assert.Equal(t, 1, ev.EventVersion)
	assert.Equal(t, "room-key", ev.PartitionKey)
	assert.Contains(t, string(ev.Payload), `"k":"v"`)
}

func TestMessageSent_JSONRoundTrip(t *testing.T) {
	sent := MessageSent{
		MessageID: "m1",
		RoomID:    "r1",
		SenderID:  "u1",
		Content:   "hello",
	}
	b, err := json.Marshal(sent)
	require.NoError(t, err)
	var out MessageSent
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, sent.MessageID, out.MessageID)
	assert.Equal(t, sent.RoomID, out.RoomID)
	assert.Equal(t, sent.Content, out.Content)
}
