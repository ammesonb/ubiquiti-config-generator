package vyos

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
)

type fakeTransport struct {
	calls  int
	closes int
	input  []byte
	result sshclient.Result
	err    error
}

func (f *fakeTransport) Run(_ context.Context, _ string, input []byte) (sshclient.Result, error) {
	f.calls++
	f.input = input
	return f.result, f.err
}

func (f *fakeTransport) Close() error { f.closes++; return nil }

func requireStage(t *testing.T, err error, expected deployment.Stage) {
	t.Helper()
	var stage *deployment.StageError
	if !errors.As(err, &stage) || stage.Stage != expected {
		t.Fatalf("expected %s failure, got %v", expected, err)
	}
}

func TestFailedPrerequisitesPreventLaterOperations(t *testing.T) {
	ctx := context.Background()
	for _, failure := range []string{"upload", "load", "commit"} {
		t.Run(failure, func(t *testing.T) {
			client := &fakeTransport{}
			s := newSession(client)
			cause := errors.New("native failure")
			if failure == "upload" {
				client.err = cause
			}
			file, err := s.Upload(ctx, deployment.Artifact{Content: []byte("configuration")})
			if failure == "upload" {
				requireStage(t, err, deployment.StageUpload)
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if failure == "load" {
					client.err = cause
				}
				err = s.LoadCandidate(ctx, file)
				if failure == "load" {
					requireStage(t, err, deployment.StageLoad)
				} else if err != nil {
					t.Fatal(err)
				}
			}
			client.err = cause
			before := client.calls
			err = s.CommitConfirmed(ctx, time.Minute)
			requireStage(t, err, deployment.StageCommit)
			if failure == "commit" {
				if !errors.Is(err, cause) {
					t.Fatal("commit failure lost its cause")
				}
			} else if client.calls != before {
				t.Fatal("commit contacted the device after a failed prerequisite")
			}
			before = client.calls
			requireStage(t, s.Confirm(ctx), deployment.StageConfirm)
			requireStage(t, s.Save(ctx), deployment.StageSave)
			if client.calls != before {
				t.Fatal("failed deployment allowed confirmation or saving")
			}
		})
	}
}

func TestUploadPreservesBytesAndLoadRejectsForeignPaths(t *testing.T) {
	client := &fakeTransport{}
	s := newSession(client)
	content := []byte("description \"'$(not-a-command)\"\n")
	if _, err := s.Upload(context.Background(), deployment.Artifact{Content: content}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(client.input, content) {
		t.Fatal("upload altered the native file")
	}
	before := client.calls
	requireStage(t, s.LoadCandidate(context.Background(), "/tmp/foreign'; command"), deployment.StageLoad)
	if client.calls != before {
		t.Fatal("foreign file path reached the device")
	}
}

func TestCommitRejectsUnrepresentableTimeout(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Minute, time.Second, 90 * time.Second} {
		client := &fakeTransport{}
		s := newSession(client)
		s.loaded = true
		requireStage(t, s.CommitConfirmed(context.Background(), timeout), deployment.StageCommit)
		if client.calls != 0 {
			t.Fatal("invalid rollback window reached the device")
		}
	}
}

func TestCloseReleasesTransportAndPreservesCleanupFailure(t *testing.T) {
	cause := errors.New("cleanup failed")
	client := &fakeTransport{err: cause}
	s := newSession(client)
	for range 2 {
		err := s.Close(context.Background())
		requireStage(t, err, deployment.StageCleanup)
		if !errors.Is(err, cause) {
			t.Fatal("cleanup failure was lost")
		}
	}
	if client.calls != 1 || client.closes != 1 {
		t.Fatal("repeated Close repeated cleanup or failed to release the transport")
	}
	_, err := s.Upload(context.Background(), deployment.Artifact{})
	requireStage(t, err, deployment.StageUpload)
	if client.calls != 1 {
		t.Fatal("closed session contacted the device")
	}
}

func TestCommandDiagnosticsAreRetainedWithoutPrinting(t *testing.T) {
	cause := errors.New("native exit")
	client := &fakeTransport{err: cause, result: sshclient.Result{Stderr: []byte("sensitive-configuration")}}
	s := newSession(client)
	_, err := s.ReadActiveConfiguration(context.Background())
	requireStage(t, err, deployment.StageInspect)
	var command *CommandError
	if !errors.As(err, &command) || string(command.Result.Stderr) != "sensitive-configuration" || !errors.Is(err, cause) {
		t.Fatal("native diagnostics or cause were lost")
	}
	if strings.Contains(err.Error(), "sensitive-configuration") {
		t.Fatal("printable error exposed native diagnostics")
	}
}
