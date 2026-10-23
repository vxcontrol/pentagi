package system

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"
)

// blackholeAddr returns a local address that never completes a TCP handshake: a
// listener with a zero backlog that never accepts holds one connection in its
// queue and Linux silently drops every SYN after it, as an unreachable host does.
func blackholeAddr(t *testing.T) string {
	t.Helper()

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("failed to create socket: %v", err)
	}
	t.Cleanup(func() { syscall.Close(fd) })

	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}); err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	if err := syscall.Listen(fd, 0); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	sa, err := syscall.Getsockname(fd)
	if err != nil {
		t.Fatalf("failed to read the bound address: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", sa.(*syscall.SockaddrInet4).Port)

	filler, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("failed to fill the accept queue: %v", err)
	}
	t.Cleanup(func() { filler.Close() })

	return addr
}

func TestGetHTTPClient_BlackholedHostFailsWithinTheDialTimeout(t *testing.T) {
	shortenConnectTimeouts(t, 300*time.Millisecond, time.Second)

	tests := []struct {
		name   string
		target func(addr string) (url, proxy string)
	}{
		{
			name:   "unreachable endpoint",
			target: func(addr string) (string, string) { return "https://" + addr + "/v1/chat/completions", "" },
		},
		{
			name: "unreachable proxy",
			target: func(addr string) (string, string) {
				return "https://api.openai.com/v1/chat/completions", "http://" + addr
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, proxyURL := tt.target(blackholeAddr(t))
			client, err := GetHTTPClient(createTestConfig("", false, proxyURL))
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// The guard only keeps a missing bound from hanging the test for minutes.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}

			start := time.Now()
			resp, err := client.Do(req)
			elapsed := time.Since(start)
			if err == nil {
				resp.Body.Close()
				t.Fatal("expected the dial to fail")
			}

			var netErr net.Error
			if ctx.Err() != nil || !errors.As(err, &netErr) || !netErr.Timeout() {
				t.Errorf("expected the dial itself to time out before the guard, got: %v", err)
			}
			if elapsed > 2*time.Second {
				t.Errorf("expected to give up within the dial timeout, took %v", elapsed)
			}
		})
	}
}

func TestGetHTTPClient_CancelDuringDialReturnsAtOnce(t *testing.T) {
	client, err := GetHTTPClient(createTestConfig("", false, ""))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+blackholeAddr(t), nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected the request to be canceled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected the cancel to stop the dial, took %v", elapsed)
	}
}
