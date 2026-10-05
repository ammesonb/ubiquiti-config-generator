package testlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvironmentPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.integration")
	if err := os.WriteFile(path, []byte("UBQ_TEST_USER='file user'\nUBQ_TEST_HOST=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UBQ_TEST_HOST", "from-process")
	get, err := Environment(path)
	if err != nil {
		t.Fatal(err)
	}
	if get("UBQ_TEST_USER") != "file user" || get("UBQ_TEST_HOST") != "from-process" {
		t.Fatal("dotenv values or environment precedence were not respected")
	}
	t.Setenv("UBQ_TEST_HOST", "")
	if get("UBQ_TEST_HOST") != "" {
		t.Fatal("an explicitly empty process value must override the file")
	}
}

func TestConfigurationRequiresExplicitEndpoint(t *testing.T) {
	_, err := fromEnvironment(func(string) string { return "" }, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "UBQ_TEST_HOST") {
		t.Fatalf("missing endpoint must fail, got %v", err)
	}
}

func TestConfigurationValidation(t *testing.T) {
	for _, tc := range []struct {
		name, field, value, want string
	}{
		{"port", "UBQ_TEST_PORT", "70000", "PORT"},
		{"malformed port", "UBQ_TEST_PORT", "not-a-number", "PORT"},
		{"malformed timeout", "UBQ_TEST_TIMEOUT", "not-a-duration", "TIMEOUT"},
		{"timeout", "UBQ_TEST_TIMEOUT", "0s", "TIMEOUT"},
		{"oversized timeout", "UBQ_TEST_TIMEOUT", "1h", "TIMEOUT"},
		{"missing auth", "UBQ_TEST_PASSWORD", "", "exactly one"},
		{"ambiguous auth", "UBQ_TEST_SSH_KEY", "a-key", "exactly one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			if err := os.WriteFile(filepath.Join(base, "known_hosts"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			values := map[string]string{
				"UBQ_TEST_HOST": "::1", "UBQ_TEST_USER": "lab", "UBQ_TEST_KNOWN_HOSTS": "known_hosts",
				"UBQ_TEST_EXPECT_HOSTNAME": "lab", "UBQ_TEST_PASSWORD": "test-only",
			}
			values[tc.field] = tc.value
			_, err := fromEnvironment(func(key string) string { return values[key] }, base)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("wanted %s error, got %v", tc.want, err)
			}
		})
	}
}

func TestConfigRelativePathsAndIPv6(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "known_hosts"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{
		"UBQ_TEST_HOST": "::1", "UBQ_TEST_USER": "lab", "UBQ_TEST_KNOWN_HOSTS": "known_hosts",
		"UBQ_TEST_EXPECT_HOSTNAME": "lab", "UBQ_TEST_PASSWORD": "test-only",
	}
	cfg, err := fromEnvironment(func(key string) string { return values[key] }, base)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "[::1]:22" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}
}

func TestMalformedDotenvDoesNotLeakCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.integration")
	if err := os.WriteFile(path, []byte("invalid secret-value =oops\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Environment(path)
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("expected a sanitized parse error, got %v", err)
	}
}

func TestConfigurationCollectsMissingSettings(t *testing.T) {
	cfg, err := fromEnvironment(func(string) string { return "" }, t.TempDir())
	if err == nil {
		t.Fatal("expected missing settings errors")
	}
	for _, field := range []string{"UBQ_TEST_HOST", "UBQ_TEST_USER", "UBQ_TEST_KNOWN_HOSTS", "UBQ_TEST_EXPECT_HOSTNAME", "exactly one"} {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("missing %s error in %v", field, err)
		}
	}
	if strings.Contains(err.Error(), "cannot load") || strings.Contains(err.Error(), "cannot read") {
		t.Fatalf("missing settings should not cause cascading file errors: %v", err)
	}
	if cfg.Address != "" || cfg.Auth != nil || cfg.HostKey != nil {
		t.Fatal("validation failure returned a partial endpoint")
	}
}

func TestConfigurationCollectsIndependentErrorsWithoutValues(t *testing.T) {
	base := t.TempDir()
	const secret = "private-key-and-passphrase-secret"
	if err := os.WriteFile(filepath.Join(base, "bad-key"), []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	values := map[string]string{
		"UBQ_TEST_HOST": "localhost", "UBQ_TEST_EXPECT_HOSTNAME": "lab",
		"UBQ_TEST_PORT": "70000", "UBQ_TEST_TIMEOUT": "1h",
		"UBQ_TEST_KNOWN_HOSTS": "missing-file", "UBQ_TEST_SSH_KEY": "bad-key",
		"UBQ_TEST_KEY_PASSPHRASE": secret,
	}
	_, err := fromEnvironment(func(key string) string { return values[key] }, base)
	if err == nil {
		t.Fatal("expected multiple errors")
	}
	for _, field := range []string{"UBQ_TEST_USER", "UBQ_TEST_PORT", "UBQ_TEST_TIMEOUT", "UBQ_TEST_KNOWN_HOSTS", "UBQ_TEST_SSH_KEY"} {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("missing %s error in %v", field, err)
		}
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "missing-file") {
		t.Fatalf("configuration values leaked through errors: %v", err)
	}
	values["UBQ_TEST_SSH_KEY"] = "unreadable-key"
	_, err = fromEnvironment(func(key string) string { return values[key] }, base)
	if err == nil || !strings.Contains(err.Error(), "cannot read UBQ_TEST_SSH_KEY") || !strings.Contains(err.Error(), "UBQ_TEST_TIMEOUT") {
		t.Fatalf("file-read failures must aggregate with other errors, got %v", err)
	}
}
