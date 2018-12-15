package message

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

func TestFormat(t *testing.T) {

	assert := assert.New(t)

	// Check length
	msg := Format("1", 0, 100)
	assert.Equal(100, len(msg), "Length should be 100")

	// Check minimal length
	msg = Format("1", 0, 1)
	assert.True(len(msg) > 1, "Length should be > 1")

	// Check very long messages
	msg = Format("1", 0, 1e4)
	assert.Equal(int(1e4), len(msg), "Length should be 1e3")
}

func TestParse(t *testing.T) {

	assert := assert.New(t)

	// Create a message
	msg := Format("1", 0, 100)
	// Make sure at least one ms passed before parsing it
	time.Sleep(1 * time.Millisecond)
	parsed, err := Parse(msg)

	require.Nil(t, err, "There should not be an error")

	assert.Equal(types.ProducerID("1"), parsed.ProducerID, "ProducerID should be 1")
	assert.Equal(types.SequenceNumber(0), parsed.Sequence, "Sequence should be 0")
	assert.True(parsed.Timestamp.Before(time.Now().UTC()), "At least one millisecond should have passed")
}

// from fib_test.go
func BenchmarkFormat(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		Format("xx", 5, 1000)
	}
}
