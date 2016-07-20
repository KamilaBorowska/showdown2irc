package main

import (
	"bytes"
	"testing"
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
	if out != expected {
		t.Errorf("Testing not enough parameters in PASS, got %#q, want %#q", out, expected)
	}
}
