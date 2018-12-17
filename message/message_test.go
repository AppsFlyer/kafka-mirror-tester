package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

func TestCreateAndExtract(t *testing.T) {
	msg := Create("1", 5, 100)
	require.NotNil(t, msg, "Message should not be nil")

	// Make sure at least one ms passed before parsing it
	time.Sleep(1 * time.Millisecond)
	data := Extract(msg)
	require.NotNil(t, data, "Data should not be nil")

	assert := assert.New(t)
	assert.Equal(types.ProducerID("1"), data.ProducerID, "ProducerID should be 1")
	assert.Equal(types.SequenceNumber(5), data.Sequence, "Sequence number should be 5")
	assert.True(data.Latency > 1, "Latency should be > 1")
}
