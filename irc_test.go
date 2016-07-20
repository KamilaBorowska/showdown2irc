package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type closeableBuffer struct {
	bytes.Buffer
}

func (*closeableBuffer) Close() error {
	return nil
}

func createBuffer(input []string) closeableBuffer {
	var buffer bytes.Buffer
	for _, elem := range input {
		buffer.WriteString(elem)
		buffer.WriteByte('\n')
	}
	buffer.WriteString("QUIT\n")
	return closeableBuffer{buffer}
}

func TestNeedMoreParams(t *testing.T) {
	buffer := createBuffer([]string{"PASS"})
	connectionListen(&buffer)
	out := buffer.String()
	expected := ":showdown 461 * PASS :Not enough parameters\r\n:showdown QUIT *\r\n"
	assert.Equal(t, out, expected, "PASS with not enough arguments")
}
