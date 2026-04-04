package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProducer(t *testing.T) {
	producer := NewProducer("localhost:9092")
	assert.NotNil(t, producer)
	assert.NotNil(t, producer.writer)
	assert.Equal(t, "localhost:9092", producer.writer.Addr.String())
}

func TestProducer_Publish(t *testing.T) {
	t.Skip("Unit test for Publish requires mocking kafka.Writer; better suited for integration tests")
}
