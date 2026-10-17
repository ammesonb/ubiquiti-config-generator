package sshclient_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
	"golang.org/x/crypto/ssh"
)

func signer(t *testing.T) ssh.Signer {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

// This server handles synthetic exec requests; it never invokes a local shell.
func server(t *testing.T, run func(ssh.Channel, string)) (sshclient.Config, *atomic.Int32) {
	t.Helper()
	key := signer(t)
	cfg := &ssh.ServerConfig{PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		if meta.User() != "router" || string(password) != "test-only" {
			return nil, errors.New("authentication rejected")
		}
		return nil, nil
	}}
	cfg.AddHostKey(key)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	var connections sync.Map
	var count atomic.Int32
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			connections.Store(conn, true)
			count.Add(1)
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer connections.Delete(conn)
				defer func() { _ = conn.Close() }()
				if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
					return
				}
				connection, channels, requests, err := ssh.NewServerConn(conn, cfg)
				if err != nil {
					return
				}
				defer func() { _ = connection.Close() }()
				go ssh.DiscardRequests(requests)
				for incoming := range channels {
					if incoming.ChannelType() != "session" {
						_ = incoming.Reject(ssh.UnknownChannelType, "session required")
						continue
					}
					channel, reqs, err := incoming.Accept()
					if err != nil {
						return
					}
					for request := range reqs {
						var command struct{ Value string }
						if request.Type != "exec" || ssh.Unmarshal(request.Payload, &command) != nil {
							_ = request.Reply(false, nil)
							continue
						}
						_ = request.Reply(true, nil)
						run(channel, command.Value)
						break
					}
					_ = channel.Close()
				}
			}()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		connections.Range(func(key, _ any) bool {
			_ = key.(net.Conn).Close()
			return true
		})
		workers.Wait()
	})
	return sshclient.Config{
		Address: listener.Addr().String(), User: "router",
		Auth:           []ssh.AuthMethod{ssh.Password("test-only")},
		HostKey:        ssh.FixedHostKey(key.PublicKey()),
		ConnectTimeout: 2 * time.Second, CommandTimeout: 2 * time.Second,
	}, &count
}

func exit(channel ssh.Channel, code uint32) {
	_, _ = channel.SendRequest("exit-status", false, binary.BigEndian.AppendUint32(nil, code))
}

func TestRunReusesConnectionAndPreservesExitStatus(t *testing.T) {
	cfg, connections := server(t, func(channel ssh.Channel, command string) {
		input, _ := io.ReadAll(channel)
		if command == "reject" {
			_, _ = io.WriteString(channel.Stderr(), "sensitive-device-diagnostic")
			exit(channel, 7)
			return
		}
		_, _ = channel.Write(input)
		exit(channel, 0)
	})
	client, err := sshclient.Dial(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	for range 2 {
		result, err := client.Run(context.Background(), "echo-input", []byte("native configuration\n"))
		if err != nil || string(result.Stdout) != "native configuration\n" || len(result.Stderr) != 0 {
			t.Fatal("exec did not return the complete stdin payload on stdout")
		}
	}
	result, err := client.Run(context.Background(), "reject", nil)
	var status *ssh.ExitError
	if !errors.As(err, &status) || status.ExitStatus() != 7 {
		t.Fatalf("native exit status was lost: %v", err)
	}
	if string(result.Stderr) != "sensitive-device-diagnostic" || strings.Contains(err.Error(), "sensitive-device-diagnostic") {
		t.Fatal("diagnostics must be returned separately from printable errors")
	}
	if connections.Load() != 1 {
		t.Fatal("commands did not reuse the authenticated connection")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if err := client.Close(); err != nil {
		t.Fatal("repeated Close failed")
	}
}

func TestDialRejectsUntrustedHostAndBadAuthentication(t *testing.T) {
	for _, scenario := range []string{"host", "authentication"} {
		t.Run(scenario, func(t *testing.T) {
			cfg, _ := server(t, func(ssh.Channel, string) { t.Error("untrusted connection executed a command") })
			var want error
			if scenario == "host" {
				want = errors.New("sensitive-host-key-diagnostic")
				cfg.HostKey = func(string, net.Addr, ssh.PublicKey) error { return want }
			} else {
				cfg.Auth = []ssh.AuthMethod{ssh.Password("wrong-password")}
			}
			client, err := sshclient.Dial(context.Background(), cfg)
			if err == nil {
				_ = client.Close()
				t.Fatal("untrusted connection was accepted")
			}
			if want != nil && !errors.Is(err, want) {
				t.Fatal("host verification cause was lost")
			}
			if strings.Contains(err.Error(), "sensitive-host-key-diagnostic") || strings.Contains(err.Error(), "wrong-password") {
				t.Fatal("connection error exposed sensitive diagnostics")
			}
		})
	}
}

func TestDialRejectsMissingSettings(t *testing.T) {
	base := sshclient.Config{
		Address: "127.0.0.1:22", User: "router", Auth: []ssh.AuthMethod{ssh.Password("test-only")},
		HostKey: ssh.FixedHostKey(signer(t).PublicKey()),
	}
	for _, change := range []func(*sshclient.Config){
		func(c *sshclient.Config) { c.Address = "missing-port" },
		func(c *sshclient.Config) { c.User = "" },
		func(c *sshclient.Config) { c.Auth = nil },
		func(c *sshclient.Config) { c.HostKey = nil },
		func(c *sshclient.Config) { c.ConnectTimeout = -time.Second },
		func(c *sshclient.Config) { c.CommandTimeout = -time.Second },
	} {
		cfg := base
		change(&cfg)
		if client, err := sshclient.Dial(context.Background(), cfg); err == nil {
			_ = client.Close()
			t.Fatal("invalid settings were accepted")
		}
	}
}

func TestRunCancellationClosesConnection(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	cfg, _ := server(t, func(ssh.Channel, string) {
		close(started)
		<-release
	})
	t.Cleanup(func() { close(release) })
	client, err := sshclient.Dial(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := client.Run(ctx, "stall", nil); done <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("command did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation cause was lost: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not interrupt command execution")
	}
	if _, err := client.Run(context.Background(), "after-cancel", nil); err == nil {
		t.Fatal("canceled connection remained usable")
	}
}

func TestHandshakeTimeoutClosesSocket(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = io.Copy(io.Discard, conn)
	}()
	cfg := sshclient.Config{
		Address: listener.Addr().String(), User: "router", Auth: []ssh.AuthMethod{ssh.Password("test-only")},
		HostKey:        ssh.FixedHostKey(signer(t).PublicKey()),
		ConnectTimeout: 100 * time.Millisecond, CommandTimeout: 2 * time.Second,
	}
	client, err := sshclient.Dial(context.Background(), cfg)
	if err == nil {
		_ = client.Close()
		t.Fatal("stalled handshake succeeded")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("handshake timeout cause was lost: %v", err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handshake timeout did not close the socket")
	}
}

func TestCommandTimeoutClosesConnection(t *testing.T) {
	release := make(chan struct{})
	cfg, _ := server(t, func(ssh.Channel, string) { <-release })
	t.Cleanup(func() { close(release) })
	cfg.CommandTimeout = 100 * time.Millisecond
	client, err := sshclient.Dial(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	done := make(chan error, 1)
	go func() { _, err := client.Run(context.Background(), "stall", nil); done <- err }()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("command timeout cause was lost: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not honor its command timeout independently of the connection timeout")
	}
	if _, err := client.Run(context.Background(), "after-timeout", nil); err == nil {
		t.Fatal("timed-out connection remained usable")
	}
}

func TestExpiredContextDoesNotStartCommandOrCloseConnection(t *testing.T) {
	var commands atomic.Int32
	cfg, _ := server(t, func(channel ssh.Channel, _ string) {
		commands.Add(1)
		exit(channel, 0)
	})
	dialCtx, stopDial := context.WithCancel(context.Background())
	defer stopDial()
	client, err := sshclient.Dial(dialCtx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	// Dial's context governs connection setup, not the lifetime of the client.
	stopDial()
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer stop()
	for _, ctx := range []context.Context{canceled, expired} {
		if _, err := client.Run(ctx, "must-not-run", nil); !errors.Is(err, ctx.Err()) {
			t.Fatalf("Run did not preserve the already-expired context error: %v", err)
		}
	}
	if _, err := client.Run(context.Background(), "still-usable", nil); err != nil {
		t.Fatalf("early cancellation or completed Dial context closed the connection: %v", err)
	}
	if commands.Load() != 1 {
		t.Fatal("an already-expired context allowed a remote command to start")
	}
}
