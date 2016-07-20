package irc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMessage(t *testing.T) {
	input := ErrNoSuchServer
	assert.Equal(t, input.GetMessage(), "%s :No such server", "%#q.GetMessage()", input)
}
