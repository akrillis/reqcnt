package hash

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHash(t *testing.T) {
	assert.Equal(t, 128, len(Hash("abcdefghihklmnopqrstuvwxyz")))
	assert.Equal(t, 128, len(Hash("0123456789")))
}
