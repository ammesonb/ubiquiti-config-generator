package test_helpers

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestExpect(t *testing.T) {
	tracker := NewUncheckedAssertionTracker(t)

	expected := map[string]any{"a": 1, "b": 3}
	actual := map[string]any{"a": 2, "b": 3}

	for key, expected := range expected {
		tracker.Expect(key, expected, actual[key])
	}

	assert.True(t, reflect.DeepEqual(expected, tracker.expected))
	assert.True(t, reflect.DeepEqual(actual, tracker.actual))
}

func TestAssertionTrackerCheck(t *testing.T) {
	tracker := NewAssertionTracker(t)

	expected := map[string]any{"a": 1, "b": 2}
	actual := map[string]any{"a": 1, "b": 2}

	for key, expected := range expected {
		tracker.Expect(key, expected, actual[key])
	}

	assert.Empty(t, tracker.GetMismatchedValues())
}

func TestAssertionTrackerFail(t *testing.T) {
	tracker := NewUncheckedAssertionTracker(t)

	expected := map[string]any{"a": 1, "b": 2}
	actual := map[string]any{"a": 3, "b": 4}

	for key, expected := range expected {
		tracker.Expect(key, expected, actual[key])
	}

	assert.Len(t, tracker.GetMismatchedValues(), 2)
	for key := range expected {
		assert.Contains(t, tracker.GetMismatchedValues(), formatMismatchedValue(key, expected[key], actual[key]))
	}
}
