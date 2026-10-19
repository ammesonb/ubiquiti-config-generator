package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployOptionValidation(t *testing.T) {
	base := []string{"config.boot", "--host", "router:22", "--user", "vyos", "--identity", "key", "--known-hosts", "hosts"}
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"valid", nil, ""},
		{"IPv6", []string{"--host", "[::1]:22"}, ""},
		{"missing host", []string{"--host", ""}, "--host is required"},
		{"missing port", []string{"--host", "router"}, "--host must use host:port"},
		{"missing hostname", []string{"--host", ":22"}, "--host must include a hostname"},
		{"invalid port", []string{"--host", "router:ssh"}, "--host port must be an integer"},
		{"zero port", []string{"--host", "router:0"}, "--host port must be an integer"},
		{"large port", []string{"--host", "router:65536"}, "--host port must be an integer"},
		{"hostname whitespace", []string{"--host", "bad host:22"}, "must not contain whitespace"},
		{"blank user", []string{"--user", " "}, "--user is required"},
		{"zero confirmation", []string{"--confirm-timeout", "0"}, "--confirm-timeout must be greater than zero"},
		{"fractional minutes", []string{"--confirm-timeout", "90s"}, "--confirm-timeout must be a whole number of minutes"},
		{"negative connection", []string{"--connect-timeout", "-1s"}, "--connect-timeout must be greater than zero"},
		{"zero command", []string{"--command-timeout", "0"}, "--command-timeout must be greater than zero"},
		{"malformed duration", []string{"--command-timeout", "later"}, "invalid value"},
		{"supported algorithm", []string{"--host-key-algorithms", "ssh-ed25519"}, ""},
		{"unsupported algorithm", []string{"--host-key-algorithms", "unknown"}, `--host-key-algorithms: "unknown" is unsupported`},
		{"empty algorithm entry", []string{"--host-key-algorithms", "ssh-ed25519,"}, `--host-key-algorithms: "" is unsupported`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append(append([]string{}, base...), tc.args...)
			_, err := parseDeployOptions(args, io.Discard)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func TestDeployReportsIndependentProblemsTogether(t *testing.T) {
	_, err := parseDeployOptions([]string{"--connect-timeout", "0", "--confirm-timeout", "90s"}, io.Discard)
	if err == nil {
		t.Fatal("accepted invalid options")
	}
	for _, want := range []string{"configuration file is required", "--host is required", "--user is required", "--identity is required", "--known-hosts is required", "--connect-timeout must be greater than zero", "--confirm-timeout must be a whole number of minutes"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q from %v", want, err)
		}
	}
}

func TestDeploymentFileErrors(t *testing.T) {
	directory := t.TempDir()
	file := filepath.Join(directory, "config.boot")
	key := filepath.Join(directory, "key")
	options := deployOptions{File: file, Identity: key}
	_, _, err := loadDeployment(options)
	if err == nil || !strings.Contains(err.Error(), "configuration file:") || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing configuration: %v", err)
	}
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err = loadDeployment(options)
	if err == nil || err.Error() != "configuration file is empty" {
		t.Fatalf("empty configuration: %v", err)
	}
	if err := os.WriteFile(file, []byte("configuration"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err = loadDeployment(options)
	if err == nil || !strings.Contains(err.Error(), "--identity:") {
		t.Fatalf("missing identity: %v", err)
	}
	if err := os.WriteFile(key, []byte("sensitive invalid key data"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err = loadDeployment(options)
	if err == nil || err.Error() != "--identity does not contain a supported SSH private key" {
		t.Fatalf("invalid identity: %v", err)
	}
}
