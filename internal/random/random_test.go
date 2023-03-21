package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestString(t *testing.T) {
	for i := 1; i < 100; i++ {
		assert.Equal(t, i, len(String(i)))
	}
}
