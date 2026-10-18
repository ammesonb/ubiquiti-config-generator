//go:build integration

package integration

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
	"github.com/ammesonb/ubiquiti-config-generator/internal/testlab"
	"github.com/ammesonb/ubiquiti-config-generator/internal/vyos"
	"golang.org/x/crypto/ssh"
)

func TestVyOSSession(t *testing.T) {
	requireTestTags(t, "deploy")
	if !*dockerLab {
		t.Skip("VyOS adapter tests require a disposable Docker router")
	}
	t.Parallel()
	lab := provisionDockerLab(t)
	cfg := lab.config
	ctx := context.Background()
	if err := testlab.Smoke(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	dialer := vyos.Dialer{SSH: sshclient.Config{
		Address: cfg.Address, User: cfg.User, Auth: []ssh.AuthMethod{cfg.Auth},
		HostKey: cfg.HostKey, HostKeyAlgorithms: cfg.HostKeyAlgorithms,
		CommandTimeout: time.Minute,
	}}
	open := func(t *testing.T) deployment.Session {
		t.Helper()
		s, err := dialer.Dial(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := s.Close(context.Background()); err != nil {
				t.Errorf("close adapter session: %v", err)
			}
		})
		return s
	}
	initial := open(t)
	baseline, err := initial.ReadActiveConfiguration(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, footer, ok := strings.Cut(string(baseline), "// Warning:")
	if !ok || !strings.Contains(footer, "// vyos-config-version:") {
		t.Fatal("adapter snapshot lacks native version metadata")
	}
	if err := initial.Close(ctx); err != nil {
		t.Fatal(err)
	}
	fixture := func(name string) deployment.Artifact {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("lab", name))
		if err != nil {
			t.Fatal(err)
		}
		return deployment.Artifact{Content: []byte(strings.ReplaceAll(strings.ReplaceAll(string(data),
			"{{SSH_KEY}}", lab.keyData), "{{CONFIG_VERSION}}", strings.TrimSpace("// Warning:"+footer)))}
	}
	a, b := fixture("deploy-a.boot.tmpl"), fixture("deploy-b.boot.tmpl")
	closeAndCheck := func(t *testing.T, s deployment.Session, file deployment.RemoteFile) {
		t.Helper()
		if err := s.Close(ctx); err != nil {
			t.Fatal(err)
		}
		dir := path.Dir(string(file))
		if _, err := docker("exec", lab.name, "test", "!", "-e", dir); err != nil {
			t.Fatal("Close left uploaded files behind")
		}
		remaining, err := docker("exec", lab.name, "find", "/opt/vyatta/config/tmp", "-mindepth", "1", "-maxdepth", "1")
		if err != nil || len(bytes.TrimSpace(remaining)) != 0 {
			t.Fatal("Close left native candidate state behind")
		}
	}
	stageFailure := func(t *testing.T, err error, expected deployment.Stage) {
		t.Helper()
		var failure *deployment.StageError
		if !errors.As(err, &failure) || failure.Stage != expected {
			t.Fatalf("expected %s failure, got %v", expected, err)
		}
	}
	var confirmed, saved string
	if !t.Run("replace_confirm_save", func(t *testing.T) {
		for _, artifact := range []deployment.Artifact{a, b} {
			s := open(t)
			before, err := s.ReadActiveConfiguration(ctx)
			if err != nil {
				t.Fatal(err)
			}
			bootBefore := savedHash(t, cfg)
			file, err := s.Upload(ctx, artifact)
			if err != nil {
				t.Fatal(err)
			}
			for name, mode := range map[string]string{string(file): "600", path.Dir(string(file)): "700"} {
				out, err := docker("exec", lab.name, "stat", "--format=%a", name)
				if err != nil || strings.TrimSpace(string(out)) != mode {
					t.Fatal("uploaded configuration is not private to the deployment user")
				}
			}
			if err := s.LoadCandidate(ctx, file); err != nil {
				t.Fatal(err)
			}
			active, err := s.ReadActiveConfiguration(ctx)
			if err != nil || !bytes.Equal(active, before) {
				t.Fatal("loading changed active configuration or ReadActiveConfiguration returned the candidate")
			}
			if err := s.CommitConfirmed(ctx, time.Minute); err != nil {
				t.Fatal(err)
			}
			if pending, err := s.ConfirmationPending(ctx); err != nil || !pending {
				t.Fatal("adapter did not report pending confirmation")
			}
			stageFailure(t, s.Save(ctx), deployment.StageSave)
			if savedHash(t, cfg) != bootBefore {
				t.Fatal("pending deployment changed persistent configuration")
			}
			if err := s.Confirm(ctx); err != nil {
				t.Fatal(err)
			}
			if pending, err := s.ConfirmationPending(ctx); err != nil || pending {
				t.Fatal("adapter did not verify confirmation")
			}
			if err := s.Save(ctx); err != nil {
				t.Fatal(err)
			}
			active, err = s.ReadActiveConfiguration(ctx)
			if err != nil || string(active) != commandOK(t, cfg, "cat /config/config.boot", "") {
				t.Fatal("saved configuration differs from the adapter's active snapshot")
			}
			live := activeConfig(t, cfg)
			if bytes.Equal(artifact.Content, a.Content) {
				requireSetting(t, live, "set system static-host-mapping host-name removed.example inet '192.0.2.20'")
			} else {
				requireSetting(t, live, "set system static-host-mapping host-name retained.example inet '198.51.100.10'")
				requireSetting(t, live, "set system static-host-mapping host-name added.example inet '203.0.113.10'")
				if strings.Contains(live, "removed.example") || strings.Contains(live, "192.0.2.10") {
					t.Fatal("adapter did not replace omitted or changed settings")
				}
			}
			closeAndCheck(t, s, file)
		}
		confirmed, saved = activeConfig(t, cfg), savedHash(t, cfg)
	}) {
		return
	}
	for _, scenario := range []struct {
		name, file string
		stage      deployment.Stage
	}{
		{"invalid_load", "invalid-load.boot", deployment.StageLoad},
		{"invalid_commit", "invalid-commit.boot.tmpl", deployment.StageCommit},
	} {
		if !t.Run(scenario.name, func(t *testing.T) {
			s := open(t)
			file, err := s.Upload(ctx, fixture(scenario.file))
			if err != nil {
				t.Fatal(err)
			}
			err = s.LoadCandidate(ctx, file)
			if scenario.stage == deployment.StageLoad {
				stageFailure(t, err, deployment.StageLoad)
			} else if err != nil {
				t.Fatal(err)
			}
			stageFailure(t, s.CommitConfirmed(ctx, time.Minute), deployment.StageCommit)
			stageFailure(t, s.Confirm(ctx), deployment.StageConfirm)
			stageFailure(t, s.Save(ctx), deployment.StageSave)
			closeAndCheck(t, s, file)
			if scenario.stage == deployment.StageCommit {
				requireTimer(t, cfg, true)
				t.Log("waiting for native recovery after rejected commit")
				waitForRollback(t, cfg, confirmed)
			}
			assertUnchanged(t, cfg, confirmed, saved)
		}) {
			return
		}
	}
	t.Run("close_preserves_rollback", func(t *testing.T) {
		s := open(t)
		file, err := s.Upload(ctx, a)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.LoadCandidate(ctx, file); err != nil {
			t.Fatal(err)
		}
		if err := s.CommitConfirmed(ctx, time.Minute); err != nil {
			t.Fatal(err)
		}
		closeAndCheck(t, s, file)
		requireTimer(t, cfg, true)
		t.Log("waiting for native rollback after closing the deployment session")
		waitForRollback(t, cfg, confirmed)
		assertUnchanged(t, cfg, confirmed, saved)
	})
}
