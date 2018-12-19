package consumer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/message"
)

func TestValidateSequence(t *testing.T) {
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
		Sequence:   0,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(4), inOrderMessagesCount)
	assert.Equal(uint64(0), skippedMessagesCount)

	// Skip a few messages
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t",
		Sequence:   5,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(0), oldMessagesCount)
	assert.Equal(uint64(4), inOrderMessagesCount)
	assert.Equal(uint64(4), skippedMessagesCount)

	// Skip an old message
	data = &message.Data{
		ProducerID: "1",
		Topic:      "t",
		Sequence:   2,
	}
	validateSequence(data)
	assert.Equal(uint64(1), sameMessagesCount)
	assert.Equal(uint64(1), oldMessagesCount)
	assert.Equal(uint64(4), inOrderMessagesCount)
	assert.Equal(uint64(4), skippedMessagesCount)
}
