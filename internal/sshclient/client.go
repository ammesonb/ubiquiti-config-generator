// Package sshclient provides bounded SSH operations for device implementations.
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

// Config describes one endpoint. HostKey must verify a previously trusted key.
// Authentication and host-key policy are supplied by the caller.
type Config struct {
	Address           string
	User              string
	Auth              []ssh.AuthMethod
	HostKey           ssh.HostKeyCallback
	HostKeyAlgorithms []string
	// ConnectTimeout bounds TCP connection setup and the SSH handshake together.
	// Zero uses 30s.
	ConnectTimeout time.Duration
	// CommandTimeout bounds each Run call, including opening its session channel.
	// Zero uses 30s.
	CommandTimeout time.Duration
}

// Client reuses one authenticated SSH connection across Run calls.
// Run calls must be sequential.
type Client struct {
	client         *ssh.Client
	commandTimeout time.Duration
	close          sync.Once
	closeErr       error
}

// Result contains remote output, including output preceding a failed command.
// Either stream can contain credentials or leak configuration, so ensure it is not logged.
type Result struct {
	Stdout []byte
	Stderr []byte
}

// Dial verifies the endpoint before returning an authenticated client.
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
	// Closing the socket interrupts handshake reads even without a ctx deadline.
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	sshConn, channels, requests, err := ssh.NewClientConn(conn, cfg.Address, &ssh.ClientConfig{
		User:              cfg.User,
		Auth:              cfg.Auth,
		HostKeyCallback:   cfg.HostKey,
		HostKeyAlgorithms: cfg.HostKeyAlgorithms,
	})
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

// Run executes command in a new SSH session, passing input unchanged to stdin.
// Command is sent unchanged to the remote shell, so callers must quote interpolated arguments.
// Cancellation or timeout during Run closes the client connection, but may not
// interrupt remote command execution.
func (c *Client) Run(ctx context.Context, command string, input []byte) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, c.commandTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return Result{}, commandError(ctx, err)
	}
	stop := context.AfterFunc(ctx, func() { _ = c.Close() })
	defer stop()
	session, err := c.client.NewSession()
	if err != nil {
		return Result{}, commandError(ctx, err)
	}
	defer func() { _ = session.Close() }()
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

// Close releases the connection and is safe to call repeatedly or during Run.
func (c *Client) Close() error {
	c.close.Do(func() { c.closeErr = c.client.Close() })
	if c.closeErr != nil {
		return operationError{"close", c.closeErr}
	}
	return nil
}

func commandError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return operationError{"command", err}
}

// Keep commands, output, and callback diagnostics out of printable errors while
// preserving native exit status, host identity errors, and context errors.
type operationError struct {
	operation string
	err       error
}

func (e operationError) Error() string { return "SSH " + e.operation + " failed" }
func (e operationError) Unwrap() error { return e.err }
