package deployment

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDeclined          = errors.New("changes left unconfirmed")
	ErrConfirmationEnded = errors.New("confirmation is no longer pending")
	ErrApprovalClosed    = errors.New("confirmation input closed without a decision")
)

// Decision conveys acceptance or an input failure without blocking device monitoring.
type Decision struct {
	Accepted bool
	Err      error
}

// Approval starts collecting a decision once changes are active. It must return promptly.
type Approval func(context.Context) <-chan Decision

// Outcome records completed steps so a later failure does not hide confirmation or saving.
type Outcome struct {
	Applied   bool
	Confirmed bool
	Saved     bool
}

// Runner orchestrates operations sequentially while monitoring the confirmation window.
type Runner struct {
	Dialer         Dialer
	PollInterval   time.Duration
	PollTimeout    time.Duration
	CleanupTimeout time.Duration
}

func (r Runner) withDefaults() Runner {
	if r.PollInterval == 0 {
		r.PollInterval = time.Second
	}
	if r.PollTimeout == 0 {
		r.PollTimeout = 5 * time.Second
	}
	if r.CleanupTimeout == 0 {
		r.CleanupTimeout = 10 * time.Second
	}
	return r
}

// Deploy applies a file and waits for explicit acceptance before confirming or saving it.
func (r Runner) Deploy(ctx context.Context, request Request, approval Approval) (outcome Outcome, err error) {
	if r.Dialer == nil || approval == nil {
		return outcome, errors.New("a device dialer and confirmation input are required")
	}
	if request.ConfirmTimeout <= 0 || r.PollInterval < 0 || r.PollTimeout < 0 || r.CleanupTimeout < 0 {
		return outcome, errors.New("deployment timeouts must be positive")
	}
	r = r.withDefaults()
	session, err := r.Dialer.Dial(ctx)
	if err != nil {
		return outcome, stepError(StageConnect, err)
	}
	defer func() {
		// Cancellation must not prevent an attempt to remove uploaded files.
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.CleanupTimeout)
		defer cancel()
		err = errors.Join(err, stepError(StageCleanup, session.Close(cleanup)))
	}()
	file, err := session.Upload(ctx, request.Artifact)
	if err != nil {
		return outcome, stepError(StageUpload, err)
	}
	if err := session.LoadCandidate(ctx, file); err != nil {
		return outcome, stepError(StageLoad, err)
	}
	// Start before commit and reserve time for confirmation to reduce the race with device rollback.
	reserve := min(5*time.Second, request.ConfirmTimeout/10)
	deadline := time.Now().Add(request.ConfirmTimeout - reserve)
	if err := session.CommitConfirmed(ctx, request.ConfirmTimeout); err != nil {
		return outcome, stepError(StageCommit, err)
	}
	outcome.Applied = true
	wait, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	if err := r.ensurePending(wait, session); err != nil {
		return outcome, err
	}
	if err := r.awaitApproval(wait, session, approval(wait)); err != nil {
		return outcome, err
	}
	if err := session.Confirm(wait); err != nil {
		return outcome, stepError(StageConfirm, err)
	}
	outcome.Confirmed = true
	cancel()
	if request.SaveAfterConfirm {
		if err := session.Save(ctx); err != nil {
			return outcome, stepError(StageSave, err)
		}
		outcome.Saved = true
	}
	return outcome, nil
}

func (r Runner) awaitApproval(ctx context.Context, session Session, decisions <-chan Decision) error {
	ticker := time.NewTicker(r.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return stepError(StageApproval, ctx.Err())
		case decision, open := <-decisions:
			if err := ctx.Err(); err != nil {
				return stepError(StageApproval, err)
			}
			if !open {
				return stepError(StageApproval, ErrApprovalClosed)
			}
			if decision.Err != nil {
				return stepError(StageApproval, decision.Err)
			}
			if !decision.Accepted {
				return stepError(StageApproval, ErrDeclined)
			}
			// Acceptance may arrive just as rollback expires, so polling alone is insufficient.
			return r.ensurePending(ctx, session)
		case <-ticker.C:
			if err := r.ensurePending(ctx, session); err != nil {
				return err
			}
		}
	}
}

func (r Runner) ensurePending(ctx context.Context, session Session) error {
	if err := ctx.Err(); err != nil {
		return stepError(StageApproval, err)
	}
	poll, cancel := context.WithTimeout(ctx, r.PollTimeout)
	defer cancel()
	pending, err := session.ConfirmationPending(poll)
	if err != nil {
		return stepError(StageInspect, err)
	}
	if err := ctx.Err(); err != nil {
		return stepError(StageApproval, err)
	}
	if !pending {
		return stepError(StageApproval, ErrConfirmationEnded)
	}
	return nil
}

func stepError(stage Stage, err error) error {
	if err == nil {
		return nil
	}
	var existing *StageError
	if errors.As(err, &existing) && (stage != StageCleanup || existing.Stage == StageCleanup) {
		return err
	}
	return &StageError{Stage: stage, Err: err}
}
