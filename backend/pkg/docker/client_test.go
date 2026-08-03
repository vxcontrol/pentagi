package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func TestStatContainerEntries_AllSucceed(t *testing.T) {
	names := []string{"c", "a", "b", "z", "m"}
	stats, failures := statContainerEntries(context.Background(), names, 20, okStat)
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %v", failures)
	}
	if len(stats) != len(names) {
		t.Fatalf("got %d stats, want %d", len(stats), len(names))
	}
}

func TestStatContainerEntries_Empty(t *testing.T) {
	stats, failures := statContainerEntries(context.Background(), nil, 20, okStat)
	if len(stats) != 0 || len(failures) != 0 {
		t.Fatalf("got %d stats / %d failures, want 0/0", len(stats), len(failures))
	}
}

// A per-entry stat error must NOT abort the batch: the readable entries are
// returned and the failing ones come back as failures — partial success.
func TestStatContainerEntries_PartialSuccessNotAborting(t *testing.T) {
	names := []string{"e0", "e1", "e2", "e3", "e4"}
	fail := map[string]bool{"e1": true, "e3": true}
	statFn := func(_ context.Context, name string) (container.PathStat, error) {
		if fail[name] {
			return container.PathStat{}, fmt.Errorf("boom %s", name)
		}
		return container.PathStat{Name: name}, nil
	}
	stats, failures := statContainerEntries(context.Background(), names, 20, statFn)

	fn := failNames(failures)
	if len(fn) != 2 || !fn["e1"] || !fn["e3"] {
		t.Fatalf("want failures {e1,e3}, got %v", fn)
	}
	wantOK := []string{"e0", "e2", "e4"} // successes, preserved in input order
	if len(stats) != len(wantOK) {
		t.Fatalf("want %d readable entries, got %d", len(wantOK), len(stats))
	}
	for i, s := range stats {
		if s.Name != wantOK[i] {
			t.Fatalf("success order at %d: got %q want %q", i, s.Name, wantOK[i])
		}
	}
}

// The pool never runs more than `workers` stat calls at once — the load-bearing
// bound against the Docker daemon.
func TestStatContainerEntries_RespectsConcurrencyLimit(t *testing.T) {
	const workers = 5
	names := make([]string, 100)
	for i := range names {
		names[i] = fmt.Sprintf("e%d", i)
	}
	var cur, max atomic.Int64
	statFn := func(_ context.Context, name string) (container.PathStat, error) {
		c := cur.Add(1)
		for {
			m := max.Load()
			if c <= m || max.CompareAndSwap(m, c) {
				break
			}
		}
		time.Sleep(2 * time.Millisecond)
		cur.Add(-1)
		return container.PathStat{Name: name}, nil
	}
	if _, failures := statContainerEntries(context.Background(), names, workers, statFn); len(failures) != 0 {
		t.Fatalf("unexpected failures: %v", failures)
	}
	if got := max.Load(); got > workers {
		t.Fatalf("concurrency exceeded limit: peak=%d, limit=%d", got, workers)
	} else if got < 2 {
		t.Fatalf("statFn never overlapped (peak=%d) — test is not exercising concurrency", got)
	}
}

// The caller's context reaches each stat call; a cancellation turns entries into
// failures rather than blanking the whole batch.
func TestStatContainerEntries_ContextPropagates(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	names := []string{"a", "b", "c", "d", "e"}
	var once atomic.Bool
	statFn := func(ctx context.Context, name string) (container.PathStat, error) {
		if once.CompareAndSwap(false, true) {
			cancel()
		}
		select {
		case <-ctx.Done():
			return container.PathStat{}, ctx.Err()
		case <-time.After(2 * time.Second):
			return container.PathStat{Name: name}, nil
		}
	}
	_, failures := statContainerEntries(ctx, names, 20, statFn)
	if len(failures) == 0 {
		t.Fatal("expected failures once the context is cancelled")
	}
	for _, f := range failures {
		if !errors.Is(f.err, context.Canceled) {
			t.Fatalf("failure %q: want context.Canceled, got %v", f.name, f.err)
		}
	}
}

// A non-positive worker count must fall back to a safe bound rather than
// deadlock (errgroup SetLimit(0)) or run unbounded (SetLimit(<0)).
func TestStatContainerEntries_NonPositiveWorkersFallBack(t *testing.T) {
	names := make([]string, 50)
	for i := range names {
		names[i] = fmt.Sprintf("e%d", i)
	}
	for _, workers := range []int{0, -1} {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			done := make(chan int, 1)
			go func() {
				stats, _ := statContainerEntries(context.Background(), names, workers, okStat)
				done <- len(stats)
			}()
			select {
			case n := <-done:
				if n != len(names) {
					t.Fatalf("got %d stats, want %d", n, len(names))
				}
			case <-time.After(3 * time.Second):
				t.Fatal("statContainerEntries hung with a non-positive worker count")
			}
		})
	}
}

// Property-based: over randomized shapes nothing is lost — every failing name
// lands in failures, every other name yields a stat, and stats+failures == n.
func TestStatContainerEntries_PropertyFuzz(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	for iter := 0; iter < 300; iter++ {
		n := rng.Intn(50)
		names := make([]string, n)
		wantFail := make(map[string]bool, n)
		for i := range names {
			names[i] = fmt.Sprintf("it%d-e%d", iter, i)
			if rng.Intn(3) == 0 {
				wantFail[names[i]] = true
			}
		}
		statFn := func(_ context.Context, name string) (container.PathStat, error) {
			if wantFail[name] {
				return container.PathStat{}, fmt.Errorf("boom %s", name)
			}
			return container.PathStat{Name: name}, nil
		}
		stats, failures := statContainerEntries(context.Background(), names, 20, statFn)

		if len(stats)+len(failures) != n {
			t.Fatalf("iter %d: lost data — %d stats + %d failures != %d names", iter, len(stats), len(failures), n)
		}
		gotFail := failNames(failures)
		if len(gotFail) != len(wantFail) {
			t.Fatalf("iter %d: want %d failures, got %d", iter, len(wantFail), len(gotFail))
		}
		for nm := range wantFail {
			if !gotFail[nm] {
				t.Fatalf("iter %d: missing failure for %q", iter, nm)
			}
		}
		for _, s := range stats {
			if wantFail[s.Name] {
				t.Fatalf("iter %d: %q failed but appears in stats", iter, s.Name)
			}
		}
	}
}

func TestParseFindEntries(t *testing.T) {
	entries, truncated := parseFindEntries([]byte("a\x00b\x00\x00c\x00"))
	if truncated {
		t.Error("small input must not truncate")
	}
	if len(entries) != 3 || entries[0] != "a" || entries[1] != "b" || entries[2] != "c" {
		t.Fatalf("split/skip-empty wrong: %q", entries)
	}

	// exactly at the cap → not truncated
	e, tr := parseFindEntries([]byte(strings.Repeat("x\x00", maxListEntries)))
	if tr || len(e) != maxListEntries {
		t.Fatalf("at cap: truncated=%v len=%d (want false, %d)", tr, len(e), maxListEntries)
	}

	// cap+1 → truncated, capped to maxListEntries
	e, tr = parseFindEntries([]byte(strings.Repeat("x\x00", maxListEntries+1)))
	if !tr || len(e) != maxListEntries {
		t.Fatalf("over cap: truncated=%v len=%d (want true, %d)", tr, len(e), maxListEntries)
	}
}

func TestDemuxExecStdout_StdoutOnly_AndByteCap(t *testing.T) {
	// stdout is returned; stderr (id 2) is discarded
	in := append(listingFrame(1, "hello"), listingFrame(2, "diagnostic")...)
	out, err := demuxExecStdout(bytes.NewReader(in), 1<<20)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if string(out) != "hello" {
		t.Fatalf("want stdout only, got %q", out)
	}

	// stdout exceeding the cap → error before materializing the full buffer
	_, err = demuxExecStdout(bytes.NewReader(listingFrame(1, strings.Repeat("x", 50))), 10)
	if err == nil || !strings.Contains(err.Error(), "listing output exceeded") {
		t.Fatalf("want cap error, got %v", err)
	}
}

func TestDemuxExecStdout_TruncatedAndSystemerr(t *testing.T) {
	// a header cut short mid-frame must error, not silently drop the tail
	torn := append(listingFrame(1, "a.txt\x00"), 0x01, 0x00, 0x00) // 3 stray header bytes
	if _, err := demuxExecStdout(bytes.NewReader(torn), 1<<20); err == nil || !strings.Contains(err.Error(), "truncated") {
		t.Fatalf("want truncated-stream error, got %v", err)
	}

	// a systemerr (stream id 3) daemon error must surface, not be discarded
	sys := append(listingFrame(1, "ok"), listingFrame(3, "daemon connection reset")...)
	if _, err := demuxExecStdout(bytes.NewReader(sys), 1<<20); err == nil || !strings.Contains(err.Error(), "systemerr") {
		t.Fatalf("want systemerr, got %v", err)
	}
}

func listingFrame(streamID byte, payload string) []byte {
	h := make([]byte, 8)
	h[0] = streamID
	binary.BigEndian.PutUint32(h[4:8], uint32(len(payload)))
	return append(h, []byte(payload)...)
}

func okStat(_ context.Context, name string) (container.PathStat, error) {
	return container.PathStat{Name: name, Size: int64(len(name))}, nil
}

func failNames(failures []statFailure) map[string]bool {
	m := make(map[string]bool, len(failures))
	for _, f := range failures {
		m[f.name] = true
	}
	return m
}

const probeImage = "alpine:3.20"

// newDaemonClient binds a client to the local daemon, skipping the test when
// none is reachable.
func newDaemonClient(t *testing.T) *dockerClient {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}

	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)
	if _, err := cli.Ping(ctx); err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}

	return &dockerClient{client: cli}
}

// A flow keeps the id of its primary container in the database. When that
// container is removed behind pentagi's back the id has to read as not running,
// otherwise the flow can never rebuild it.
func TestIsContainerRunningRemovedContainer(t *testing.T) {
	dc := newDaemonClient(t)
	ctx := context.Background()

	created, err := dc.client.ContainerCreate(ctx, &container.Config{
		Image:      probeImage,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
	}, nil, nil, nil, "")
	if client.IsErrNotFound(err) {
		t.Skipf("%s is not present locally", probeImage)
	}
	if err != nil {
		t.Fatalf("create probe container: %v", err)
	}
	t.Cleanup(func() {
		dc.client.ContainerRemove(context.Background(), created.ID, container.RemoveOptions{Force: true})
	})

	if err := dc.client.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		t.Fatalf("start probe container: %v", err)
	}

	running, err := dc.IsContainerRunning(ctx, created.ID)
	if err != nil || !running {
		t.Fatalf("got running=%v err=%v, want true and no error", running, err)
	}

	if err := dc.client.ContainerRemove(ctx, created.ID, container.RemoveOptions{Force: true}); err != nil {
		t.Fatalf("remove probe container: %v", err)
	}

	running, err = dc.IsContainerRunning(ctx, created.ID)
	if err != nil {
		t.Fatalf("removed container: got error %v, want none", err)
	}
	if running {
		t.Fatal("removed container reported as running")
	}
}

func TestIsContainerRunningUnknownContainer(t *testing.T) {
	dc := newDaemonClient(t)

	running, err := dc.IsContainerRunning(context.Background(), "pentagi-container-that-does-not-exist")
	if err != nil {
		t.Fatalf("unknown container: got error %v, want none", err)
	}
	if running {
		t.Fatal("unknown container reported as running")
	}
}
