package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"pentagi/pkg/database"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

func TestClient_StatContainerEntries_ReturnsEveryEntryAsAStatOrAFailure(t *testing.T) {
	tests := map[string]struct {
		names        []string
		failing      map[string]bool
		wantStats    []string
		wantFailures []string
	}{
		"no names": {},
		"every stat succeeds": {
			names:     []string{"c", "a", "b", "z", "m"},
			wantStats: []string{"c", "a", "b", "z", "m"},
		},
		"some stats fail": {
			names:        []string{"e0", "e1", "e2", "e3", "e4"},
			failing:      map[string]bool{"e1": true, "e3": true},
			wantStats:    []string{"e0", "e2", "e4"},
			wantFailures: []string{"e1: stat e1 failed", "e3: stat e3 failed"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// fewer workers than names, so results come back through a queued pool
			stats, failures := statContainerEntries(context.Background(), tc.names, 2,
				func(_ context.Context, name string) (container.PathStat, error) {
					if tc.failing[name] {
						return container.PathStat{}, fmt.Errorf("stat %s failed", name)
					}
					return container.PathStat{Name: name}, nil
				})

			var gotStats, gotFailures []string
			for _, stat := range stats {
				gotStats = append(gotStats, stat.Name)
			}
			for _, failure := range failures {
				gotFailures = append(gotFailures, failure.name+": "+failure.err.Error())
			}
			require.Equal(t, tc.wantStats, gotStats)
			require.Equal(t, tc.wantFailures, gotFailures)
		})
	}
}

func TestClient_StatContainerEntries_BoundsConcurrentStats(t *testing.T) {
	tests := map[string]struct {
		workers int
		limit   int64
	}{
		"the requested bound":                   {workers: 5, limit: 5},
		"a zero bound falls back to twenty":     {workers: 0, limit: 20},
		"a negative bound falls back to twenty": {workers: -1, limit: 20},
	}

	names := make([]string, 100)
	for i := range names {
		names[i] = fmt.Sprintf("e%d", i)
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// every stat holds until limit of them run at once, so the peak is the bound itself
			gate := make(chan struct{})
			var opened atomic.Bool
			stuck, release := context.WithTimeout(context.Background(), 5*time.Second)
			defer release()

			var running, peak atomic.Int64
			statFn := func(_ context.Context, name string) (container.PathStat, error) {
				current := running.Add(1)
				for {
					seen := peak.Load()
					if current <= seen || peak.CompareAndSwap(seen, current) {
						break
					}
				}
				if current == tc.limit && opened.CompareAndSwap(false, true) {
					close(gate)
				}
				select {
				case <-gate:
				case <-stuck.Done():
				}
				time.Sleep(2 * time.Millisecond)
				running.Add(-1)
				return container.PathStat{Name: name}, nil
			}

			done := make(chan [2]int, 1)
			go func() {
				stats, failures := statContainerEntries(context.Background(), names, tc.workers, statFn)
				done <- [2]int{len(stats), len(failures)}
			}()

			select {
			case counts := <-done:
				require.Equal(t, [2]int{len(names), 0}, counts, "stats and failures")
			case <-time.After(10 * time.Second):
				t.Fatal("statContainerEntries hung")
			}
			require.Equal(t, tc.limit, peak.Load(), "stats running at once")
		})
	}
}

func TestClient_StatContainerEntries_PassesTheCallerContextToEachStat(t *testing.T) {
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

	require.Len(t, failures, len(names), "a cancelled context must turn every entry into a failure")
	for _, failure := range failures {
		require.ErrorIs(t, failure.err, context.Canceled, failure.name)
	}
}

func TestClient_ParseFindEntries_SplitsOnNULAndCapsTheCount(t *testing.T) {
	tests := map[string]struct {
		output        string
		want          []string
		wantTruncated bool
	}{
		"empty tokens are dropped":           {"a\x00b\x00\x00c\x00", []string{"a", "b", "c"}, false},
		"exactly the cap is kept whole":      {strings.Repeat("x\x00", 10000), slices.Repeat([]string{"x"}, 10000), false},
		"one past the cap is cut to the cap": {strings.Repeat("x\x00", 10001), slices.Repeat([]string{"x"}, 10000), true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			entries, truncated := parseFindEntries([]byte(tc.output))

			require.Equal(t, tc.want, entries)
			require.Equal(t, tc.wantTruncated, truncated)
		})
	}
}

func TestClient_DemuxExecStdout_ReturnsStdoutOrRefusesTheStream(t *testing.T) {
	tests := map[string]struct {
		stream  []byte
		limit   int
		want    string
		wantErr string
	}{
		"stdout is kept and stderr dropped": {
			stream: slices.Concat(listingFrame(1, "hello"), listingFrame(2, "diagnostic")),
			limit:  1 << 20,
			want:   "hello",
		},
		"stdout beyond the cap is refused": {
			stream:  listingFrame(1, strings.Repeat("x", 50)),
			limit:   10,
			wantErr: "listing output exceeded 10 bytes",
		},
		"a header cut short is refused": {
			stream:  append(listingFrame(1, "a.txt\x00"), 0x01, 0x00, 0x00),
			limit:   1 << 20,
			wantErr: "truncated exec stream: unexpected EOF",
		},
		"a frame body cut short is refused": {
			stream:  listingFrame(1, "hello")[:10],
			limit:   1 << 20,
			wantErr: "EOF",
		},
		"a daemon systemerr is surfaced": {
			stream:  slices.Concat(listingFrame(1, "ok"), listingFrame(3, "daemon connection reset")),
			limit:   1 << 20,
			wantErr: "docker exec systemerr: daemon connection reset",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := demuxExecStdout(bytes.NewReader(tc.stream), tc.limit)

			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				require.Nil(t, out)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, string(out))
		})
	}
}

func listingFrame(streamID byte, payload string) []byte {
	h := make([]byte, 8)
	h[0] = streamID
	binary.BigEndian.PutUint32(h[4:8], uint32(len(payload)))
	return append(h, []byte(payload)...)
}

func TestClient_IsContainerRunning_ReportsRunningUntilTheContainerIsRemoved(t *testing.T) {
	dc := newDaemonClient(t)
	containerID := startProbeSandbox(t, dc, probeImage)

	running, err := dc.IsContainerRunning(t.Context(), containerID)
	require.NoError(t, err)
	require.True(t, running)

	_, err = dc.client.ContainerRemove(t.Context(), containerID, client.ContainerRemoveOptions{Force: true})
	require.NoError(t, err)

	running, err = dc.IsContainerRunning(t.Context(), containerID)
	require.NoError(t, err, "a removed container is missing, not an inspection failure")
	require.False(t, running)
}

// containerRecorder captures RunContainer's database writes; the nil Querier panics on any other query.
type containerRecorder struct {
	database.Querier

	row      database.Container
	created  database.CreateContainerParams
	statuses []database.ContainerStatus
	images   []string
}

func (r *containerRecorder) CreateContainer(
	ctx context.Context, arg database.CreateContainerParams,
) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}
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
	ctx context.Context, arg database.UpdateContainerImageParams,
) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}
	r.images = append(r.images, arg.Image)
	r.row.Image = arg.Image
	return r.row, nil
}

func (r *containerRecorder) UpdateContainerStatusLocalID(
	ctx context.Context, arg database.UpdateContainerStatusLocalIDParams,
) (database.Container, error) {
	if err := ctx.Err(); err != nil {
		return database.Container{}, err
	}
	r.statuses = append(r.statuses, arg.Status)
	r.row.Status = arg.Status
	r.row.LocalID = arg.LocalID
	return r.row, nil
}

// newRunContainerClient leaves every branch-selecting field neutral, so a test sets only the one it is about.
func newRunContainerClient(t *testing.T) (*dockerClient, *containerRecorder) {
	t.Helper()

	dc := newDaemonClient(t)
	recorder := &containerRecorder{}
	dc.db = recorder
	dc.dataDir = t.TempDir()
	dc.defImage = probeImage
	dc.labels = map[string]string{"pentagi.test": t.Name()}

	if err := dc.pullImage(t.Context(), probeImage); err != nil {
		t.Skipf("probe image %s unavailable: %v", probeImage, err)
	}

	return dc, recorder
}

// probeName is daemon-unique: RunContainer derives the volume name and the hostname from it.
func probeName(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("pentagi-probe-%d", rand.Uint64())
}

// probeConfig is fresh on every call because RunContainer mutates the config it is handed.
func probeConfig() *container.Config {
	return &container.Config{
		Image:      probeImage,
		Entrypoint: []string{"tail", "-f", "/dev/null"},
	}
}

// runProbeContainer returns the row RunContainer produced and the daemon's view of the container.
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

// cleanupProbeContainer also removes the named work volume, which RemoveVolumes leaves behind.
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

// requireMountSource matches by suffix because Docker Desktop rewrites bind sources into its VM namespace.
func requireMountSource(t *testing.T, mountPoint container.MountPoint, hostPath string) {
	t.Helper()

	require.Truef(t, strings.HasSuffix(mountPoint.Source, hostPath),
		"mount at %q has source %q, which does not resolve to host path %q",
		mountPoint.Destination, mountPoint.Source, hostPath)
}

// reserveFreePortBase picks a ports base whose derived flow ports are all free right now.
func reserveFreePortBase(t *testing.T, flowID int64) int {
	t.Helper()

	for range 50 {
		// below the 63535 ceiling GetPrimaryContainerPorts enforces, so the requested base is used
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

// dockerProbeNetwork creates a uniquely named network through ensureDockerNetwork and removes it afterwards.
func dockerProbeNetwork(t *testing.T, dc *dockerClient) string {
	t.Helper()

	name := probeName(t)
	ctx := context.WithoutCancel(t.Context())
	t.Cleanup(func() {
		if _, err := dc.client.NetworkRemove(ctx, name, client.NetworkRemoveOptions{}); err != nil &&
			!cerrdefs.IsNotFound(err) {
			t.Errorf("cleanup: failed to remove network %q: %v", name, err)
		}
	})
	require.NoError(t, ensureDockerNetwork(t.Context(), dc.client, name))

	return name
}

// Hardening applied unconditionally, the row lifecycle, and no daemon socket without DOCKER_INSIDE.
func TestClient_RunContainer_AppliesSandboxDefaults(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	dc.socket = filepath.Join(t.TempDir(), "docker.sock")
	name := probeName(t)

	row, inspect := runProbeContainer(t, dc, name, 7, nil, nil)

	require.Equal(t, fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(name))), inspect.Config.Hostname)
	require.Equal(t, "/work", inspect.Config.WorkingDir)
	require.Equal(t, t.Name(), inspect.Config.Labels["pentagi.test"])

	require.Equal(t, container.RestartPolicyOnFailure, inspect.HostConfig.RestartPolicy.Name)
	require.Equal(t, 5, inspect.HostConfig.RestartPolicy.MaximumRetryCount)
	require.NotNil(t, inspect.HostConfig.PidsLimit)
	require.Equal(t, int64(2048), *inspect.HostConfig.PidsLimit)
	require.Equal(t, "json-file", inspect.HostConfig.LogConfig.Type)
	require.Equal(t, "10m", inspect.HostConfig.LogConfig.Config["max-size"])
	require.Equal(t, "5", inspect.HostConfig.LogConfig.Config["max-file"])

	require.Len(t, inspect.HostConfig.Binds, 1)
	require.Contains(t, inspect.HostConfig.Binds[0], ":/work")
	for _, mountPoint := range inspect.Mounts {
		require.NotEqual(t, "/var/run/docker.sock", mountPoint.Destination)
	}

	require.Equal(t, database.ContainerStatusStarting, recorder.created.Status)
	require.Equal(t, database.ContainerTypePrimary, recorder.created.Type)
	require.Equal(t, []database.ContainerStatus{database.ContainerStatusRunning}, recorder.statuses)
	require.Equal(t, inspect.ID, row.LocalID.String)
	require.Equal(t, database.ContainerStatusRunning, row.Status)

	running, err := dc.IsContainerRunning(t.Context(), row.LocalID.String)
	require.NoError(t, err)
	require.True(t, running)
}

func TestClient_RunContainer_KeepsTheCallerPidsLimit(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	callerLimit := int64(64)

	_, inspect := runProbeContainer(t, dc, probeName(t), 8, nil, &container.HostConfig{
		Resources: container.Resources{PidsLimit: &callerLimit},
	})

	require.NotNil(t, inspect.HostConfig.PidsLimit)
	require.Equal(t, callerLimit, *inspect.HostConfig.PidsLimit)
}

// Without a host-side data directory /work is a named, labelled volume and the row records no host path.
func TestClient_RunContainer_BacksWorkDirWithAVolume(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	dc.hostDir = ""
	name := probeName(t)

	_, inspect := runProbeContainer(t, dc, name, 3, nil, nil)

	workMount := mountAt(t, inspect, "/work")
	require.Equal(t, mount.TypeVolume, workMount.Type)
	require.Equal(t, name+"-data", workMount.Name)
	require.True(t, workMount.RW)
	require.Empty(t, recorder.created.LocalDir.String)

	volume, err := dc.client.VolumeInspect(t.Context(), name+"-data", client.VolumeInspectOptions{})
	require.NoError(t, err)
	require.Equal(t, "local", volume.Volume.Driver)
	require.Equal(t, t.Name(), volume.Volume.Labels["pentagi.test"])

	listing, err := dc.ListContainerDir(t.Context(), inspect.ID, "/work")
	require.NoError(t, err)
	require.Empty(t, listing.Files)
}

// With a host-side data directory /work binds the per-flow subdirectory, created on the host and recorded on the row.
func TestClient_RunContainer_BindsThePerFlowHostDir(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	// the daemon runs on this machine, so the host path equals the local path
	dc.hostDir = dc.dataDir
	flowDir := filepath.Join(dc.dataDir, "flow-11")

	_, inspect := runProbeContainer(t, dc, probeName(t), 11, nil, nil)

	require.DirExists(t, flowDir)
	require.Equal(t, flowDir, recorder.created.LocalDir.String)

	workMount := mountAt(t, inspect, "/work")
	require.Equal(t, mount.TypeBind, workMount.Type)
	requireMountSource(t, workMount, flowDir)
	require.True(t, workMount.RW)

	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "marker.txt"), []byte("payload"), 0o600))
	listing, err := dc.ListContainerDir(t.Context(), inspect.ID, "/work")
	require.NoError(t, err)
	require.Len(t, listing.Files, 1, "a file written on the host must be listed inside the sandbox")
	require.Equal(t, "marker.txt", listing.Files[0].Name)
	require.Equal(t, int64(len("payload")), listing.Files[0].Size)
}

// In bridge mode the sandbox joins the configured network and publishes the flow's ports on the public IP.
func TestClient_RunContainer_JoinsTheNetworkAndPublishesFlowPorts(t *testing.T) {
	dc, _ := newRunContainerClient(t)
	dc.network = dockerProbeNetwork(t, dc)
	dc.publicIP = "127.0.0.1"
	flowID := int64(4)
	dc.portsBase = reserveFreePortBase(t, flowID)

	config := probeConfig()
	_, inspect := runProbeContainer(t, dc, probeName(t), flowID, config, nil)

	require.Contains(t, inspect.NetworkSettings.Networks, dc.network)
	require.NotEmpty(t, inspect.NetworkSettings.Networks[dc.network].IPAddress)

	require.Len(t, config.ExposedPorts, 2)
	for _, port := range []int{dc.portsBase + 8, dc.portsBase + 9} {
		containerPort, ok := network.PortFrom(uint16(port), network.TCP)
		require.True(t, ok)

		require.Contains(t, config.ExposedPorts, containerPort)

		published := inspect.NetworkSettings.Ports[containerPort]
		require.Len(t, published, 1)
		require.Equal(t, strconv.Itoa(port), published[0].HostPort)
		require.Equal(t, "127.0.0.1", published[0].HostIP.String())
	}
}

func TestClient_EnsureDockerNetwork_CreatesAMissingNetworkOnce(t *testing.T) {
	dc := newDaemonClient(t)
	name := dockerProbeNetwork(t, dc)

	created, err := dc.client.NetworkInspect(t.Context(), name, client.NetworkInspectOptions{})
	require.NoError(t, err)
	require.Equal(t, "bridge", created.Network.Driver)

	require.NoError(t, ensureDockerNetwork(t.Context(), dc.client, name),
		"a second call must be a no-op rather than a duplicate-network failure")
}

// An unparseable public IP is rejected before anything reaches the daemon.
func TestClient_RunContainer_RejectsAnInvalidPublicIP(t *testing.T) {
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

// A missing config is refused before a database row is inserted.
func TestClient_RunContainer_RejectsAMissingConfig(t *testing.T) {
	dc, recorder := newRunContainerClient(t)

	_, err := dc.RunContainer(t.Context(), probeName(t), database.ContainerTypePrimary, 1, nil, nil)

	require.ErrorContains(t, err, "no config found")
	require.Empty(t, recorder.created.Name)
	require.Empty(t, recorder.statuses)
}

func TestClient_RunContainer_MarksTheRowFailedWhenNoImageCanBePulled(t *testing.T) {
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

// A container created but not started is recorded as failed with its local id, and nothing is left behind.
func TestClient_RunContainer_MarksTheRowFailedWhenTheStartFails(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	dc.publicIP = "127.0.0.1"
	flowID := int64(21)
	dc.portsBase = reserveFreePortBase(t, flowID)

	// a held host port lets the create through and refuses the start with "port is already allocated"
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

// The daemon acknowledges a start whose entrypoint dies at once, so RunContainer must report it.
func TestClient_RunContainer_RejectsAContainerThatDoesNotStayRunning(t *testing.T) {
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
			require.Contains(t, startupErr.LogTail, "startup-diagnostic")

			require.Equal(t, []database.ContainerStatus{database.ContainerStatusFailed}, recorder.statuses)

			_, inspectErr := dc.client.ContainerInspect(t.Context(), name, client.ContainerInspectOptions{})
			require.True(t, cerrdefs.IsNotFound(inspectErr),
				"dead container must not be left behind, got: %v", inspectErr)
		})
	}
}

// A stopped leftover holds the name, so finding it takes a listing of every container, not just live ones.
func TestClient_RunContainer_ReplacesTheContainerHoldingTheName(t *testing.T) {
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

// Host network mode skips port publishing even when a public IP and ports base are set.
func TestClient_RunContainer_HostNetworkSkipsPortPublishing(t *testing.T) {
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
	require.Equal(t, "/work", mountAt(t, inspect, "/work").Destination)
}

// An image that cannot be pulled falls back to the default in the config, the row and the daemon.
func TestClient_RunContainer_FallsBackToTheDefaultImage(t *testing.T) {
	dc, recorder := newRunContainerClient(t)
	config := probeConfig()
	config.Image = probeName(t) + ":0" // a repository that cannot exist

	row, inspect := runProbeContainer(t, dc, probeName(t), 9, config, nil)

	require.Equal(t, probeImage, config.Image)
	require.Equal(t, []string{probeImage}, recorder.images)
	require.Equal(t, probeImage, row.Image)
	require.Equal(t, probeImage, inspect.Config.Image)
}

// With DOCKER_INSIDE the sandbox gets the socket, the client environment and the TLS material read-only.
func TestClient_RunContainer_WiresDockerInsideAccess(t *testing.T) {
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

	socketMount := mountAt(t, inspect, "/var/run/docker.sock")
	requireMountSource(t, socketMount, socketPath)

	certMount := mountAt(t, inspect, certPath)
	requireMountSource(t, certMount, certPath)
	require.False(t, certMount.RW)

	require.Subset(t, inspect.Config.Env, dc.insideEnv)
}

func TestClient_SweepFlowCommands_RunsTheTermThenTheKillPass(t *testing.T) {
	failure := errors.New("the sweep could not run")

	tests := map[string]struct {
		termOutput  string
		failing     string
		wantSignals []string
		wantGrace   bool
	}{
		"a TERM pass that signalled something is given a grace period": {
			termOutput: "hit\n", wantSignals: []string{"TERM", "KILL"}, wantGrace: true,
		},
		"a TERM pass that signalled nothing goes straight to KILL": {wantSignals: []string{"TERM", "KILL"}},
		"a TERM pass that fails stops the sweep":                   {failing: "TERM", wantSignals: []string{"TERM"}},
		"a KILL pass that fails is reported":                       {failing: "KILL", wantSignals: []string{"TERM", "KILL"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var signals []string
			started := time.Now()

			err := sweepFlowCommands(func(signal string) (string, error) {
				signals = append(signals, signal)
				if signal == tc.failing {
					return "", failure
				}
				return tc.termOutput, nil
			})

			if tc.failing != "" {
				require.ErrorIs(t, err, failure)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantSignals, signals)
			require.Equal(t, tc.wantGrace, time.Since(started) >= time.Second,
				"the grace period belongs between a signalled TERM pass and the KILL pass")
		})
	}
}

func dockerWaitForCommand(t *testing.T, dc *dockerClient, containerID, listed string) {
	t.Helper()

	require.Eventually(t, func() bool {
		commands, ok := liveCommands(t.Context(), dc, containerID)
		return ok && strings.Contains(commands, listed)
	}, 15*time.Second, 100*time.Millisecond, "%s never started", listed)
}

func TestClient_KillFlowCommands_SkipsASandboxThatIsNotRunning(t *testing.T) {
	dc := newDaemonClient(t)

	tests := map[string]func(ctx context.Context, containerID string) error{
		"a stopped sandbox": func(ctx context.Context, containerID string) error {
			_, err := dc.client.ContainerStop(ctx, containerID, client.ContainerStopOptions{})
			return err
		},
		"a removed sandbox": func(ctx context.Context, containerID string) error {
			_, err := dc.client.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{Force: true})
			return err
		},
	}

	for name, stop := range tests {
		t.Run(name, func(t *testing.T) {
			containerID := startProbeSandbox(t, dc, probeImage)
			require.NoError(t, stop(t.Context(), containerID))

			require.NoError(t, dc.KillFlowCommands(t.Context(), containerID))
		})
	}
}

func TestClient_KillFlowCommands_FinishesTheSweepAfterTheCallerGivesUp(t *testing.T) {
	dc := newDaemonClient(t)

	tests := map[string]time.Duration{
		"the caller gave up before the stop":     0,
		"the caller gives up between the passes": 300 * time.Millisecond,
	}

	for name, giveUpAfter := range tests {
		t.Run(name, func(t *testing.T) {
			containerID := startProbeSandbox(t, dc, probeImage)
			startInSandbox(t, dc, containerID, FlowCommand(`trap "" TERM; while :; do sleep 1; done`), false)
			dockerWaitForCommand(t, dc, containerID, "do sleep 1")

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if giveUpAfter == 0 {
				cancel()
			} else {
				time.AfterFunc(giveUpAfter, cancel)
			}
			require.NoError(t, dc.KillFlowCommands(ctx, containerID))

			require.NotContains(t, readLiveCommands(t, dc, containerID), "do sleep 1",
				"a command that ignores TERM outlived a stop whose caller gave up")
		})
	}
}

func TestClient_KillFlowCommands_SweepsAnUnhealthySandbox(t *testing.T) {
	dc := newDaemonClient(t)
	containerID := startSandbox(t, dc, &container.Config{
		Image: probeImage,
		Healthcheck: &container.HealthConfig{
			Test:     []string{"CMD-SHELL", "exit 1"},
			Interval: 100 * time.Millisecond,
			Timeout:  time.Second,
			Retries:  1,
		},
	}, &container.HostConfig{})

	startInSandbox(t, dc, containerID, FlowCommand(`sleep 961`), false)

	var healthErr error
	require.Eventually(t, func() bool {
		var operational bool
		operational, healthErr = dc.IsContainerRunning(t.Context(), containerID)
		if healthErr != nil {
			return true
		}

		commands, ok := liveCommands(t.Context(), dc, containerID)

		return !operational && ok && strings.Contains(commands, "sleep961")
	}, 15*time.Second, 200*time.Millisecond, "the sandbox never turned unhealthy with the flow command running")
	require.NoError(t, healthErr)

	require.NoError(t, dc.KillFlowCommands(t.Context(), containerID))

	require.NotContains(t, readLiveCommands(t, dc, containerID), "sleep961",
		"a flow command outlived a stop because its sandbox was unhealthy")
}

func TestClient_KillFlowCommands_RunsTheShellFoundOnThePath(t *testing.T) {
	dc := newDaemonClient(t)
	containerID := startProbeSandbox(t, dc, probeImage)

	runInSandbox(t, dc, containerID, `mv /bin/sh /usr/local/bin/sh`)
	startInSandbox(t, dc, containerID, FlowCommand(`sleep 921`), false)
	dockerWaitForCommand(t, dc, containerID, "sleep921")

	require.NoError(t, dc.KillFlowCommands(t.Context(), containerID))
	require.NotContains(t, readLiveCommands(t, dc, containerID), "sleep921",
		"the flow command outlived a stop in a sandbox whose shell is not at /bin/sh")
}

func TestClient_KillFlowCommands_ReportsASweepThatCannotRun(t *testing.T) {
	dc := newDaemonClient(t)
	containerID := startProbeSandbox(t, dc, probeImage)

	runInSandbox(t, dc, containerID, `rm /bin/sh`)

	require.ErrorContains(t, dc.KillFlowCommands(t.Context(), containerID),
		"the TERM sweep for '"+containerID+"' exited with code")
}
