//go:build integration

package integration

import (
	"flag"
	"slices"
	"strings"
	"testing"
	"unicode"
)

var testTags = flag.String("test-tags", "", "select integration groups by comma- or space-separated tags; empty runs all groups")

func requireTestTags(t *testing.T, groups ...string) {
	t.Helper()
	selected := strings.FieldsFunc(*testTags, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
	if len(selected) == 0 {
		return
	}
	for _, group := range groups {
		if slices.Contains(selected, group) {
			return
		}
	}
	t.Skip("integration group not selected by test-tags")
}
