package irc

import "testing"

func TestGetMessage(t *testing.T) {
	input := ErrNoSuchServer
	message := input.GetMessage()
	expected := "%s :No such server"
	if message != expected {
		t.Errorf("%#q.GetMessage() => %#q, want %#q", input, message, expected)
	}
}
