package vyos

import (
	"bufio"
	"context"
	"crypto/rand"
	_ "embed"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
	"golang.org/x/crypto/ssh"
)

type Dialer struct {
	SSH sshclient.Config
}

var _ deployment.Dialer = Dialer{}
var _ deployment.Session = (*session)(nil)

// Dial connects to VyOS and prepares a configuration session for staging changes.
func (d Dialer) Dial(ctx context.Context) (deployment.Session, error) {
	cfg := d.SSH.WithDefaults()
	client, err := sshclient.Dial(ctx, cfg)
	if err != nil {
		return nil, stageError(deployment.StageConnect, err)
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.CommandTimeout)
	defer cancel()
	// Until preparation succeeds, this connection belongs to Dial and can be discarded on cancellation.
	// After the session is ready, unregister this failsafe to ensure we do NOT close the connection.
	skipClose := context.AfterFunc(ctx, func() { _ = client.Close() })
	s := newSession(client)
	err = s.startConfigurationSession(ctx, client)
	unregistered := skipClose()
	if err == nil && (!unregistered || ctx.Err() != nil) {
		err = stageError(deployment.StageAwaitReady, ctx.Err())
	}
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	return s, nil
}

type transport interface {
	Run(context.Context, string, []byte) (sshclient.Result, error)
	Close() error
}

type session struct {
	client                  transport
	id                      string
	files                   map[deployment.RemoteFile]bool
	loaded                  bool
	committed               bool
	confirmed               bool
	closed                  bool
	closeErr                error
	configurationSessionPID string
	shell                   *ssh.Session
	shellInput              io.WriteCloser
	shellDone               chan struct{}
	shellErr                error
}

//go:embed configuration_session.vbash
var configurationSessionScript string

func (s *session) startConfigurationSession(ctx context.Context, client *sshclient.Client) (err error) {
	stage := deployment.StagePrepare
	defer func() {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		err = stageError(stage, err)
	}()
	shell, err := client.NewSession(ctx)
	if err != nil {
		return err
	}
	s.shell = shell
	s.shellInput, err = shell.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := shell.StdoutPipe()
	if err != nil {
		return err
	}
	shell.Stderr = io.Discard
	if err := shell.Start("/bin/vbash -c " + quoteShellArgument(configurationSessionScript) + " -- " + quoteShellArgument(s.directory())); err != nil {
		return err
	}
	ready := make(chan string, 1)
	go func() {
		// Keep draining after readiness so later shell output cannot block the persistent process.
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if id, ok := strings.CutPrefix(scanner.Text(), "UBQ_READY:"); ok {
				select {
				case ready <- id:
				default:
				}
			}
		}
	}()
	s.shellDone = make(chan struct{})
	go func() { s.shellErr = shell.Wait(); close(s.shellDone) }()
	stage = deployment.StageAwaitReady
	return s.waitForSessionReady(ctx, ready)
}

func (s *session) waitForSessionReady(ctx context.Context, ready <-chan string) error {
	select {
	case id := <-ready:
		pid, err := strconv.Atoi(id)
		if err != nil || pid <= 0 {
			return errors.New("invalid configuration session process ID")
		}
		s.configurationSessionPID = id
		return nil
	case <-s.shellDone:
		return errors.Join(errors.New("configuration shell exited before readiness"), s.shellErr)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func newSession(client transport) *session {
	return &session{client: client, id: "ubq-" + rand.Text(), files: make(map[deployment.RemoteFile]bool)}
}

func (s *session) directory() string { return "/tmp/" + s.id }

// Upload stores the supplied configuration unchanged in a private temporary file.
func (s *session) Upload(ctx context.Context, artifact deployment.Artifact) (deployment.RemoteFile, error) {
	s.loaded = false
	if s.closed {
		return "", stageError(deployment.StageUpload, errors.New("session is closed"))
	}
	file := deployment.RemoteFile(s.directory() + "/candidate-" + strconv.Itoa(len(s.files)) + ".boot")
	// Track the path before writing so cleanup can remove a partial upload.
	s.files[file] = false
	_, err := s.execute(ctx, deployment.StageUpload, "umask 077; cat > "+quoteShellArgument(string(file)), artifact.Content)
	if err != nil {
		return "", err
	}
	s.files[file] = true
	return file, nil
}

// LoadCandidate replaces staged settings with the uploaded file without changing the running configuration.
func (s *session) LoadCandidate(ctx context.Context, file deployment.RemoteFile) error {
	s.loaded, s.committed, s.confirmed = false, false, false
	if !s.files[file] {
		return stageError(deployment.StageLoad, errors.New("file was not uploaded by this session"))
	}
	_, err := s.run(ctx, deployment.StageLoad, s.withConfigurationSession(requireNoPendingConfirmation+
		"load "+quoteShellArgument(string(file))+" || builtin exit 1\n"))
	s.loaded = err == nil
	return err
}

// CommitConfirmed applies staged changes and arms rollback unless they are confirmed in time.
func (s *session) CommitConfirmed(ctx context.Context, timeout time.Duration) error {
	if timeout < time.Minute || timeout%time.Minute != 0 {
		return stageError(deployment.StageCommit, errors.New("VyOS confirmation timeout must be a positive whole number of minutes"))
	}
	if !s.loaded {
		return stageError(deployment.StageCommit, errors.New("no successfully loaded candidate"))
	}
	s.loaded, s.committed, s.confirmed = false, false, false
	_, err := s.run(ctx, deployment.StageCommit, s.withConfigurationSession(requireNoPendingConfirmation+
		"commit-confirm "+strconv.FormatInt(int64(timeout/time.Minute), 10)+" no-prompt || builtin exit 1\n"+
		// commit-confirm can return zero after rejecting a candidate, so verify
		// that it consumed the staged changes and armed rollback.
		"cli-shell-api sessionChanged\n[ $? -eq 1 ] || builtin exit 1\n"+
		"systemctl is-active --quiet commit-confirm.timer || builtin exit 1\n"))
	s.committed = err == nil
	return err
}

// ReadActiveConfiguration snapshots the running settings without changing the saved startup configuration.
func (s *session) ReadActiveConfiguration(ctx context.Context) (deployment.Configuration, error) {
	file := s.directory() + "/active.boot"
	result, err := s.run(ctx, deployment.StageInspect, s.withConfigurationSession(
		"rm -f -- "+quoteShellArgument(file)+" || builtin exit 1\n"+
			saveActiveConfiguration(file)+" >/dev/null || builtin exit 1\n"+
			"sudo cat -- "+quoteShellArgument(file)+" || builtin exit 1\n"))
	if err != nil {
		return nil, err
	}
	return deployment.Configuration(result.Stdout), nil
}

// ConfirmationPending checks whether VyOS is waiting for confirmation before rolling back.
func (s *session) ConfirmationPending(ctx context.Context) (bool, error) {
	result, err := s.run(ctx, deployment.StageInspect,
		"state=$(systemctl is-active commit-confirm.timer)\n"+
			"case \"$state\" in\nactive) printf 'pending';;\ninactive) printf 'clear';;\n*) builtin exit 1;;\nesac\n")
	if err != nil {
		return false, err
	}
	switch string(result.Stdout) {
	case "pending":
		return true, nil
	case "clear":
		return false, nil
	default:
		return false, stageError(deployment.StageInspect, errors.New("unrecognized confirmation state"))
	}
}

// Confirm accepts this session's committed changes by disabling their rollback timer.
func (s *session) Confirm(ctx context.Context) error {
	if !s.committed {
		return stageError(deployment.StageConfirm, errors.New("no successful commit awaiting confirmation in this session"))
	}
	s.committed, s.confirmed = false, false
	_, err := s.run(ctx, deployment.StageConfirm, s.withConfigurationSession(
		"systemctl is-active --quiet commit-confirm.timer || builtin exit 1\n"+
			"confirm || builtin exit 1\n"+requireNoPendingConfirmation))
	s.confirmed = err == nil
	return err
}

// Save persists the confirmed configuration for use after restart.
func (s *session) Save(ctx context.Context) error {
	if !s.confirmed {
		return stageError(deployment.StageSave, errors.New("configuration has not been confirmed by this session"))
	}
	_, err := s.run(ctx, deployment.StageSave, s.withConfigurationSession(requireNoPendingConfirmation+
		saveActiveConfiguration("/config/config.boot")+" || builtin exit 1\n"))
	return err
}

// Close releases session resources without accepting or persisting pending changes.
func (s *session) Close(ctx context.Context) error {
	if s.closed {
		return s.closeErr
	}
	paths := []string{quoteShellArgument(s.directory() + "/active.boot")}
	for file := range s.files {
		paths = append(paths, quoteShellArgument(string(file)))
	}
	_, cleanupErr := s.run(ctx, deployment.StageCleanup, s.configurationEnvironment()+
		"status=0\ncli-shell-api teardownSession || status=1\n"+
		"rm -f -- "+strings.Join(paths, " ")+" || status=1\n"+
		"rmdir -- "+quoteShellArgument(s.directory())+" || status=1\nbuiltin exit \"$status\"\n")
	s.closed = true
	var shellErr error
	if s.shell != nil {
		_ = s.shellInput.Close()
		wait, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		select {
		case <-s.shellDone:
			shellErr = s.shellErr
		case <-wait.Done():
			shellErr = wait.Err()
		}
	}
	s.closeErr = stageError(deployment.StageCleanup, errors.Join(cleanupErr, shellErr, s.client.Close()))
	return s.closeErr
}

func (s *session) withConfigurationSession(command string) string {
	// Set the session environment before sourcing helpers that make its variables read-only.
	// Route setup output to stderr so it cannot contaminate the command's stdout.
	return "{\n" + s.configurationEnvironment() + "cli-shell-api inSession || builtin exit 1\n" +
		"source /opt/vyatta/etc/functions/script-template || builtin exit 1\n} >&2\n" + command
}

func (s *session) configurationEnvironment() string {
	return "session_env=$(cli-shell-api getSessionEnv " + quoteShellArgument(s.configurationSessionPID) + ") || builtin exit 1\n" +
		"eval \"$session_env\"\n"
}

func (s *session) run(ctx context.Context, stage deployment.Stage, script string) (sshclient.Result, error) {
	return s.execute(ctx, stage, "/bin/bash -s", []byte(script))
}

func (s *session) execute(ctx context.Context, stage deployment.Stage, command string, input []byte) (sshclient.Result, error) {
	if s.closed {
		return sshclient.Result{}, stageError(stage, errors.New("session is closed"))
	}
	if stage != deployment.StageCleanup && stage != deployment.StageInspect && s.shellDone != nil {
		select {
		case <-s.shellDone:
			return sshclient.Result{}, stageError(stage, errors.Join(errors.New("configuration shell exited"), s.shellErr))
		default:
		}
	}
	result, err := s.client.Run(ctx, command, input)
	if err != nil {
		err = &CommandError{Result: result, Err: err}
	}
	return result, stageError(stage, err)
}

const requireNoPendingConfirmation = "state=$(systemctl is-active commit-confirm.timer)\n[ \"$state\" = inactive ] || builtin exit 1\n"

func saveActiveConfiguration(path string) string {
	// Call the save script directly because the shell wrapper can mask its failure with successful bookkeeping.
	return "sudo sg vyattacfg -c " + quoteShellArgument("umask 077; /usr/libexec/vyos/vyos-save-config.py "+quoteShellArgument(path))
}

func quoteShellArgument(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func stageError(stage deployment.Stage, err error) error {
	if err == nil {
		return nil
	}
	return &deployment.StageError{Stage: stage, Err: err}
}
