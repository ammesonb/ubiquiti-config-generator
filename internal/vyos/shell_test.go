package vyos

import (
	"bytes"
	"context"
	"errors"
	"math/rand/v2"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
)

func runBash(t *testing.T, script string) ([]byte, []byte) {
	t.Helper()
	bash := "bash"
	if runtime.GOOS == "windows" {
		t.Skip("requires Linux Bash byte semantics, run with make test-docker")
	}
	if _, err := exec.LookPath(bash); err != nil {
		t.Skip("shell behavior tests require Bash")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bash, "--noprofile", "--norc", "-s")
	cmd.Dir = t.TempDir()
	cmd.Stdin = strings.NewReader(script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Bash failed: %v\n%s", err, stderr.Bytes())
	}
	return stdout.Bytes(), stderr.Bytes()
}

func TestShellArgumentRoundTrip(t *testing.T) {
	values := []string{
		"", "plain", "two words", "\t\r\n", "'", "\"", "\\", "a'b\"c",
		"$HOME ${HOME} $((1+2))", "$(printf INJECTED >&2)", "`printf INJECTED >&2`",
		"'; printf INJECTED >&2; #", "\"; printf INJECTED >&2; #",
		"a && printf INJECTED >&2", "a || printf INJECTED >&2", "a | cat",
		"> file < input 2>&1", "* ? [abc] {a,b} ~", "--help", "# comment",
		"line one\n$(printf INJECTED >&2)\nline three", "café 路由器 🚀",
	}
	// NUL is excluded because a shell argument cannot represent it.
	for b := 1; b < 256; b++ {
		values = append(values, string([]byte{byte(b)}))
	}
	rng := rand.New(rand.NewPCG(1, 2))
	for range 1000 {
		value := make([]byte, rng.IntN(128))
		for i := range value {
			value[i] = byte(1 + rng.IntN(255))
		}
		values = append(values, string(value))
	}
	var script strings.Builder
	var expected bytes.Buffer
	for _, value := range values {
		script.WriteString("printf '%s\\0' " + quoteShellArgument(value) + "\n")
		expected.WriteString(value)
		expected.WriteByte(0)
	}
	stdout, stderr := runBash(t, script.String())
	if !bytes.Equal(stdout, expected.Bytes()) || len(stderr) != 0 {
		parts := bytes.Split(stdout, []byte{0})
		for i, value := range values {
			if i >= len(parts) {
				t.Fatalf("missing argument %d", i)
			}
			if string(parts[i]) != value {
				t.Fatalf("argument %d changed: want %q, got %q; stderr %q", i, value, parts[i], stderr)
			}
		}
		t.Fatalf("unexpected shell output: stderr %q", stderr)
	}
}

func TestSessionSetupOutputDoesNotEnterConfiguration(t *testing.T) {
	s := newSession(nil)
	s.configurationSessionPID = "123"
	script := `cli-shell-api() {
  if [ "$1" = getSessionEnv ]; then printf 'export TEST_SESSION=ready'; else echo session-noise; fi
}
source() { echo helper-noise; }
` + s.withConfigurationSession("printf '%s' \"configuration:$TEST_SESSION\"\n")
	stdout, stderr := runBash(t, script)
	if string(stdout) != "configuration:ready" {
		t.Fatalf("setup contaminated the configuration: %q", stdout)
	}
	if !strings.Contains(string(stderr), "session-noise") || !strings.Contains(string(stderr), "helper-noise") {
		t.Fatal("setup diagnostics were lost")
	}
}

func TestSessionReadinessTimeout(t *testing.T) {
	s := newSession(nil)
	s.shellDone = make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err := stageError(deployment.StageAwaitReady, s.waitForSessionReady(ctx, make(chan string)))
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "wait for session readiness") || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("readiness timeout lost its operation or cause: %v", err)
	}
}
