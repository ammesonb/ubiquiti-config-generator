package logging

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLevelsAndColor(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		var output bytes.Buffer
		log := New(&output)
		log.color = enabled
		log.now = func() time.Time { return time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC) }
		for _, tc := range []struct {
			level, code string
			write       func(string) error
		}{
			{"INFO", "36", log.Info}, {"WARN", "33", log.Warn}, {" ERR", "31", log.Error}, {"CRIT", "91", log.Critical},
		} {
			output.Reset()
			if err := tc.write("first\nsecond"); err != nil {
				t.Fatal(err)
			}
			var expected string
			for _, message := range []string{"first", "second"} {
				line := "2026-10-18T12:00:00Z [" + tc.level + "] " + message
				if enabled {
					line = "\x1b[" + tc.code + "m" + line + "\x1b[0m"
				}
				expected += line + "\n"
			}
			if output.String() != expected {
				t.Fatalf("got %q, want %q", output.String(), expected)
			}
		}
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }
func TestWriteFailure(t *testing.T) {
	failure := errors.New("closed output")
	if err := New(failingWriter{failure}).Warn("message"); !errors.Is(err, failure) {
		t.Fatal(err)
	}
}
func TestRedirectedOutputIsPlain(t *testing.T) {
	var output bytes.Buffer
	if SupportsColor(&output) {
		t.Fatal("buffer reported as a color terminal")
	}
	if err := New(&output).Info("message"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "\x1b[") {
		t.Fatal("redirected output contains styling")
	}
}
