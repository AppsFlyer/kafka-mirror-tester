package message

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/lib/types"
)

//go:generate go run ../gen/main/code-gen.go

// Format a message based on the parameters
func format(
	id types.ProducerID,
	seq types.SequenceNumber,
	timestamp time.Time,
	messageSize types.MessageSize,
) string {
	var b strings.Builder
	// build the header first
	fmt.Fprintf(&b, "%s;%d;%d;", id, seq, timestamp.UTC().UnixNano())

	// See how much space left for payload and add chars based on the space left
	left := int(messageSize) - b.Len()
	if left > 0 {
		fmt.Fprintf(&b, payload[:left])
	}
	return b.String()
}

// Parse parses the string message into the Data structure.
func parse(msg string) (data parsedData, err error) {
	parts := strings.Split(msg, ";")
	if len(parts) != 4 {
		err = errors.Errorf("msg should contain 4 parts but it doesn't. %s...", msg[:30])
		return
	}

	data.producerID = types.ProducerID(parts[0])
	sq, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	data.sequence = types.SequenceNumber(sq)

	ts, err := parseTs(parts[2])
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	data.timestamp = ts

	data.payload = []byte(parts[3])
	return
}

func parseTs(ts string) (time.Time, error) {
	i, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		log.Fatalf("Malformed timestamp %s. %+v", ts, err)
	}
	nano := i % 1e9
	sec := i / 1e9
	t := time.Unix(sec, nano).UTC()
	return t, nil
}
