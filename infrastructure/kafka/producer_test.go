package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) WriteMessages(ctx context.Context, msgs ...interface{}) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func TestNewProducer(t *testing.T) {
	producer := NewProducer("localhost:9092")
	assert.NotNil(t, producer)
	assert.NotNil(t, producer.writer)
	assert.Equal(t, "localhost:9092", producer.writer.Addr.String())
}

func TestProducer_Publish(t *testing.T) {
	// This is more of an integration test
	t.Skip("Unit test for Publish requires mocking kafka.Writer; better suited for integration tests")
}
