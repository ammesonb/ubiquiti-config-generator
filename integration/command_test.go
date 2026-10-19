//go:build integration

package integration

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/cli"
	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
	"github.com/ammesonb/ubiquiti-config-generator/internal/testlab"
	"github.com/ammesonb/ubiquiti-config-generator/internal/vyos"
	"golang.org/x/crypto/ssh"
)

func TestDeployCommand(t *testing.T) {
	requireTestTags(t, "deploy")
	if !*dockerLab {
		t.Skip("deployment command requires a disposable Docker router")
	}
	t.Parallel()
	lab := provisionDockerLab(t)
	cfg := lab.config
	ctx := context.Background()
	if err := testlab.Smoke(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	session, err := (vyos.Dialer{SSH: sshclient.Config{Address: cfg.Address, User: cfg.User, Auth: []ssh.AuthMethod{cfg.Auth}, HostKey: cfg.HostKey, HostKeyAlgorithms: cfg.HostKeyAlgorithms, CommandTimeout: time.Minute}}).Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	baseline, readErr := session.ReadActiveConfiguration(ctx)
	closeErr := session.Close(ctx)
	if readErr != nil || closeErr != nil {
		t.Fatalf("snapshot: %v; cleanup: %v", readErr, closeErr)
	}
	_, footer, ok := strings.Cut(string(baseline), "// Warning:")
	if !ok {
		t.Fatal("missing version metadata")
	}
	fixture := func(name string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join("lab", name))
		if err != nil {
			t.Fatal(err)
		}
		content = []byte(strings.ReplaceAll(strings.ReplaceAll(string(content), "{{SSH_KEY}}", lab.keyData), "{{CONFIG_VERSION}}", strings.TrimSpace("// Warning:"+footer)))
		file := filepath.Join(t.TempDir(), "config.boot")
		if err := os.WriteFile(file, content, 0o600); err != nil {
			t.Fatal(err)
		}
		return file
	}
	run := func(file string, input io.Reader, save bool) (int, string) {
		t.Helper()
		args := []string{"deploy", file, "--host", cfg.Address, "--user", cfg.User, "--identity", lab.identityPath, "--known-hosts", lab.knownHostsPath, "--host-key-algorithms", "ssh-ed25519", "--confirm-timeout", "1m", "--command-timeout", "1m"}
		if save {
			args = append(args, "--save")
		}
		var output bytes.Buffer
		code := cli.Run(ctx, args, cli.Streams{Input: input, Output: &output, Error: &output, Interactive: true})
		return code, output.String()
	}
	a, b := fixture("deploy-a.boot.tmpl"), fixture("deploy-b.boot.tmpl")
	if code, output := run(a, strings.NewReader("yes\n"), true); code != 0 {
		t.Fatalf("accept: %d %s", code, output)
	}
	accepted := activeConfig(t, cfg)
	saved := savedHash(t, cfg)
	if code, output := run(b, strings.NewReader("yes\n"), false); code != 0 {
		t.Fatalf("accept without save: %d %s", code, output)
	}
	if activeConfig(t, cfg) == accepted || savedHash(t, cfg) != saved {
		t.Fatal("confirmation without save did not preserve startup configuration")
	}
	stable := activeConfig(t, cfg)
	input, writer := io.Pipe()
	defer func() { _ = input.Close() }()
	defer func() { _ = writer.Close() }()
	started := time.Now()
	code, output := run(a, input, true)
	if code != 1 || !strings.Contains(output, "context deadline exceeded") {
		t.Fatalf("unanswered prompt: %d %s", code, output)
	}
	if time.Since(started) > 75*time.Second {
		t.Fatal("blocked input delayed return beyond confirmation window")
	}
	waitForRollback(t, cfg, stable)
	if savedHash(t, cfg) != saved {
		t.Fatal("unanswered prompt changed startup configuration")
	}
	code, output = run(fixture("invalid-load.boot"), strings.NewReader("yes\n"), true)
	if code != 1 || !strings.Contains(output, "deployment load failed") || strings.Contains(output, "Confirm these changes?") {
		t.Fatalf("invalid load: %d %s", code, output)
	}
}
