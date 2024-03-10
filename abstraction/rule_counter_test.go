package abstraction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasCounter(t *testing.T) {
	resetCounters()
	assert.False(t, HasCounter("nonexistent"), "Nonexistent counter should not exist")

	MakeCounter("counter", 1, 10)
	assert.Truef(t, HasCounter("counter"), "Counter should exist")
}

func TestGetCounter(t *testing.T) {
	resetCounters()
	assert.Falsef(t, HasCounter("counter"), "Counter should not exist yet")
	assert.Nilf(t, GetCounter("counter"), "Nonexistent counter should be nil")
	MakeCounter("counter", 1, 10)

	counter := GetCounter("counter")

	assert.Equal(t, 1, counter.Next(), "Initial value should be the default")
	assert.Equal(t, 11, counter.Next(), "Next counter value should be increased by step")
}
