package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testWaitTimeout only bounds how long a broken test can hang; a passing test
// never waits on it.
const testWaitTimeout = 10 * time.Second

var errSimulatedServerFailure = errors.New("accept tcp: too many open files")

// --- helpers ---------------------------------------------------------------

type shutdownRecorder struct {
	mu     sync.Mutex
	events []string
}

func (r *shutdownRecorder) add(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
}

func (r *shutdownRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.events...)
}

func requireEvents(t *testing.T, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
}

// ctxSnapshot captures a context as the callee saw it. The production code
// cancels its contexts on return, so they must be inspected during the call.
type ctxSnapshot struct {
	err         error
	deadline    time.Time
	hasDeadline bool
	taken       time.Time
}

func snapshotCtx(ctx context.Context) ctxSnapshot {
	deadline, ok := ctx.Deadline()

	return ctxSnapshot{err: ctx.Err(), deadline: deadline, hasDeadline: ok, taken: time.Now()}
}

// requireFresh asserts the context was live (not derived from the cancelled
// signal context) and bounded by a deadline no longer than limit.
func (c ctxSnapshot) requireFresh(t *testing.T, name string, limit time.Duration) {
	t.Helper()

	if c.err != nil {
		t.Errorf("%s context was already done: %v", name, c.err)
	}
	if !c.hasDeadline {
		t.Errorf("%s context has no deadline", name)
		return
	}
	if remaining := c.deadline.Sub(c.taken); remaining <= 0 || remaining > limit {
		t.Errorf("%s deadline is %v away, want within (0, %v]", name, remaining, limit)
	}
}

// stubHTTPServer lets a test decide how Shutdown and Close behave.
type stubHTTPServer struct {
	shutdownFn func(ctx context.Context) error
	closeFn    func() error
}

func (s *stubHTTPServer) Shutdown(ctx context.Context) error { return s.shutdownFn(ctx) }
func (s *stubHTTPServer) Close() error                       { return s.closeFn() }

// newGracefulStub returns a server whose serve func blocks until it is shut
// down or closed and then reports http.ErrServerClosed, like net/http does.
// Shutdown runs on the caller's goroutine, so *shutdownCtx is safe to read once
// the call under test has returned.
func newGracefulStub(emit func(string)) (srv *stubHTTPServer, serve func() error, shutdownCtx *ctxSnapshot) {
	stopped := make(chan struct{})
	var once sync.Once
	stop := func() { once.Do(func() { close(stopped) }) }

	shutdownCtx = &ctxSnapshot{}
	srv = &stubHTTPServer{
		shutdownFn: func(ctx context.Context) error {
			*shutdownCtx = snapshotCtx(ctx)
			emit("shutdown")
			stop()
			return nil
		},
		closeFn: func() error {
			emit("close")
			stop()
			return nil
		},
	}
	serve = func() error {
		<-stopped
		return http.ErrServerClosed
	}

	return srv, serve, shutdownCtx
}

// spyHTTPServer wraps a real *http.Server and observes how it is stopped.
type spyHTTPServer struct {
	*http.Server

	shutdownStarted chan struct{}
	startOnce       sync.Once
	closeCalls      atomic.Int32
}

func (s *spyHTTPServer) Shutdown(ctx context.Context) error {
	s.startOnce.Do(func() { close(s.shutdownStarted) })
	return s.Server.Shutdown(ctx)
}

func (s *spyHTTPServer) Close() error {
	s.closeCalls.Add(1)
	return s.Server.Close()
}

// blockedHandler holds every request until it is released.
type blockedHandler struct {
	entered     chan struct{}
	release     chan struct{}
	enterOnce   sync.Once
	releaseOnce sync.Once
}

func (h *blockedHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.enterOnce.Do(func() { close(h.entered) })
	<-h.release
	_, _ = io.WriteString(w, "done")
}

func (h *blockedHandler) unblock() { h.releaseOnce.Do(func() { close(h.release) }) }

// startBlockedServer prepares a real loopback server whose handler blocks.
// The server is not serving yet: the returned serve func starts it, so it can
// be handed to superviseHTTPServer exactly as production does.
func startBlockedServer(t *testing.T) (spy *spyHTTPServer, serve func() error, h *blockedHandler, url string) {
	t.Helper()

	h = &blockedHandler{entered: make(chan struct{}), release: make(chan struct{})}
	ts := httptest.NewUnstartedServer(h)

	// Cleanups run last-in first-out: release the handler, then close.
	t.Cleanup(ts.Close)
	t.Cleanup(h.unblock)

	spy = &spyHTTPServer{Server: ts.Config, shutdownStarted: make(chan struct{})}
	serve = func() error { return ts.Config.Serve(ts.Listener) }

	return spy, serve, h, "http://" + ts.Listener.Addr().String()
}

type requestResult struct {
	status int
	body   string
	err    error
}

// startRequest issues a GET on its own connection and reports the outcome.
func startRequest(url string) <-chan requestResult {
	result := make(chan requestResult, 1)

	go func() {
		client := &http.Client{
			Timeout:   testWaitTimeout,
			Transport: &http.Transport{DisableKeepAlives: true},
		}

		resp, err := client.Get(url)
		if err != nil {
			result <- requestResult{err: err}
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		result <- requestResult{status: resp.StatusCode, body: string(body), err: err}
	}()

	return result
}

func waitClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(testWaitTimeout):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func waitError(t *testing.T, ch <-chan error, what string) error {
	t.Helper()

	select {
	case err := <-ch:
		return err
	case <-time.After(testWaitTimeout):
		t.Fatalf("timed out waiting for %s", what)
		return nil
	}
}

func waitResult(t *testing.T, ch <-chan requestResult, what string) requestResult {
	t.Helper()

	select {
	case res := <-ch:
		return res
	case <-time.After(testWaitTimeout):
		t.Fatalf("timed out waiting for %s", what)
		return requestResult{}
	}
}

// --- graceful shutdown against a real http.Server ---------------------------

func TestSuperviseHTTPServer_ActiveRequestCompletes(t *testing.T) {
	spy, serve, handler, url := startBlockedServer(t)

	ctx, sendSignal := context.WithCancel(context.Background())
	defer sendSignal()

	done := make(chan error, 1)
	go func() { done <- superviseHTTPServer(ctx, spy, serve, httpShutdownTimeout) }()

	response := startRequest(url)
	waitClosed(t, handler.entered, "the request to reach the handler")

	sendSignal() // stands in for SIGTERM
	waitClosed(t, spy.shutdownStarted, "Shutdown to be called")

	// The handler is still running, so a graceful Shutdown cannot be done yet.
	select {
	case err := <-done:
		t.Fatalf("shutdown returned while a request was in flight: %v", err)
	default:
	}

	handler.unblock()

	if err := waitError(t, done, "graceful shutdown to finish"); err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}

	res := waitResult(t, response, "the in-flight response")
	if res.err != nil || res.status != http.StatusOK || res.body != "done" {
		t.Fatalf("in-flight request = (status %d, body %q, err %v), want it to complete normally",
			res.status, res.body, res.err)
	}

	if n := spy.closeCalls.Load(); n != 0 {
		t.Errorf("Close called %d times after a successful graceful shutdown, want 0", n)
	}
}

func TestSuperviseHTTPServer_DeadlineExpiryForcesClose(t *testing.T) {
	spy, serve, handler, url := startBlockedServer(t)

	ctx, sendSignal := context.WithCancel(context.Background())
	defer sendSignal()

	// The handler is never released before the deadline, so expiry is certain
	// whatever the machine's speed.
	done := make(chan error, 1)
	go func() { done <- superviseHTTPServer(ctx, spy, serve, 100*time.Millisecond) }()

	response := startRequest(url)
	waitClosed(t, handler.entered, "the request to reach the handler")

	sendSignal()

	err := waitError(t, done, "shutdown to give up")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want it to report context.DeadlineExceeded", err)
	}

	if n := spy.closeCalls.Load(); n != 1 {
		t.Errorf("Close called %d times, want exactly 1 forced close", n)
	}

	// The forced close must really cut the stuck connection.
	if res := waitResult(t, response, "the stuck request to be cut"); res.err == nil {
		t.Errorf("stuck request finished with status %d, want its connection closed", res.status)
	}
}

func TestShutdownHTTPServer_ForceCloseErrorIsReported(t *testing.T) {
	closeErr := errors.New("close failed")
	srv := &stubHTTPServer{
		shutdownFn: func(context.Context) error { return context.DeadlineExceeded },
		closeFn:    func() error { return closeErr },
	}

	err := shutdownHTTPServer(srv, httpShutdownTimeout)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want it to keep the graceful-shutdown failure", err)
	}
	if !errors.Is(err, closeErr) {
		t.Errorf("err = %v, want it to include the forced-close failure", err)
	}
}

// --- expected closure vs. real failure --------------------------------------

func TestSuperviseHTTPServer_ErrServerClosedAfterShutdownIsExpected(t *testing.T) {
	rec := &shutdownRecorder{}
	srv, serve, shutdownCtx := newGracefulStub(rec.add)

	ctx, sendSignal := context.WithCancel(context.Background())
	sendSignal() // the signal context is already cancelled when shutdown starts

	if err := superviseHTTPServer(ctx, srv, serve, httpShutdownTimeout); err != nil {
		t.Fatalf("http.ErrServerClosed after a requested shutdown must not be an error, got: %v", err)
	}

	requireEvents(t, rec.snapshot(), []string{"shutdown"})
	shutdownCtx.requireFresh(t, "Shutdown", httpShutdownTimeout)
}

func TestSuperviseHTTPServer_ErrServerClosedWithoutShutdownIsUnexpected(t *testing.T) {
	rec := &shutdownRecorder{}
	srv, _, _ := newGracefulStub(rec.add)
	serve := func() error { return http.ErrServerClosed }

	// Nothing asked the server to stop, so its exit is a failure, not success.
	err := superviseHTTPServer(context.Background(), srv, serve, httpShutdownTimeout)

	if !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("err = %v, want the unrequested closure reported", err)
	}
}

func TestSuperviseHTTPServer_ServerErrorRacingSignalIsNotLost(t *testing.T) {
	stopped := make(chan struct{})
	srv := &stubHTTPServer{
		shutdownFn: func(context.Context) error { close(stopped); return nil },
		closeFn:    func() error { return nil },
	}
	serve := func() error {
		<-stopped
		return errSimulatedServerFailure
	}

	ctx, sendSignal := context.WithCancel(context.Background())
	sendSignal()

	err := superviseHTTPServer(ctx, srv, serve, httpShutdownTimeout)

	if !errors.Is(err, errSimulatedServerFailure) {
		t.Fatalf("err = %v, want the genuine server error reported", err)
	}
}

func TestSuperviseHTTPServer_UnexpectedFailureIsReportedAndServerStopped(t *testing.T) {
	rec := &shutdownRecorder{}
	srv, _, shutdownCtx := newGracefulStub(rec.add)
	serve := func() error { return errSimulatedServerFailure }

	// context.Background() is never cancelled: this is the failure path, not a
	// signal.
	err := superviseHTTPServer(context.Background(), srv, serve, httpShutdownTimeout)

	if !errors.Is(err, errSimulatedServerFailure) {
		t.Fatalf("err = %v, want the server failure propagated", err)
	}

	requireEvents(t, rec.snapshot(), []string{"shutdown"})
	shutdownCtx.requireFresh(t, "Shutdown", httpShutdownTimeout)
}

// --- exit status and cleanup ordering ---------------------------------------

func TestServeUntilShutdown_SignalShutsDownThenDrains(t *testing.T) {
	rec := &shutdownRecorder{}
	srv, serve, _ := newGracefulStub(rec.add)

	var drainCtx ctxSnapshot
	drain := func(ctx context.Context) error {
		drainCtx = snapshotCtx(ctx)
		rec.add("drain")
		return nil
	}

	ctx, sendSignal := context.WithCancel(context.Background())
	sendSignal()

	if code := serveUntilShutdown(ctx, srv, serve, httpShutdownTimeout, drain); code != 0 {
		t.Fatalf("exit code = %d, want 0 after a clean signal shutdown", code)
	}

	// HTTP first, so requests finishing during shutdown can still emit telemetry.
	requireEvents(t, rec.snapshot(), []string{"shutdown", "drain"})
	drainCtx.requireFresh(t, "telemetry drain", telemetryDrainTimeout)
}

func TestServeUntilShutdown_DrainFailureDoesNotChangeExitCode(t *testing.T) {
	rec := &shutdownRecorder{}
	srv, serve, _ := newGracefulStub(rec.add)
	drain := func(context.Context) error { return errors.New("collector unreachable") }

	ctx, sendSignal := context.WithCancel(context.Background())
	sendSignal()

	if code := serveUntilShutdown(ctx, srv, serve, httpShutdownTimeout, drain); code != 0 {
		t.Fatalf("exit code = %d, want 0: an incomplete drain is only logged", code)
	}
}

// failureScenario has the shape of runApplication: cleanup is deferred, and the
// exit code comes back from serveUntilShutdown instead of from a Fatal call.
// The server fails on its own, with no signal.
func failureScenario(emit func(string)) int {
	defer emit("cleanup")

	srv := &stubHTTPServer{
		shutdownFn: func(context.Context) error { emit("shutdown"); return nil },
		closeFn:    func() error { emit("close"); return nil },
	}
	serve := func() error { return errSimulatedServerFailure }
	drain := func(context.Context) error { emit("drain"); return nil }

	return serveUntilShutdown(context.Background(), srv, serve, httpShutdownTimeout, drain)
}

// This test finishing at all is part of the proof: the old logrus.Fatalf would
// have terminated the test binary here.
func TestServeUntilShutdown_UnexpectedFailureCleansUpAndReturnsFailure(t *testing.T) {
	rec := &shutdownRecorder{}

	code := failureScenario(rec.add)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}

	// HTTP stopped, telemetry drained, and only then the deferred cleanup.
	requireEvents(t, rec.snapshot(), []string{"shutdown", "drain", "cleanup"})
}

func TestServeUntilShutdown_ListenFailureOnRealServer(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer occupied.Close()

	srv := &http.Server{Addr: occupied.Addr().String(), Handler: http.NotFoundHandler()}

	// The timeout turns a (wrongly) successful bind into a failed assertion
	// instead of a hang.
	ctx, cancel := context.WithTimeout(context.Background(), testWaitTimeout)
	defer cancel()

	drained := false
	drain := func(context.Context) error { drained = true; return nil }

	code := serveUntilShutdown(ctx, srv, srv.ListenAndServe, httpShutdownTimeout, drain)

	if code != 1 {
		t.Errorf("exit code = %d, want 1 when the address is already in use", code)
	}
	if !drained {
		t.Error("telemetry was not drained after the listener failed")
	}
}

const failureHelperEnv = "PENTAGI_TEST_SERVER_FAILURE_HELPER"

// TestServerFailureHelperProcess is not a test of its own: it is the child
// process of TestUnexpectedServerFailureExitStatus, and skips otherwise.
func TestServerFailureHelperProcess(t *testing.T) {
	if os.Getenv(failureHelperEnv) != "1" {
		t.Skip("helper process for TestUnexpectedServerFailureExitStatus")
	}

	// The same pattern as main: exit only with the code a returned function chose.
	os.Exit(failureScenario(func(event string) { fmt.Println("EVENT:" + event) }))
}

// Proves the real process exit status, not just a returned integer: the child
// must exit non-zero, and its deferred cleanup must have run before it did.
// It covers serveUntilShutdown and the os.Exit(run()) pattern; runApplication
// itself needs PostgreSQL and Docker and is not executed here.
func TestUnexpectedServerFailureExitStatus(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestServerFailureHelperProcess$")
	cmd.Env = append(os.Environ(), failureHelperEnv+"=1")

	out, err := cmd.Output()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("child err = %v, want a non-zero exit\nstdout:\n%s", err, out)
	}
	if got := exitErr.ExitCode(); got != 1 {
		t.Fatalf("child exit status = %d, want 1\nstdout:\n%s\nstderr:\n%s", got, out, exitErr.Stderr)
	}

	var events []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, "EVENT:") {
			events = append(events, strings.TrimPrefix(line, "EVENT:"))
		}
	}

	requireEvents(t, events, []string{"shutdown", "drain", "cleanup"})
}
