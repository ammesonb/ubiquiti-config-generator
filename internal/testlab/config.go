// Package testlab supports tests against dedicated router instances.
// It is not the application's deployment implementation.
package testlab

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Environment reads optional dotenv settings without changing the process environment.
// Process variables take precedence, including explicitly empty values.
func Environment(path string) (func(string) string, error) {
	values, err := godotenv.Read(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// Do not include parser errors, which may contain credential values.
		return nil, fmt.Errorf("cannot read dotenv file %s; check permissions and syntax", path)
	}
	return func(key string) string {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return values[key]
	}, nil
}

// Config identifies a lab endpoint and the expected active configuration.
type Config struct {
	Address           string
	User              string
	Auth              ssh.AuthMethod
	HostKey           ssh.HostKeyCallback
	HostKeyAlgorithms []string
	ExpectedHost      string
	Timeout           time.Duration
}

// Load resolves relative credential paths against the dotenv file directory.
func Load(path string) (Config, error) {
	get, err := Environment(path)
	if err != nil {
		return Config{}, err
	}
	return fromEnvironment(get, filepath.Dir(path))
}

// endpointSettings keeps environment names, required fields, and defaults together.
// Port and timeout are parsed separately so validation errors never echo input values.
type endpointSettings struct {
	Host          string `env:"UBQ_TEST_HOST,notEmpty"`
	Port          string `env:"UBQ_TEST_PORT" envDefault:"22"`
	User          string `env:"UBQ_TEST_USER,notEmpty"`
	KnownHosts    string `env:"UBQ_TEST_KNOWN_HOSTS,notEmpty"`
	ExpectedHost  string `env:"UBQ_TEST_EXPECT_HOSTNAME,notEmpty"`
	Timeout       string `env:"UBQ_TEST_TIMEOUT" envDefault:"2m"`
	Password      string `env:"UBQ_TEST_PASSWORD"`
	SSHKey        string `env:"UBQ_TEST_SSH_KEY"`
	KeyPassphrase string `env:"UBQ_TEST_KEY_PASSPHRASE"`
}

func fromEnvironment(get func(string) string, base string) (Config, error) {
	var settings endpointSettings
	fields, err := env.GetFieldParams(&settings)
	if err != nil {
		return Config{}, fmt.Errorf("invalid lab environment tags: %w", err)
	}
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		values[field.Key] = get(field.Key)
	}
	var issues []error
	if err := env.ParseWithOptions(&settings, env.Options{Environment: values}); err != nil {
		var aggregate env.AggregateError
		if errors.As(err, &aggregate) {
			issues = append(issues, aggregate.Errors...)
		} else {
			issues = append(issues, err)
		}
	}
	cfg := Config{
		Address: net.JoinHostPort(settings.Host, settings.Port),
		User:    settings.User, ExpectedHost: settings.ExpectedHost,
	}
	port, err := strconv.Atoi(settings.Port)
	if err != nil || port < 1 || port > 65535 {
		issues = append(issues, errors.New("UBQ_TEST_PORT must be between 1 and 65535"))
	}
	cfg.Timeout, err = time.ParseDuration(settings.Timeout)
	if err != nil || cfg.Timeout <= 0 || cfg.Timeout > 3*time.Minute {
		issues = append(issues, errors.New("UBQ_TEST_TIMEOUT must be a positive duration no greater than 3m"))
	}
	resolve := func(path string) string {
		if filepath.IsAbs(path) {
			return path
		}
		return filepath.Join(base, path)
	}
	// Missing paths already have a required-field error; avoid cascading file errors.
	if settings.KnownHosts != "" {
		cfg.HostKey, err = knownhosts.New(resolve(settings.KnownHosts))
		if err != nil {
			issues = append(issues, errors.New("cannot load UBQ_TEST_KNOWN_HOSTS"))
		}
	}
	if (settings.Password == "") == (settings.SSHKey == "") {
		issues = append(issues, errors.New("set exactly one of UBQ_TEST_PASSWORD or UBQ_TEST_SSH_KEY"))
	} else if settings.Password != "" {
		cfg.Auth = ssh.Password(settings.Password)
	} else {
		cfg.Auth, err = keyAuthentication(resolve(settings.SSHKey), settings.KeyPassphrase)
		if err != nil {
			issues = append(issues, err)
		}
	}
	if err := errors.Join(issues...); err != nil {
		// Never return a partially usable endpoint when any validation failed.
		return Config{}, fmt.Errorf("invalid lab configuration:\n%w", err)
	}
	return cfg, nil
}

func keyAuthentication(path, passphrase string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("cannot read UBQ_TEST_SSH_KEY")
	}
	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(key)
	}
	if err != nil {
		return nil, errors.New("cannot parse UBQ_TEST_SSH_KEY; check format and passphrase")
	}
	return ssh.PublicKeys(signer), nil
}
