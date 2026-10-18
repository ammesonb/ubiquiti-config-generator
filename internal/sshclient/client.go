package sshclient

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config selects the device and the trust policy used to connect to it.
type Config struct {
	Address           string
	User              string
	Auth              []ssh.AuthMethod
	HostKey           ssh.HostKeyCallback
	HostKeyAlgorithms []string
	ConnectTimeout    time.Duration
	CommandTimeout    time.Duration
}

// Client is the shared SSH connection used for commands and persistent processes.
type Client struct {
	client         *ssh.Client
	commandTimeout time.Duration
	close          sync.Once
	closeErr       error
}

// Result retains command output for inspection even when execution fails.
type Result struct {
	Stdout []byte
	Stderr []byte
}

// Dial establishes an authenticated connection that can be reused across commands.
func Dial(ctx context.Context, cfg Config) (*Client, error) {
	if _, _, err := net.SplitHostPort(cfg.Address); err != nil {
		return nil, errors.New("SSH address must contain a host and port")
	}
	if cfg.User == "" || len(cfg.Auth) == 0 || cfg.HostKey == nil {
		return nil, errors.New("SSH user, authentication, and host key verification are required")
	}
	if cfg.ConnectTimeout < 0 || cfg.CommandTimeout < 0 {
		return nil, errors.New("SSH timeouts must not be negative")
	}
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 30 * time.Second
	}
	if cfg.CommandTimeout == 0 {
		cfg.CommandTimeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", cfg.Address)
	if err != nil {
		return nil, operationError{"connect", err}
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	sshConn, channels, requests, err := ssh.NewClientConn(conn, cfg.Address, &ssh.ClientConfig{
		User:              cfg.User,
		Auth:              cfg.Auth,
		HostKeyCallback:   cfg.HostKey,
		HostKeyAlgorithms: cfg.HostKeyAlgorithms,
	})
	// Cancellation can race with a successful handshake, leaving the connection already closing.
	stopped := stop()
	if err != nil || !stopped || ctx.Err() != nil {
		_ = conn.Close()
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, operationError{"handshake", err}
	}
	return &Client{client: ssh.NewClient(sshConn, channels, requests), commandTimeout: cfg.CommandTimeout}, nil
}

// Run executes an independent command and collects its output, including output produced before failure.
func (c *Client) Run(ctx context.Context, command string, input []byte) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, c.commandTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, commandError(ctx, err)
	}
	session, err := c.NewSession(ctx)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = session.Close() }()
	stop := context.AfterFunc(ctx, func() { _ = session.Close() })
	defer stop()
	var stdout, stderr bytes.Buffer
	session.Stdin = bytes.NewReader(input)
	session.Stdout, session.Stderr = &stdout, &stderr
	err = session.Run(command)
	result := Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if err != nil || ctx.Err() != nil {
		return result, commandError(ctx, err)
	}
	return result, nil
}

// NewSession opens a command channel for interacting with a process as it runs.
func (c *Client) NewSession(ctx context.Context) (*ssh.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, c.commandTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, operationError{"open session channel", err}
	}
	// The SSH library cannot cancel an individual channel-open request, so interrupt it by closing the connection.
	stop := context.AfterFunc(ctx, func() { _ = c.Close() })
	defer stop()
	session, err := c.client.NewSession()
	if err != nil || ctx.Err() != nil {
		if session != nil {
			_ = session.Close()
		}
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, operationError{"open session channel", err}
	}
	return session, nil
}

// Close disconnects all sessions sharing this client.
func (c *Client) Close() error {
	c.close.Do(func() { c.closeErr = c.client.Close() })
	if c.closeErr != nil {
		return operationError{"close", c.closeErr}
	}
	return nil
}

func commandError(ctx context.Context, err error) error {
	// Preserve the cancellation that caused the channel to close, rather than the resulting transport error.
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return operationError{"command", err}
}

type operationError struct {
	operation string
	err       error
}

func (e operationError) Error() string {
	// Omit underlying transport and authentication diagnostics because they may contain credentials.
	message := "SSH " + e.operation + " failed"
	if errors.Is(e.err, context.DeadlineExceeded) {
		return message + ": context deadline exceeded"
	}
	if errors.Is(e.err, context.Canceled) {
		return message + ": context canceled"
	}
	return message
}
func (e operationError) Unwrap() error { return e.err }
