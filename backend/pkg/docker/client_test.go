package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"pentagi/pkg/database"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
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

const probeImage = "alpine:3.23.5"

// newDaemonClient binds a client to the local daemon, skipping the test when
// none is reachable.
func newDaemonClient(t *testing.T) *dockerClient {
	t.Helper()

	cli, err := client.New(client.FromEnv)
	if err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}

	ctx := t.Context()
	if _, err := cli.Ping(ctx, client.PingOptions{NegotiateAPIVersion: true}); err != nil {
		t.Skipf("docker daemon unavailable: %v", err)
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	return &dockerClient{client: cli, logger: logger}
}

func TestIsContainerRunningRemovedContainer(t *testing.T) {
	dc := newDaemonClient(t)
	ctx := t.Context()

	created, err := dc.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:      probeImage,
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
	})
	if cerrdefs.IsNotFound(err) {
		t.Skipf("%s is not present locally", probeImage)
	}
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.WithoutCancel(ctx)
		// the happy path already removes the container below; only report a
		// cleanup failure if it is still there for some other reason.
		if _, err := dc.client.ContainerRemove(ctx, created.ID, client.ContainerRemoveOptions{Force: true}); err != nil && !cerrdefs.IsNotFound(err) {
			t.Errorf("cleanup: failed to remove container %q: %v", created.ID, err)
		}
	})

	_, err = dc.client.ContainerStart(ctx, created.ID, client.ContainerStartOptions{})
	require.NoError(t, err)

	running, err := dc.IsContainerRunning(ctx, created.ID)
	require.NoError(t, err)
	require.True(t, running)

	_, err = dc.client.ContainerRemove(ctx, created.ID, client.ContainerRemoveOptions{Force: true})
	require.NoError(t, err)

	running, err = dc.IsContainerRunning(ctx, created.ID)
	require.NoError(t, err)
	require.False(t, running)
}

func TestIsContainerRunningUnknownContainer(t *testing.T) {
	dc := newDaemonClient(t)

	running, err := dc.IsContainerRunning(t.Context(), "pentagi-container-that-does-not-exist")

	require.NoError(t, err)
	require.False(t, running)
}

// The tests below drive RunContainer against the local daemon and assert on the
// daemon's own view of the result, so the container spec is proven as the engine
// actually applied it rather than as the caller intended it.

// containerRecorder captures the database writes RunContainer performs. The
// embedded nil Querier makes any query the code under test starts issuing panic
// loudly instead of silently returning zero values.
type containerRecorder struct {
	database.Querier

	row      database.Container
	created  database.CreateContainerParams
	statuses []database.ContainerStatus
	images   []string
}

func (r *containerRecorder) CreateContainer(
	_ context.Context, arg database.CreateContainerParams,
) (database.Container, error) {
	r.created = arg
	r.row = database.Container{
		ID:       1,
		Type:     arg.Type,
		Name:     arg.Name,
		Image:    arg.Image,
		Status:   arg.Status,
		LocalID:  arg.LocalID,
		LocalDir: arg.LocalDir,
		FlowID:   arg.FlowID,
	}
	return r.row, nil
}

func (r *containerRecorder) UpdateContainerImage(
	_ context.Context, arg database.UpdateContainerImageParams,
) (database.Container, error) {
	r.images = append(r.images, arg.Image)
	r.row.Image = arg.Image
	return r.row, nil
}

func (r *containerRecorder) UpdateContainerStatusLocalID(
	_ context.Context, arg database.UpdateContainerStatusLocalIDParams,
) (database.Container, error) {
	r.statuses = append(r.statuses, arg.Status)
	r.row.Status = arg.Status
	r.row.LocalID = arg.LocalID
	return r.row, nil
}

// newRunContainerClient extends the daemon-bound client with everything
// RunContainer additionally needs — a recording database and a private data
// directory — leaving every branch-selecting field at its neutral default so a
// test only sets the one field its branch is about.
func newRunContainerClient(t *testing.T) (*dockerClient, *containerRecorder) {
	t.Helper()

	dc := newDaemonClient(t)
	recorder := &containerRecorder{}
	dc.db = recorder
	dc.dataDir = t.TempDir()
	dc.defImage = probeImage
	dc.labels = map[string]string{"pentagi.test": t.Name()}

	// Pull up front so a slow or unreachable registry surfaces as a skip here
	// rather than as a confusing failure inside a branch assertion.
	if err := dc.pullImage(t.Context(), probeImage); err != nil {
		t.Skipf("probe image %s unavailable: %v", probeImage, err)
	}

	return dc, recorder
}

// probeName builds a daemon-unique container name. RunContainer derives both the
// work volume name and the container hostname from it, so collisions between
// runs would corrupt exactly the things these tests assert on.
func probeName(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("pentagi-probe-%d", rand.Uint64())
}

// probeConfig returns a fresh config on every call: RunContainer mutates the one
// it is handed (hostname, working dir, labels, exposed ports), so tests must
// never share a single value.
func probeConfig() *container.Config {
	return &container.Config{
		Image:      probeImage,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
	}
}

// runProbeContainer runs a container through RunContainer and returns the row it
// produced together with the daemon's view of the container. Cleanup is armed
// before the run so a container created by a half-failed run is still reaped.
func runProbeContainer(
	t *testing.T,
	dc *dockerClient,
	name string,
	flowID int64,
	config *container.Config,
	hostConfig *container.HostConfig,
) (database.Container, container.InspectResponse) {
	t.Helper()

	if config == nil {
		config = probeConfig()
	}
	cleanupProbeContainer(t, dc, name)

	row, err := dc.RunContainer(t.Context(), name, database.ContainerTypePrimary, flowID, config, hostConfig)
	require.NoError(t, err)
	require.NotEmpty(t, row.LocalID.String)

	inspect, err := dc.client.ContainerInspect(t.Context(), row.LocalID.String, client.ContainerInspectOptions{})
	require.NoError(t, err)

	return row, inspect.Container
}

// cleanupProbeContainer removes the container and the work volume RunContainer
// may have created for it. The volume is named after the container and survives
// RemoveVolumes (which only reaps anonymous ones), so it has to go explicitly.
func cleanupProbeContainer(t *testing.T, dc *dockerClient, name string) {
	t.Helper()

	ctx := context.WithoutCancel(t.Context())
	t.Cleanup(func() {
		_, err := dc.client.ContainerRemove(ctx, name, client.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil && !cerrdefs.IsNotFound(err) {
			t.Errorf("cleanup: failed to remove container %q: %v", name, err)
		}

		_, err = dc.client.VolumeRemove(ctx, name+WorkerVolumeNameSuffix, client.VolumeRemoveOptions{Force: true})
		if err != nil && !cerrdefs.IsNotFound(err) {
			t.Errorf("cleanup: failed to remove volume for %q: %v", name, err)
		}
	})
}

// mountAt returns the mount the daemon reports at dst. The mount table is how
// /work provisioning and Docker-in-Docker wiring are proven, so a missing entry
// must fail rather than degrade into a zero-value comparison.
func mountAt(t *testing.T, inspect container.InspectResponse, dst string) container.MountPoint {
	t.Helper()

	for _, mountPoint := range inspect.Mounts {
		if mountPoint.Destination == dst {
			return mountPoint
		}
	}

	t.Fatalf("container has no mount at %q, mounts: %+v", dst, inspect.Mounts)
	return container.MountPoint{}
}

// requireMountSource asserts a mount resolves to the given host path. Docker
// Desktop rewrites bind sources into its VM namespace (/host_mnt/private/...),
// so the daemon-reported source is matched by suffix instead of by equality;
// the paths under test are unique temporary directories, so a suffix match
// still pins the mount to exactly one host location.
func requireMountSource(t *testing.T, mountPoint container.MountPoint, hostPath string) {
	t.Helper()

	require.Truef(t, strings.HasSuffix(mountPoint.Source, hostPath),
		"mount at %q has source %q, which does not resolve to host path %q",
		mountPoint.Destination, mountPoint.Source, hostPath)
}

// reserveFreePortBase picks a ports base whose derived flow ports are all free
// right now, so asserting on published ports cannot collide with whatever else
// happens to listen on the machine running the test.
func reserveFreePortBase(t *testing.T, flowID int64) int {
	t.Helper()

	for range 50 {
		// stay below the 63535 ceiling GetPrimaryContainerPorts enforces, so the
		// requested base is the one actually used
		base := 30000 + rand.Intn(20000)
		if portsAreFree(GetPrimaryContainerPorts(base, flowID)) {
			return base
		}
	}

	t.Skip("no free port range available for the published-port test")
	return 0
}

func portsAreFree(ports []int) bool {
	for _, port := range ports {
		listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			return false
		}
		listener.Close()
	}
	return true
}

// The sandbox hardening RunContainer applies unconditionally, plus the database
// lifecycle: a row is inserted as starting and flipped to running with the real
// container id once the daemon has started it.
func TestRunContainerAppliesSandboxDefaults(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	name := probeName(t)

	row, inspect := runProbeContainer(t, dc, name, 7, nil, nil)

	require.Equal(t, fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(name))), inspect.Config.Hostname)
	require.Equal(t, WorkFolderPathInContainer, inspect.Config.WorkingDir)
	require.Equal(t, t.Name(), inspect.Config.Labels["pentagi.test"])

	require.Equal(t, container.RestartPolicyOnFailure, inspect.HostConfig.RestartPolicy.Name)
	require.Equal(t, 5, inspect.HostConfig.RestartPolicy.MaximumRetryCount)
	require.NotNil(t, inspect.HostConfig.PidsLimit)
	require.Equal(t, int64(2048), *inspect.HostConfig.PidsLimit)
	require.Equal(t, "json-file", inspect.HostConfig.LogConfig.Type)
	require.Equal(t, "10m", inspect.HostConfig.LogConfig.Config["max-size"])
	require.Equal(t, "5", inspect.HostConfig.LogConfig.Config["max-file"])

	require.Equal(t, database.ContainerStatusStarting, recorder.created.Status)
	require.Equal(t, database.ContainerTypePrimary, recorder.created.Type)
	require.Equal(t, []database.ContainerStatus{database.ContainerStatusRunning}, recorder.statuses)
	require.Equal(t, inspect.ID, row.LocalID.String)
	require.Equal(t, database.ContainerStatusRunning, row.Status)

	running, err := dc.IsContainerRunning(t.Context(), row.LocalID.String)
	require.NoError(t, err)
	require.True(t, running)
}

// A caller-supplied host config is extended, not replaced: its pids limit
// survives the default and its binds keep the appended /work mount company.
func TestRunContainerPreservesCallerHostConfig(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	callerLimit := int64(64)

	_, inspect := runProbeContainer(t, dc, probeName(t), 8, nil, &container.HostConfig{
		Resources: container.Resources{PidsLimit: &callerLimit},
	})

	require.NotNil(t, inspect.HostConfig.PidsLimit)
	require.Equal(t, callerLimit, *inspect.HostConfig.PidsLimit)
	require.Len(t, inspect.HostConfig.Binds, 1)
	require.Contains(t, inspect.HostConfig.Binds[0], ":"+WorkFolderPathInContainer)
}

// Without a host-side data directory /work has to come from a named, labelled
// volume owned by the container, and the database row records no host path.
func TestRunContainerBacksWorkDirWithVolume(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	dc.hostDir = ""
	name := probeName(t)

	_, inspect := runProbeContainer(t, dc, name, 3, nil, nil)

	workMount := mountAt(t, inspect, WorkFolderPathInContainer)
	require.Equal(t, mount.TypeVolume, workMount.Type)
	require.Equal(t, name+WorkerVolumeNameSuffix, workMount.Name)
	require.True(t, workMount.RW)
	require.Empty(t, recorder.created.LocalDir.String)

	volume, err := dc.client.VolumeInspect(t.Context(), name+WorkerVolumeNameSuffix, client.VolumeInspectOptions{})
	require.NoError(t, err)
	require.Equal(t, "local", volume.Volume.Driver)
	require.Equal(t, t.Name(), volume.Volume.Labels["pentagi.test"])

	// the mount is a usable directory, not just a spec the daemon accepted
	listing, err := dc.ListContainerDir(t.Context(), inspect.ID, WorkFolderPathInContainer)
	require.NoError(t, err)
	require.Empty(t, listing.Files)
}

// With a host-side data directory /work is a bind of the per-flow subdirectory,
// which is created on the host and recorded on the row.
func TestRunContainerBindsPerFlowHostDir(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	// the daemon runs on this machine, so the host path equals the local path
	dc.hostDir = dc.dataDir
	flowID := int64(11)
	flowDir := filepath.Join(dc.dataDir, fmt.Sprintf(containerLocalCwdTemplate, flowID))

	_, inspect := runProbeContainer(t, dc, probeName(t), flowID, nil, nil)

	require.DirExists(t, flowDir)
	require.Equal(t, flowDir, recorder.created.LocalDir.String)

	workMount := mountAt(t, inspect, WorkFolderPathInContainer)
	require.Equal(t, mount.TypeBind, workMount.Type)
	requireMountSource(t, workMount, flowDir)
	require.True(t, workMount.RW)

	// the bind is live: a file written on the host is visible inside the sandbox
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "marker.txt"), []byte("payload"), 0o600))
	stat, err := dc.ContainerStatPath(t.Context(), inspect.ID, WorkFolderPathInContainer+"/marker.txt")
	require.NoError(t, err)
	require.Equal(t, "marker.txt", stat.Name)
	require.Equal(t, int64(len("payload")), stat.Size)
}

// In bridge mode the flow's ports are exposed in the image config and published
// on the configured public IP, which is what makes reverse connections reachable.
func TestRunContainerPublishesFlowPorts(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	dc.publicIP = "127.0.0.1"
	flowID := int64(4)
	dc.portsBase = reserveFreePortBase(t, flowID)

	config := probeConfig()
	_, inspect := runProbeContainer(t, dc, probeName(t), flowID, config, nil)

	ports := GetPrimaryContainerPorts(dc.portsBase, flowID)
	require.Len(t, ports, containerPortsNumber)
	for _, port := range ports {
		containerPort, ok := network.PortFrom(uint16(port), network.TCP)
		require.True(t, ok)

		require.Contains(t, config.ExposedPorts, containerPort)

		published := inspect.NetworkSettings.Ports[containerPort]
		require.Len(t, published, 1)
		require.Equal(t, strconv.Itoa(port), published[0].HostPort)
		require.Equal(t, "127.0.0.1", published[0].HostIP.String())
	}
}

// An unparseable public IP is rejected before anything reaches the daemon, so a
// misconfigured deployment cannot leave half-created sandboxes behind.
func TestRunContainerRejectsInvalidPublicIP(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	dc.publicIP = "definitely-not-an-ip"
	name := probeName(t)
	cleanupProbeContainer(t, dc, name)

	_, err := dc.RunContainer(t.Context(), name, database.ContainerTypePrimary, 2, probeConfig(), nil)
	require.ErrorContains(t, err, "invalid Docker public IP")

	_, inspectErr := dc.client.ContainerInspect(t.Context(), name, client.ContainerInspectOptions{})
	require.True(t, cerrdefs.IsNotFound(inspectErr), "no container must exist, got: %v", inspectErr)
	require.Equal(t, []database.ContainerStatus{database.ContainerStatusFailed}, recorder.statuses)
}

// A missing config is refused before a database row is inserted, so a caller
// bug cannot leave an orphan row behind.
func TestRunContainerRejectsMissingConfig(t *testing.T) {
	dc, recorder := newRunContainerClient(t)

	_, err := dc.RunContainer(t.Context(), probeName(t), database.ContainerTypePrimary, 1, nil, nil)

	require.ErrorContains(t, err, "no config found")
	require.Empty(t, recorder.created.Name)
	require.Empty(t, recorder.statuses)
}

// When neither the requested nor the default image can be pulled the row is
// marked failed, which is what lets cleanup treat the flow as dead instead of
// leaving it stuck in starting.
func TestRunContainerMarksRowFailedWhenNoImageAvailable(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	unavailable := probeName(t) + ":0" // a repository that cannot exist
	dc.defImage = unavailable

	config := probeConfig()
	config.Image = unavailable

	_, err := dc.RunContainer(t.Context(), probeName(t), database.ContainerTypePrimary, 20, config, nil)

	require.ErrorContains(t, err, "failed to pull default image")
	require.Equal(t, []string{unavailable}, recorder.images)
	require.Equal(t, []database.ContainerStatus{database.ContainerStatusFailed}, recorder.statuses)
}

// A container that is created but cannot start is recorded as failed together
// with its local id, and nothing is left behind on the host.
func TestRunContainerMarksRowFailedWhenStartFails(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	dc.publicIP = "127.0.0.1"
	flowID := int64(21)
	dc.portsBase = reserveFreePortBase(t, flowID)

	// hold one of the flow's host ports so the daemon accepts the create but
	// refuses the start with "port is already allocated"
	ports := GetPrimaryContainerPorts(dc.portsBase, flowID)
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(ports[0])))
	require.NoError(t, err)
	defer listener.Close()

	name := probeName(t)
	cleanupProbeContainer(t, dc, name)

	_, err = dc.RunContainer(t.Context(), name, database.ContainerTypePrimary, flowID, probeConfig(), nil)

	require.ErrorContains(t, err, "failed to start container")
	require.Equal(t, []database.ContainerStatus{database.ContainerStatusFailed}, recorder.statuses)
	require.NotEmpty(t, recorder.row.LocalID.String)

	_, inspectErr := dc.client.ContainerInspect(t.Context(), name, client.ContainerInspectOptions{})
	require.True(t, cerrdefs.IsNotFound(inspectErr), "container must not be left behind, got: %v", inspectErr)
}

// The product hands the container straight to the agents, so a sandbox whose
// entrypoint does not survive the start must be reported as a failure here —
// whether it crashed or simply ran to completion — instead of being handed over
// dead. The daemon acknowledges such a start as successful, so this is the only
// place the situation can still be recognised.
func TestRunContainerRejectsContainerThatDoesNotStayRunning(t *testing.T) {
	tests := []struct {
		name       string
		entrypoint []string
		exitCode   int
	}{
		{
			name:       "entrypoint crashes",
			entrypoint: []string{"sh", "-c", "echo startup-diagnostic >&2; exit 3"},
			exitCode:   3,
		},
		{
			name:       "entrypoint runs to completion",
			entrypoint: []string{"sh", "-c", "echo startup-diagnostic; exit 0"},
			exitCode:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dc, recorder := newRunContainerClient(t)
			name := probeName(t)
			cleanupProbeContainer(t, dc, name)

			config := probeConfig()
			config.Entrypoint = test.entrypoint

			_, err := dc.RunContainer(t.Context(), name, database.ContainerTypePrimary, 31, config, nil)

			var startupErr *ContainerStartupError
			require.ErrorAs(t, err, &startupErr)
			require.Equal(t, name, startupErr.ContainerName)
			require.Equal(t, test.exitCode, startupErr.ExitCode)
			// the container's own output travels with the error, so the reason is
			// visible without reaching for the host
			require.Contains(t, startupErr.LogTail, "startup-diagnostic")

			require.Equal(t, []database.ContainerStatus{database.ContainerStatusFailed}, recorder.statuses)

			_, inspectErr := dc.client.ContainerInspect(t.Context(), name, client.ContainerInspectOptions{})
			require.True(t, cerrdefs.IsNotFound(inspectErr),
				"dead container must not be left behind, got: %v", inspectErr)
		})
	}
}

// A container left behind by an earlier run holds the name and would otherwise
// make every later create fail with a name conflict. It is also not running, so
// finding it at all requires listing every container, not just the live ones.
func TestRunContainerReplacesContainerHoldingTheName(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	name := probeName(t)

	stale, err := dc.client.ContainerCreate(t.Context(), client.ContainerCreateOptions{
		Config: &container.Config{Image: probeImage, Entrypoint: []string{"true"}},
		Name:   name,
	})
	require.NoError(t, err)

	row, inspect := runProbeContainer(t, dc, name, 30, nil, nil)

	require.NotEqual(t, stale.ID, row.LocalID.String)
	require.NotNil(t, inspect.State)
	require.True(t, inspect.State.Running)

	_, inspectErr := dc.client.ContainerInspect(t.Context(), stale.ID, client.ContainerInspectOptions{})
	require.True(t, cerrdefs.IsNotFound(inspectErr),
		"the container holding the name must be removed, got: %v", inspectErr)
}

// Host network mode hands the container the host stack directly, so port
// publishing must be skipped entirely even when a public IP and base are set.
func TestRunContainerHostNetworkSkipsPortPublishing(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	dc.network = "host"
	dc.publicIP = "127.0.0.1"
	dc.portsBase = BaseContainerPortsNumber

	config := probeConfig()
	_, inspect := runProbeContainer(t, dc, probeName(t), 5, config, nil)

	require.Equal(t, container.NetworkMode("host"), inspect.HostConfig.NetworkMode)
	require.Empty(t, config.ExposedPorts)
	require.Empty(t, inspect.HostConfig.PortBindings)
	require.Empty(t, inspect.NetworkSettings.Ports)
	require.Contains(t, inspect.NetworkSettings.Networks, "host")

	// /work is still provisioned; host networking only changes reachability
	require.Equal(t, WorkFolderPathInContainer, mountAt(t, inspect, WorkFolderPathInContainer).Destination)
}

// A configured bridge network is created on demand and the sandbox is attached
// to it, which is what isolates workers from the default bridge.
func TestRunContainerAttachesConfiguredNetwork(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	networkName := probeName(t)
	ctx := context.WithoutCancel(t.Context())

	require.NoError(t, ensureDockerNetwork(t.Context(), dc.client, networkName))
	t.Cleanup(func() {
		if _, err := dc.client.NetworkRemove(ctx, networkName, client.NetworkRemoveOptions{}); err != nil &&
			!cerrdefs.IsNotFound(err) {
			t.Errorf("cleanup: failed to remove network %q: %v", networkName, err)
		}
	})
	// a second call must be a no-op rather than a duplicate-network failure
	require.NoError(t, ensureDockerNetwork(t.Context(), dc.client, networkName))

	dc.network = networkName
	dc.publicIP = "127.0.0.1"
	flowID := int64(6)
	dc.portsBase = reserveFreePortBase(t, flowID)

	_, inspect := runProbeContainer(t, dc, probeName(t), flowID, nil, nil)

	require.Contains(t, inspect.NetworkSettings.Networks, networkName)
	require.NotEmpty(t, inspect.NetworkSettings.Networks[networkName].IPAddress)
}

// An image that cannot be pulled falls back to the configured default image,
// and the fallback is reflected in the config, the database row and the daemon.
func TestRunContainerFallsBackToDefaultImage(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	config := probeConfig()
	config.Image = probeName(t) + ":0" // a repository that cannot exist

	row, inspect := runProbeContainer(t, dc, probeName(t), 9, config, nil)

	require.Equal(t, probeImage, config.Image)
	require.Equal(t, []string{probeImage}, recorder.images)
	require.Equal(t, probeImage, row.Image)
	require.Equal(t, probeImage, inspect.Config.Image)
}

// With DOCKER_INSIDE the sandbox gets the designated daemon socket, the
// matching client environment, and the TLS material at the same read-only path
// on both sides.
func TestRunContainerWiresDockerInsideAccess(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	socketPath := filepath.Join(t.TempDir(), "docker.sock")
	require.NoError(t, os.WriteFile(socketPath, nil, 0o600))
	certPath := t.TempDir()

	dc.inside = true
	dc.socket = socketPath
	dc.insideCertPath = certPath
	dc.insideEnv = []string{"DOCKER_HOST=tcp://daemon.internal:2376", "DOCKER_TLS_VERIFY=1"}

	config := probeConfig()
	_, inspect := runProbeContainer(t, dc, probeName(t), 12, config, nil)

	socketMount := mountAt(t, inspect, defaultDockerSocketPath)
	requireMountSource(t, socketMount, socketPath)

	certMount := mountAt(t, inspect, certPath)
	requireMountSource(t, certMount, certPath)
	require.False(t, certMount.RW)

	require.Subset(t, inspect.Config.Env, dc.insideEnv)
}

// Without DOCKER_INSIDE no daemon socket is handed to the sandbox at all.
func TestRunContainerWithoutInsideKeepsDaemonUnreachable(t *testing.T) {
	dc, _ := newRunContainerClient(t)

	_, inspect := runProbeContainer(t, dc, probeName(t), 13, nil, nil)

	for _, mountPoint := range inspect.Mounts {
		require.NotEqual(t, defaultDockerSocketPath, mountPoint.Destination)
	}
	require.Len(t, inspect.HostConfig.Binds, 1)
}
