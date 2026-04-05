package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageSent(t *testing.T) {
	event := MessageSent{
		RoomID: "1",
	}

	assert.Equal(t, "1", event.RoomID)
}
