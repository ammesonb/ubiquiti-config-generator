package deployment

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type runnerSession struct {
	calls               []string
	fail                string
	failure             error
	pending             func(context.Context) (bool, error)
	cleanupErr          error
	cleanupContextError error
	cleanupDeadline     bool
}

func (s *runnerSession) step(name string) error {
	s.calls = append(s.calls, name)
	if s.fail == name {
		return s.failure
	}
	return nil
}
func (s *runnerSession) Upload(context.Context, Artifact) (RemoteFile, error) {
	return "uploaded", s.step("upload")
}
func (s *runnerSession) LoadCandidate(context.Context, RemoteFile) error { return s.step("load") }
func (s *runnerSession) CommitConfirmed(context.Context, time.Duration) error {
	return s.step("commit")
}
func (s *runnerSession) ReadActiveConfiguration(context.Context) (Configuration, error) {
	return nil, s.step("read")
}
func (s *runnerSession) ConfirmationPending(ctx context.Context) (bool, error) {
	if err := s.step("pending"); err != nil {
		return false, err
	}
	if s.pending != nil {
		return s.pending(ctx)
	}
	return true, nil
}
func (s *runnerSession) Confirm(context.Context) error { return s.step("confirm") }
func (s *runnerSession) Save(context.Context) error    { return s.step("save") }
func (s *runnerSession) Close(ctx context.Context) error {
	s.cleanupContextError = ctx.Err()
	_, s.cleanupDeadline = ctx.Deadline()
	_ = s.step("close")
	return s.cleanupErr
}

type runnerDialer struct {
	s   *runnerSession
	err error
}

func (d runnerDialer) Dial(context.Context) (Session, error) { return d.s, d.err }
func readyDecision(d Decision) Approval {
	return func(context.Context) <-chan Decision { ch := make(chan Decision, 1); ch <- d; return ch }
}
func runRequest() Request {
	return Request{Artifact: Artifact{Content: []byte("config")}, ConfirmTimeout: time.Minute, SaveAfterConfirm: true}
}
func testRunner(s *runnerSession) Runner {
	return Runner{Dialer: runnerDialer{s: s}, PollInterval: time.Millisecond, PollTimeout: 20 * time.Millisecond}
}

func TestRunnerSteps(t *testing.T) {
	failure := errors.New("failure")
	for _, tc := range []struct {
		name, fail string
		save       bool
		want       []string
		outcome    Outcome
	}{
		{"save", "", true, []string{"upload", "load", "commit", "pending", "pending", "confirm", "save", "close"}, Outcome{true, true, true}},
		{"no save", "", false, []string{"upload", "load", "commit", "pending", "pending", "confirm", "close"}, Outcome{true, true, false}},
		{"upload", "upload", true, []string{"upload", "close"}, Outcome{}},
		{"load", "load", true, []string{"upload", "load", "close"}, Outcome{}},
		{"commit", "commit", true, []string{"upload", "load", "commit", "close"}, Outcome{}},
		{"inspect", "pending", true, []string{"upload", "load", "commit", "pending", "close"}, Outcome{Applied: true}},
		{"confirm", "confirm", true, []string{"upload", "load", "commit", "pending", "pending", "confirm", "close"}, Outcome{Applied: true}},
		{"save fails", "save", true, []string{"upload", "load", "commit", "pending", "pending", "confirm", "save", "close"}, Outcome{Applied: true, Confirmed: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &runnerSession{fail: tc.fail, failure: failure}
			r := testRunner(s)
			r.PollInterval = time.Hour
			req := runRequest()
			req.SaveAfterConfirm = tc.save
			got, err := r.Deploy(context.Background(), req, readyDecision(Decision{Accepted: true}))
			if (tc.fail != "") != errors.Is(err, failure) {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.outcome || !reflect.DeepEqual(s.calls, tc.want) {
				t.Fatalf("outcome=%+v calls=%v", got, s.calls)
			}
			if s.cleanupContextError != nil || !s.cleanupDeadline {
				t.Fatal("cleanup needs an independent bounded context")
			}
		})
	}
	t.Run("dial", func(t *testing.T) {
		r := Runner{Dialer: runnerDialer{err: failure}}
		_, err := r.Deploy(context.Background(), runRequest(), readyDecision(Decision{}))
		if !errors.Is(err, failure) {
			t.Fatal(err)
		}
	})
}

func TestRunnerInterruptsApproval(t *testing.T) {
	for _, reason := range []string{"decline", "input error", "closed input", "initial expiry", "poll expiry", "acceptance expiry", "poll timeout", "cancel", "deadline"} {
		t.Run(reason, func(t *testing.T) {
			s := &runnerSession{}
			r := testRunner(s)
			req := runRequest()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			inputErr := errors.New("input failure")
			approvalCalled := false
			approval := func(context.Context) <-chan Decision {
				approvalCalled = true
				ch := make(chan Decision, 1)
				switch reason {
				case "decline":
					ch <- Decision{}
				case "input error":
					ch <- Decision{Err: inputErr}
				case "closed input":
					close(ch)
				case "acceptance expiry":
					ch <- Decision{Accepted: true}
				case "cancel":
					cancel()
				}
				return ch
			}
			want := ErrConfirmationEnded
			count := 0
			s.pending = func(ctx context.Context) (bool, error) {
				count++
				switch reason {
				case "initial expiry":
					return false, nil
				case "poll expiry", "acceptance expiry":
					return count == 1, nil
				case "poll timeout":
					if count > 1 {
						<-ctx.Done()
						return false, ctx.Err()
					}
				}
				return true, nil
			}
			switch reason {
			case "decline":
				want = ErrDeclined
			case "input error":
				want = inputErr
			case "closed input":
				want = ErrApprovalClosed
			case "cancel":
				want = context.Canceled
			case "deadline":
				req.ConfirmTimeout = 30 * time.Millisecond
				want = context.DeadlineExceeded
			case "poll timeout":
				want = context.DeadlineExceeded
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				_, err := r.Deploy(ctx, req, approval)
				if !errors.Is(err, want) {
					t.Errorf("got %v, want %v", err, want)
				}
			}()
			select {
			case <-done:
			case <-time.After(time.Second):
				cancel()
				<-done
				t.Fatal("approval was not interrupted")
			}
			for _, call := range s.calls {
				if call == "confirm" || call == "save" {
					t.Fatalf("unexpected %s", call)
				}
			}
			if reason == "initial expiry" && approvalCalled {
				t.Fatal("prompt started after expiry")
			}
			if s.cleanupContextError != nil {
				t.Fatal("cleanup inherited cancellation")
			}
		})
	}
}

func TestRunnerPreservesCleanupFailure(t *testing.T) {
	first := errors.New("upload failure")
	last := errors.New("cleanup failure")
	s := &runnerSession{fail: "upload", failure: first, cleanupErr: last}
	_, err := testRunner(s).Deploy(context.Background(), runRequest(), readyDecision(Decision{}))
	if !errors.Is(err, first) || !errors.Is(err, last) {
		t.Fatalf("lost error: %v", err)
	}
}
