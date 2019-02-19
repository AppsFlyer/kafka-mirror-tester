package consumer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/appsflyer/kafka-mirror-tester/lib/message"
)

func TestValidateSequence(t *testing.T) {
	initPrometheus()
	assert := assert.New(t)

	// validate initial state
	assert.Equal(uint64(0), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(0), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Now start sending messages and observe counts
	data := &message.Data{
		ProducerID: "1",
		Topic:      "t",
		MessageKey: 1,
		Sequence:   0,
	}
	validateSequence(data)
	assert.Equal(uint64(0), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(1), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Same message again
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(1), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// increase sequence
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t",
		MessageKey: 1,
		Sequence:   1,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(2), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Send to a different topic
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t2",
		MessageKey: 1,
		Sequence:   0,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(3), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Send from a different producer
	data = &message.Data{
		ProducerID: "2",
		Topic:      "t",
		MessageKey: 1,
		Sequence:   0,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(4), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Send with a different message key
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t",
		MessageKey: 2,
		Sequence:   0,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(5), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Skip a few messages
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t",
		MessageKey: 1,
		Sequence:   5,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(5), inOrderMessagesCount)
	assert.Equal(uint64(4), skippedMessagesCount)

	// Skip an old message
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t",
		MessageKey: 1,
		Sequence:   2,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(1), oldMessagesCount)
	assert.Equal(uint64(5), inOrderMessagesCount)
	assert.Equal(uint64(4), skippedMessagesCount)
}
