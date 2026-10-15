package testlab

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

// This server tests SSH transport behavior, not VyOS configuration semantics.
func testServer(t *testing.T, output string, status uint32, extraKeys ...ssh.Signer) (string, ssh.PublicKey) {
	t.Helper()
	signer := testSigner(t)
	serverConfig := &ssh.ServerConfig{PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		if meta.User() == "lab" && string(password) == "test-only" {
			return nil, nil
		}
		return nil, errors.New("authentication rejected")
	}}
	serverConfig.AddHostKey(signer)
	for _, key := range extraKeys {
		serverConfig.AddHostKey(key)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer func() { _ = conn.Close() }()
				if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
					return
				}
				server, channels, requests, err := ssh.NewServerConn(conn, serverConfig)
				if err != nil {
					return
				}
				defer func() { _ = server.Close() }()
				go ssh.DiscardRequests(requests)
				for incoming := range channels {
					channel, reqs, err := incoming.Accept()
					if err != nil {
						return
					}
					for request := range reqs {
						var command struct{ Value string }
						if request.Type != "exec" || ssh.Unmarshal(request.Payload, &command) != nil || command.Value != ShowConfiguration {
							_ = request.Reply(false, nil)
							continue
						}
						_ = request.Reply(true, nil)
						_, _ = io.WriteString(channel, output)
						_, _ = channel.SendRequest("exit-status", false, binary.BigEndian.AppendUint32(nil, status))
						_ = channel.Close()
						break
					}
				}
			}()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		workers.Wait()
	})
	return listener.Addr().String(), signer.PublicKey()
}

func TestSmokeSSH(t *testing.T) {
	for _, tc := range []struct {
		name, output, want string
		status             uint32
	}{
		{"ready", "set system host-name 'lab'\n", "", 0},
		{"wrong configuration", "set system host-name 'other'\n", "hostname does not match", 0},
		{"command failure", "sensitive-config-output", "read active configuration", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			address, key := testServer(t, tc.output, tc.status)
			cfg := Config{Address: address, User: "lab", Auth: ssh.Password("test-only"), HostKey: ssh.FixedHostKey(key), ExpectedHost: "lab", Timeout: 200 * time.Millisecond}
			err := Smoke(context.Background(), cfg)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
			if err != nil && strings.Contains(err.Error(), "sensitive-config-output") {
				t.Fatal("configuration output leaked through the error")
			}
		})
	}
}

func TestSmokeRejectsUnknownHostKey(t *testing.T) {
	address, _ := testServer(t, "set system host-name lab\n", 0)
	cfg := Config{Address: address, User: "lab", Auth: ssh.Password("test-only"), HostKey: ssh.FixedHostKey(testSigner(t).PublicKey()), ExpectedHost: "lab", Timeout: time.Second}
	err := Smoke(context.Background(), cfg)
	var hostErr *knownHostError
	if !errors.As(err, &hostErr) {
		t.Fatalf("expected immediate host identity failure, got %v", err)
	}
}

func TestHostnameParsing(t *testing.T) {
	for _, value := range []string{"lab", "'lab'", "\"lab\""} {
		if err := checkHostname("set system host-name "+value+"\r\n", "lab"); err != nil {
			t.Fatal(err)
		}
	}
	for _, output := range []string{"", "set system host-name lab-extra", "banner set system host-name lab", "set system description lab"} {
		if checkHostname(output, "lab") == nil {
			t.Fatalf("unexpected hostname match in %q", output)
		}
	}
}

func TestSmokeHandshakeDeadline(t *testing.T) {
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
	cfg := Config{Address: listener.Addr().String(), User: "lab", Auth: ssh.Password("test-only"), HostKey: ssh.FixedHostKey(testSigner(t).PublicKey()), ExpectedHost: "lab", Timeout: 100 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := Smoke(ctx, cfg); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("stalled SSH handshake must reach the deadline, got %v", err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed-out handshake did not close the TCP connection")
	}
}

func TestSmokeNegotiatesTrustedKeyType(t *testing.T) {
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	alternate, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	address, trusted := testServer(t, "set system host-name lab\n", 0, alternate)
	cfg := Config{
		Address: address, User: "lab", Auth: ssh.Password("test-only"),
		HostKey: ssh.FixedHostKey(trusted), ExpectedHost: "lab", Timeout: time.Second,
	}
	// With both keys offered, the default negotiation selects the untrusted ECDSA key.
	var hostErr *knownHostError
	if err := Smoke(context.Background(), cfg); !errors.As(err, &hostErr) {
		t.Fatalf("expected rejection of the other host key, got %v", err)
	}
	cfg.HostKeyAlgorithms = []string{trusted.Type()}
	if err := Smoke(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}
