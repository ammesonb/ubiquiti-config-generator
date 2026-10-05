//go:build integration

package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/testlab"
	"golang.org/x/crypto/ssh"
)

func docker(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("docker %s failed: %w", args[0], err)
	}
	return output, nil
}

func startDockerLab(t *testing.T) testlab.Config {
	t.Helper()
	get, err := testlab.Environment(filepath.Join("..", ".env.integration"))
	if err != nil {
		t.Fatal(err)
	}
	image := get("UBQ_VYOS_IMAGE")
	imageHash, err := hex.DecodeString(strings.TrimPrefix(image, "sha256:"))
	if err != nil || !strings.HasPrefix(image, "sha256:") || len(imageHash) != 32 {
		t.Fatal("UBQ_VYOS_IMAGE must be the immutable sha256 image ID of an imported VyOS OCI image; see docs/testing.md")
	}
	run := func(args ...string) string {
		t.Helper()
		output, err := docker(args...)
		if err != nil {
			t.Fatalf("%v: %s", err, output)
		}
		return strings.TrimSpace(string(output))
	}
	if serverOS := run("info", "--format", "{{.OSType}}"); serverOS != "linux" {
		t.Fatal("the VyOS lab requires Docker using Linux containers")
	}
	run("image", "inspect", image)
	name := "ubq-smoke-" + rand.Text()[:12]
	t.Cleanup(func() {
		// Cleanup uses independent timeouts even if the smoke test timed out.
		if t.Failed() {
			output, err := docker("logs", "--tail", "100", name)
			if err == nil {
				t.Logf("router startup logs:\n%s", output)
			}
		}
		if output, err := docker("rm", "--force", name); err != nil && !strings.Contains(string(output), "No such container") {
			t.Errorf("lab container cleanup: %v: %s", err, output)
		}
		if output, err := docker("network", "rm", name); err != nil && !strings.Contains(string(output), "not found") {
			t.Errorf("lab network cleanup: %v: %s", err, output)
		}
	})
	run("network", "create", "--ipv6", "--label", "ubq.test=smoke", name)
	run("create", "--name", name, "--label", "ubq.test=smoke", "--network", name,
		"--privileged", "--publish", "127.0.0.1::22", "--volume", "/lib/modules:/lib/modules:ro",
		"--tmpfs", "/run", "--tmpfs", "/run/lock", image, "/sbin/init")

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(filepath.Join("lab", "config.boot.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	keyData := strings.Fields(string(ssh.MarshalAuthorizedKey(publicKey)))[1]
	config := strings.ReplaceAll(string(fixture), "{{SSH_KEY}}", keyData)
	configPath := filepath.Join(t.TempDir(), "config.boot")
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	// docker cp also works with stopped containers and does not require bind mounts.
	run("cp", configPath, name+":/opt/vyatta/etc/config/config.boot")
	run("start", name)
	address := run("port", name, "22/tcp")
	if !strings.HasPrefix(address, "127.0.0.1:") {
		t.Fatalf("unexpected lab SSH binding: %s", address)
	}
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		// Trust this disposable host's key via Docker, never an unauthenticated SSH scan.
		keyBytes, err := docker("exec", name, "cat", "/etc/ssh/ssh_host_ed25519_key.pub")
		if err == nil {
			hostKey, _, _, _, parseErr := ssh.ParseAuthorizedKey(keyBytes)
			if parseErr != nil {
				t.Fatal("invalid SSH host public key in lab container")
			}
			return testlab.Config{
				Address: address, User: "ubqtest", Auth: ssh.PublicKeys(signer),
				HostKey: ssh.FixedHostKey(hostKey), HostKeyAlgorithms: []string{hostKey.Type()}, ExpectedHost: "ubq-test-router", Timeout: 2 * time.Minute,
			}
		}
		if running := run("inspect", "--format", "{{.State.Running}}", name); running != "true" {
			t.Fatal("VyOS exited during startup; see container logs")
		}
		time.Sleep(time.Second)
	}
	t.Fatal("VyOS did not generate an SSH host key within 2 minutes")
	return testlab.Config{}
}
