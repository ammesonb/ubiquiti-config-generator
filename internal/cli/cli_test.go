package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
)

func TestPreflight(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		want    int
		message string
	}{
		{nil, 0, "Usage:"},
		{[]string{"deploy", "--help"}, 0, "confirm-timeout"},
		{[]string{"bogus"}, 2, "expected deploy"},
		{[]string{"deploy", "file"}, 2, "--host is required"},
		{[]string{"deploy", "file", "--host", "router:22", "--user", "vyos", "--identity", "key", "--known-hosts", "hosts"}, 2, "interactive terminal"},
		{[]string{"deploy", "--host", "router:22", "--user", "vyos", "--identity", "key", "--known-hosts", "hosts", "file"}, 2, "interactive terminal"},
	} {
		var output bytes.Buffer
		got := Run(context.Background(), tc.args, Streams{Input: strings.NewReader(""), Output: &output, Error: &output})
		if got != tc.want || !strings.Contains(output.String(), tc.message) {
			t.Fatalf("%v: status=%d output=%s", tc.args, got, output.String())
		}
	}
}
func TestApprovalInput(t *testing.T) {
	for _, tc := range []struct {
		input    string
		accepted bool
	}{{"yes\n", true}, {" Y \n", true}, {"no\n", false}, {"\n", false}, {"yes", false}, {"", false}} {
		var output bytes.Buffer
		decision := <-approval(Streams{Input: strings.NewReader(tc.input), Output: &output}, true)(context.Background())
		if decision.Accepted != tc.accepted || decision.Err != nil {
			t.Fatalf("input %q: %+v", tc.input, decision)
		}
	}
}
func TestApprovalReturnsWhileInputBlocked(t *testing.T) {
	input, writer := io.Pipe()
	defer func() { _ = input.Close() }()
	defer func() { _ = writer.Close() }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	returned := make(chan (<-chan deployment.Decision), 1)
	go func() { returned <- approval(Streams{Input: input, Output: io.Discard}, false)(ctx) }()
	select {
	case <-returned:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("prompt blocked runner")
	}
}
func TestReportDoesNotExposeDeviceOutput(t *testing.T) {
	var output bytes.Buffer
	err := &deployment.StageError{Stage: deployment.StageCommit, Err: errors.New("secret config")}
	if report(Streams{Output: &output, Error: &output}, deployment.Outcome{}, err) != 1 || strings.Contains(output.String(), "secret config") {
		t.Fatal(output.String())
	}
}

type failedWriter struct{ err error }

func (w failedWriter) Write([]byte) (int, error) { return 0, w.err }
func TestApprovalDoesNotReadWhenPromptFails(t *testing.T) {
	failure := errors.New("terminal unavailable")
	decision := <-approval(Streams{Output: failedWriter{failure}}, false)(context.Background())
	if !errors.Is(decision.Err, failure) || decision.Accepted {
		t.Fatalf("%+v", decision)
	}
}
func TestReportRetainsCleanupFailure(t *testing.T) {
	var output bytes.Buffer
	err := errors.Join(&deployment.StageError{Stage: deployment.StageApproval, Err: deployment.ErrDeclined}, &deployment.StageError{Stage: deployment.StageCleanup, Err: errors.New("secret")})
	report(Streams{Output: &output, Error: &output}, deployment.Outcome{Applied: true}, err)
	if !strings.Contains(output.String(), "cleanup failed") || strings.Contains(output.String(), "secret") {
		t.Fatal(output.String())
	}
}
