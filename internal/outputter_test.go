package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOutputIO(t *testing.T) {
	io := NewOutputIO()
	assert.NotNil(t, io)
}
