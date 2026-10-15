//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/testlab"
	"golang.org/x/crypto/ssh"
)

// Mutating tests have no external-endpoint mode: each run owns its router.
func TestDeploy(t *testing.T) {
	requireTestTags(t, "deploy")
	if !*dockerLab {
		t.Skip("deployment scenarios require a disposable Docker router")
	}
	t.Parallel()
	lab := provisionDockerLab(t)
	cfg := lab.config
	if err := testlab.Smoke(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	version := commandOK(t, cfg, "/bin/vbash -s", configSession+"save /tmp/ubq-baseline.boot || builtin exit 1\ntail -n 3 /tmp/ubq-baseline.boot\n")
	_, footer, ok := strings.Cut(version, "// Warning:")
	if !ok || !strings.Contains(footer, "// vyos-config-version:") {
		t.Fatal("native save did not include config version metadata")
	}
	version = "// Warning:" + footer
	keyLine := lab.keyData
	fixture := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join("lab", name))
		if err != nil {
			t.Fatal(err)
		}
		return strings.ReplaceAll(strings.ReplaceAll(string(data), "{{SSH_KEY}}", keyLine), "{{CONFIG_VERSION}}", strings.TrimSpace(version))
	}
	a, b := fixture("deploy-a.boot.tmpl"), fixture("deploy-b.boot.tmpl")
	var confirmed, saved, deployedA string

	if !t.Run("replacement_confirmation_and_save", func(t *testing.T) {
		for _, desired := range []struct{ data, retained string }{
			{a, "192.0.2.10"},
			{b, "198.51.100.10"},
		} {
			bootBefore := savedHash(t, cfg)
			out, err := deployFile(t, cfg, desired.data)
			if err != nil {
				t.Fatalf("deployment failed at %s: %v", deployStage(out), err)
			}
			if err := testlab.Smoke(context.Background(), cfg); err != nil {
				t.Fatal(err)
			}
			active := activeConfig(t, cfg)
			requireSetting(t, active, "set system static-host-mapping host-name retained.example inet '"+desired.retained+"'")
			if desired.data == a {
				requireSetting(t, active, "set system static-host-mapping host-name removed.example inet '192.0.2.20'")
				deployedA = active
			} else {
				requireSetting(t, active, "set system static-host-mapping host-name added.example inet '203.0.113.10'")
				if strings.Contains(active, "removed.example") || strings.Contains(active, "192.0.2.10") {
					t.Fatal("full-file replacement retained settings omitted or changed in B")
				}
			}
			requireTimer(t, cfg, true)
			if savedHash(t, cfg) != bootBefore {
				t.Fatal("pending deployment changed saved boot configuration")
			}
			confirmAndSave(t, cfg)
			requireTimer(t, cfg, false)
			// Native serialization must match the saved boot file after confirmation.
			commandOK(t, cfg, "/bin/vbash -s", configSession+"save /tmp/ubq-active.boot || builtin exit 1\ncmp /tmp/ubq-active.boot /config/config.boot\n")
		}
		confirmed, saved = activeConfig(t, cfg), savedHash(t, cfg)
	}) {
		return
	}

	if !t.Run("upload_failure", func(t *testing.T) {
		out, err := deployFileAt(t, cfg, b, "/proc/ubq-config.boot")
		requireRemoteFailure(t, err)
		if deployStage(out) != "upload" {
			t.Fatal("upload failure did not stop the deployment before load")
		}
		requireTimer(t, cfg, false)
		assertUnchanged(t, cfg, confirmed, saved)
	}) {
		return
	}

	for _, scenario := range []struct{ name, file, stage, diagnostic string }{
		{"invalid_load", "invalid-load.boot", "load", ""},
		{"invalid_commit", "invalid-commit.boot.tmpl", "commit", "stop address must be greater or equal"},
	} {
		if !t.Run(scenario.name, func(t *testing.T) {
			out, err := deployFile(t, cfg, fixture(scenario.file))
			requireRemoteFailure(t, err)
			if deployStage(out) != scenario.stage {
				t.Fatalf("expected %s failure, got %s", scenario.stage, deployStage(out))
			}
			if scenario.diagnostic != "" && !strings.Contains(out, scenario.diagnostic) {
				t.Fatal("native commit validation did not report the expected DHCP range error")
			}
			if scenario.stage == "commit" {
				if savedHash(t, cfg) != saved {
					t.Fatal("failed commit changed the saved boot configuration")
				}
				requireTimer(t, cfg, true)
				t.Log("native commit rejected the candidate but armed rollback; waiting for expiry")
				waitForRollback(t, cfg, confirmed)
				assertUnchanged(t, cfg, confirmed, saved)
			} else {
				assertUnchanged(t, cfg, confirmed, saved)
			}
			requireTimer(t, cfg, false)
		}) {
			return
		}
	}

	if !t.Run("unconfirmed_rollback", func(t *testing.T) {
		out, err := deployFile(t, cfg, a)
		if err != nil {
			t.Fatalf("deployment failed at %s: %v", deployStage(out), err)
		}
		if err := testlab.Smoke(context.Background(), cfg); err != nil {
			t.Fatal(err)
		}
		if activeConfig(t, cfg) != deployedA {
			t.Fatal("unconfirmed deployment did not apply the complete A snapshot")
		}
		requireTimer(t, cfg, true)
		if savedHash(t, cfg) != saved {
			t.Fatal("unconfirmed deployment changed saved boot configuration")
		}
		t.Log("A is active and unconfirmed; waiting for native one-minute rollback to B")
		waitForRollback(t, cfg, confirmed)
		if err := testlab.Smoke(context.Background(), cfg); err != nil {
			t.Fatal(err)
		}
		assertUnchanged(t, cfg, confirmed, saved)
		requireTimer(t, cfg, false)
	}) {
		return
	}

	t.Run("saved_configuration_survives_restart", func(t *testing.T) {
		if output, err := docker("restart", lab.name); err != nil {
			t.Fatalf("restart lab: %v: %s", err, output)
		}
		port, err := docker("port", lab.name, "22/tcp")
		if err != nil {
			t.Fatalf("inspect restarted lab SSH port: %v", err)
		}
		cfg.Address = strings.TrimSpace(string(port))
		if !strings.HasPrefix(cfg.Address, "127.0.0.1:") {
			t.Fatal("restarted lab SSH port is not on loopback")
		}
		if err := testlab.Smoke(context.Background(), cfg); err != nil {
			t.Fatal(err)
		}
		assertUnchanged(t, cfg, confirmed, saved)
	})
}

const configSession = "source /opt/vyatta/etc/functions/script-template\nconfigure\ntrap 'cli-shell-api teardownSession' EXIT\n"

func labCommand(cfg testlab.Config, command, input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	out, err := testlab.Command(ctx, cfg, command, strings.NewReader(input))
	return string(out), err
}

func commandOK(t *testing.T, cfg testlab.Config, command, input string) string {
	t.Helper()
	out, err := labCommand(cfg, command, input)
	if err != nil {
		t.Fatalf("lab operation failed: %v", err)
	}
	return out
}

func deployFile(t *testing.T, cfg testlab.Config, data string) (string, error) {
	t.Helper()
	path := "/tmp/ubq-deploy-" + fmt.Sprint(time.Now().UnixNano()) + ".boot"
	return deployFileAt(t, cfg, data, path)
}

func deployFileAt(t *testing.T, cfg testlab.Config, data, path string) (string, error) {
	t.Helper()
	out, err := labCommand(cfg, "umask 077; cat > "+path, data)
	if err != nil {
		return "UBQ_STAGE:upload\n" + out, err
	}
	defer func() { _, _ = labCommand(cfg, "rm -f "+path, "") }()
	script, err := os.ReadFile(filepath.Join("lab", "deploy.vbash"))
	if err != nil {
		t.Fatal(err)
	}
	return labCommand(cfg, "/bin/vbash -s -- "+path, string(script))
}

func deployStage(output string) string {
	stage := "session"
	for line := range strings.SplitSeq(output, "\n") {
		if strings.HasPrefix(line, "UBQ_STAGE:") {
			stage = strings.TrimPrefix(line, "UBQ_STAGE:")
		}
	}
	return stage
}

func confirmAndSave(t *testing.T, cfg testlab.Config) {
	t.Helper()
	commandOK(t, cfg, "/bin/vbash -s", configSession+"confirm || builtin exit 1\nsave /config/config.boot || builtin exit 1\n")
}

func canonical(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	slices.Sort(lines)
	return strings.Join(lines, "\n")
}

func activeConfig(t *testing.T, cfg testlab.Config) string {
	t.Helper()
	return canonical(commandOK(t, cfg, testlab.ShowConfiguration, ""))
}

func savedHash(t *testing.T, cfg testlab.Config) string {
	t.Helper()
	return strings.TrimSpace(commandOK(t, cfg, "sha256sum /config/config.boot", ""))
}

func requireSetting(t *testing.T, active, expected string) {
	t.Helper()
	if !slices.Contains(strings.Split(active, "\n"), expected) {
		t.Fatalf("active configuration is missing %s", expected)
	}
}

func assertUnchanged(t *testing.T, cfg testlab.Config, active, saved string) {
	t.Helper()
	if activeConfig(t, cfg) != active {
		t.Fatal("active configuration differs from the complete confirmed snapshot")
	}
	if savedHash(t, cfg) != saved {
		t.Fatal("saved boot configuration changed unexpectedly")
	}
}

func requireRemoteFailure(t *testing.T, err error) {
	t.Helper()
	var exit *ssh.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("expected a native remote rejection, got %v", err)
	}
}

func requireTimer(t *testing.T, cfg testlab.Config, pending bool) {
	t.Helper()
	out, err := labCommand(cfg, "systemctl is-active commit-confirm.timer", "")
	if pending {
		if err != nil || strings.TrimSpace(out) != "active" {
			t.Fatal("native rollback timer is not active")
		}
	} else if strings.TrimSpace(out) != "inactive" {
		t.Fatal("native rollback timer remains active or cannot be inspected")
	}
}

func waitForRollback(t *testing.T, cfg testlab.Config, expected string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		out, err := labCommand(cfg, testlab.ShowConfiguration, "")
		timer, _ := labCommand(cfg, "systemctl is-active commit-confirm.timer", "")
		if err == nil && canonical(out) == expected && strings.TrimSpace(timer) == "inactive" {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatal("native rollback did not finish with the complete confirmed configuration within 2 minutes")
}
