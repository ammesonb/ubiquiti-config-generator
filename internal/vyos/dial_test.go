package vyos

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/deployment"
	"github.com/ammesonb/ubiquiti-config-generator/internal/sshclient"
	"golang.org/x/crypto/ssh"
)

func TestDialCancellationDuringPreparation(t *testing.T) {
	for _, stall := range []string{"start", "readiness"} {
		for _, trigger := range []string{"cancel", "timeout"} {
			t.Run(stall+"/"+trigger, func(t *testing.T) {
				cfg, started, disconnected := preparationServer(t, stall)
				if trigger == "timeout" {
					cfg.CommandTimeout = 200 * time.Millisecond
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				done := make(chan error, 1)
				go func() {
					s, err := (Dialer{SSH: cfg}).Dial(ctx)
					if s != nil {
						_ = s.Close(context.Background())
					}
					done <- err
				}()
				awaitSignal(t, started)
				want := context.DeadlineExceeded
				if trigger == "cancel" {
					want = context.Canceled
					cancel()
				}
				select {
				case err := <-done:
					if !errors.Is(err, want) {
						t.Fatalf("preparation lost its context error: %v", err)
					}
					var failure *deployment.StageError
					if !errors.As(err, &failure) {
						t.Fatalf("preparation lost its failing stage: %v", err)
					}
					if trigger == "timeout" {
						stage := deployment.StagePrepare
						if stall == "readiness" {
							stage = deployment.StageAwaitReady
						}
						if failure.Stage != stage {
							t.Fatalf("expected %s failure, got %s", stage, failure.Stage)
						}
					}
				case <-time.After(3 * time.Second):
					t.Fatal("preparation did not stop")
				}
				awaitSignal(t, disconnected)
			})
		}
	}
}

func TestDialContextDoesNotOwnPreparedSession(t *testing.T) {
	cfg, _, disconnected := preparationServer(t, "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, err := (Dialer{SSH: cfg}).Dial(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close(context.Background()) })
	cancel()
	if _, err := s.Upload(context.Background(), deployment.Artifact{Content: []byte("configuration")}); err != nil {
		t.Fatalf("completed Dial context interrupted the deployment session: %v", err)
	}
	if err := s.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	awaitSignal(t, disconnected)
}

func awaitSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(3 * time.Second):
		t.Fatal("SSH server did not reach the expected state")
	}
}

// Simulate startup and readiness separately to test cancellation without running VyOS commands.
func preparationServer(t *testing.T, stall string) (sshclient.Config, <-chan struct{}, <-chan struct{}) {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	serverConfig := &ssh.ServerConfig{NoClientAuth: true}
	serverConfig.AddHostKey(signer)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	accepted := make(chan net.Conn, 1)
	started, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		accepted <- conn
		defer func() { _ = conn.Close() }()
		connection, channels, requests, err := ssh.NewServerConn(conn, serverConfig)
		if err != nil {
			return
		}
		defer func() { _ = connection.Close() }()
		go ssh.DiscardRequests(requests)
		disconnected := make(chan struct{})
		var workers sync.WaitGroup
		first := true
		for incoming := range channels {
			channel, requests, err := incoming.Accept()
			if err != nil {
				break
			}
			startup := first
			first = false
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer func() { _ = channel.Close() }()
				for request := range requests {
					if request.Type != "exec" {
						_ = request.Reply(false, nil)
						continue
					}
					if startup {
						close(started)
						if stall == "start" {
							<-disconnected
							return
						}
					}
					_ = request.Reply(true, nil)
					if startup && stall == "" {
						_, _ = io.WriteString(channel, "UBQ_READY:123\n")
					}
					_, _ = io.Copy(io.Discard, channel)
					_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
					return
				}
			}()
		}
		close(disconnected)
		workers.Wait()
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case conn := <-accepted:
			_ = conn.Close()
		default:
		}
		awaitSignal(t, done)
	})
	return sshclient.Config{
		Address: listener.Addr().String(), User: "test",
		Auth: []ssh.AuthMethod{ssh.Password("unused")}, HostKey: ssh.FixedHostKey(signer.PublicKey()),
		ConnectTimeout: 2 * time.Second, CommandTimeout: 2 * time.Second,
	}, started, done
}
