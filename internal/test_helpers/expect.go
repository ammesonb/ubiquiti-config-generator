// Test Helpers around expectations

package test_helpers

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertionTracker adds a set of expected and actual values, for single-assertion testing
type AssertionTracker struct {
	t        *testing.T
	expected map[string]any
	actual   map[string]any
}

// Expect stores the expected and actual values under a given key
func (tracker *AssertionTracker) Expect(key string, expected any, actual any) {
	tracker.expected[key] = expected
	tracker.actual[key] = actual
}

func (tracker *AssertionTracker) GetMismatchedValues() []string {
	var values []string
	for key, expected := range tracker.expected {
		actual := tracker.actual[key]
		if !reflect.DeepEqual(expected, actual) {
			values = append(values, formatMismatchedValue(key, expected, actual))
		}
	}
	return values
}

func formatMismatchedValue(key string, expected any, actual any) string {
	return fmt.Sprintf("%s: Expected '%v' but got '%v'", key, expected, actual)
}

// NewAssertionTracker creates a new AssertionTracker that automatically checks results on cleanup,
// and will fail the test if values do not match
func NewAssertionTracker(t *testing.T) *AssertionTracker {
	tracker := &AssertionTracker{
		t:        t,
		expected: make(map[string]any),
		actual:   make(map[string]any),
	}

	t.Cleanup(func() {
		t.Helper()
		for _, mismatch := range tracker.GetMismatchedValues() {
			t.Error(mismatch)
		}
	})

	return tracker
}

// NewUncheckedAssertionTracker creates a new AssertionTracker that does not automatically check results
func NewUncheckedAssertionTracker(t *testing.T) *AssertionTracker {
	return &AssertionTracker{
		t:        t,
		expected: make(map[string]any),
		actual:   make(map[string]any),
	}
}
