package testlab

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const ShowConfiguration = "/opt/vyatta/bin/vyatta-op-cmd-wrapper show configuration commands"

// Smoke waits for SSH and a readable active configuration, then verifies hostname.
// It issues one read-only operational command and never loads or commits settings.
func Smoke(ctx context.Context, cfg Config) error {
	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	var last error
	for {
		attempt, stop := context.WithTimeout(ctx, 5*time.Second)
		output, err := readConfiguration(attempt, cfg)
		stop()
		if err == nil {
			err = checkHostname(output, cfg.ExpectedHost)
		}
		if err == nil {
			return nil
		}
		var keyErr *knownHostError
		if errors.As(err, &keyErr) {
			return err
		}
		last = err
		select {
		case <-ctx.Done():
			return fmt.Errorf("router smoke test did not become ready: %w (last check: %v)", ctx.Err(), last)
		case <-time.After(500 * time.Millisecond):
		}
	}
}

type knownHostError struct{ err error }

func (e *knownHostError) Error() string { return "SSH host key verification failed" }
func (e *knownHostError) Unwrap() error { return e.err }

func readConfiguration(ctx context.Context, cfg Config) (string, error) {
	output, err := Command(ctx, cfg, ShowConfiguration, nil)
	if err != nil {
		return "", fmt.Errorf("read active configuration: %w", err)
	}
	return string(output), nil
}

// Command runs a lab command with verified SSH identity and a bounded context.
// Output is returned to the caller but never included in errors.
func Command(ctx context.Context, cfg Config, command string, input io.Reader) ([]byte, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("connect to lab: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("set SSH deadline: %w", err)
		}
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	var hostKeyErr error
	clientConfig := &ssh.ClientConfig{
		User:              cfg.User,
		Auth:              []ssh.AuthMethod{cfg.Auth},
		HostKeyAlgorithms: cfg.HostKeyAlgorithms,
		HostKeyCallback: func(host string, remote net.Addr, key ssh.PublicKey) error {
			hostKeyErr = cfg.HostKey(host, remote, key)
			return hostKeyErr
		},
	}
	sshConn, channels, requests, err := ssh.NewClientConn(conn, cfg.Address, clientConfig)
	if err != nil {
		if hostKeyErr != nil {
			return nil, &knownHostError{hostKeyErr}
		}
		return nil, fmt.Errorf("SSH handshake: %w", err)
	}
	client := ssh.NewClient(sshConn, channels, requests)
	defer func() { _ = client.Close() }()
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("open SSH session: %w", err)
	}
	defer func() { _ = session.Close() }()
	session.Stdin = input
	output, err := session.CombinedOutput(command)
	if err != nil {
		// Configuration and command output may contain secrets; do not log them.
		return output, fmt.Errorf("execute lab command: %w", err)
	}
	return output, nil
}

func checkHostname(output, expected string) error {
	for line := range strings.SplitSeq(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == "set" && fields[1] == "system" && fields[2] == "host-name" {
			if strings.Trim(fields[3], "\"'") == expected {
				return nil
			}
			return errors.New("active hostname does not match UBQ_TEST_EXPECT_HOSTNAME")
		}
	}
	return errors.New("active configuration does not contain a hostname setting")
}
