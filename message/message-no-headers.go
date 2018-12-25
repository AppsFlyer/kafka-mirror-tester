package message

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"gitlab.appsflyer.com/rantav/kafka-mirror-tester/types"
)

//go:generate go run ../code-gen.go

// Format a message based on the parameters
func format(
	id types.ProducerID,
	seq types.SequenceNumber,
	messageSize types.MessageSize,
) string {
	var b strings.Builder
	// build the header first
	fmt.Fprintf(&b, "%s;%d;", id, seq)

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
	if len(parts) != 3 {
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
	data.payload = []byte(parts[2])
	return
}
