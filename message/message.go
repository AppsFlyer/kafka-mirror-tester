package message

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

// Data represent the data sent in a message.
type Data struct {
	ProducerID types.ProducerID
	Sequence   types.SequenceNumber
	Timestamp  time.Time
	Payload    string
}

// Format a message based on the parameters
func Format(
	id types.ProducerID,
	seq types.SequenceNumber,
	messageSize types.MessageSize,
) string {
	ts := time.Now().UTC().UnixNano()
	var b strings.Builder
	// build the header first
	fmt.Fprintf(&b, "%s;%d;%d;", id, seq, ts)

	// See how much space left for payload and add chars based on the space left

	for left := int(messageSize) - b.Len(); left > 0; left-- {
		fmt.Fprintf(&b, "X")
	}

	return b.String()
}

// Parse parses the string message into the Data structure.
func Parse(msg string) (data Data, err error) {
	parts := strings.Split(msg, ";")
	if len(parts) != 4 {
		err = errors.Errorf("msg should contain 4 parts but it doesn't. %s...", msg[:30])
		return
	}

	data.ProducerID = types.ProducerID(parts[0])
	sq, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	data.Sequence = types.SequenceNumber(sq)
	ts, err := parseTs(parts[2])
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	data.Timestamp = ts
	data.Payload = parts[3]
	return
}

func parseTs(ts string) (time.Time, error) {
	i, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		panic(err)
	}
	nano := i % 1e9
	sec := i / 1e9

	t := time.Unix(sec, nano).UTC()
	return t, nil
}
