package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

// listingDone terminates a listing written from inside the container, so a partial file is not read as one.
const listingDone = "LISTING-COMPLETE"

// listCommandsToFile writes the listing with shell builtins only, so it costs no pid at the pids limit.
func listCommandsToFile(path string) string {
	return "{ " + listCommandsScript + "\necho " + listingDone + "; } >" + path
}

// sandboxListing reads a listCommandsToFile listing through the copy API; it is "" until complete.
func sandboxListing(ctx context.Context, dc *dockerClient, containerID, path string) string {
	out := sandboxFile(ctx, dc, containerID, path)
	body, done := strings.CutSuffix(strings.TrimSpace(out), listingDone)
	if !done {
		return ""
	}

	return body
}

func putSandboxFile(t *testing.T, dc *dockerClient, containerID, name, content string) {
	t.Helper()

	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}))
	_, err := tw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tw.Close())

	require.NoError(t, dc.CopyToContainer(t.Context(), containerID, "/tmp", &archive, client.CopyToContainerOptions{}))
}

func sandboxFile(ctx context.Context, dc *dockerClient, containerID, path string) string {
	reader, _, err := dc.CopyFromContainer(ctx, containerID, path)
	if err != nil {
		return ""
	}
	defer reader.Close()

	tr := tar.NewReader(reader)
	if _, err := tr.Next(); err != nil {
		return ""
	}
	content, _ := io.ReadAll(tr)

	return string(content)
}

func TestFlowCommands_FlowCommand_StopEndsItsCommandsAndNothingElse(t *testing.T) {
	dc := newDaemonClient(t)

	type launch struct {
		command string
		tty     bool
		listed  string
	}
	flowCommands := map[string]launch{
		"a foreground command":                       {`sleep 901`, true, "sleep901"},
		"a command that ignores TERM":                {`trap "" TERM; while :; do sleep 1; done`, true, "do sleep 1"},
		"a child that inherited an ignored TERM":     {`trap "" TERM; sleep 910 & wait`, false, "sleep910"},
		"an orphan left by nohup":                    {`nohup sleep 903 >/dev/null 2>&1 &`, false, "sleep903"},
		"an orphan in a session of its own":          {`setsid sleep 904 >/dev/null 2>&1 &`, false, "sleep904"},
		"a command that clears its environment":      {`env -i sleep 905`, false, "sleep905"},
		"an orphan that cleared its environment too": {`env -i setsid sleep 909 >/dev/null 2>&1 &`, false, "sleep909"},
		"a foreground command that clears its env":   {`env -i sleep 912`, true, "sleep912"},
	}
	others := map[string]launch{
		"the assistant's command":            {`sleep 906`, true, "sleep906"},
		"the assistant's orphan":             {`nohup sleep 907 >/dev/null 2>&1 &`, false, "sleep907"},
		"the assistant's session of its own": {`setsid sleep 911 >/dev/null 2>&1 &`, false, "sleep911"},
	}
	images := map[string]string{
		"a busybox sandbox":         probeImage,
		"the default sandbox image": defaultImage,
	}

	for name, image := range images {
		t.Run(name, func(t *testing.T) {
			containerID := startProbeSandbox(t, dc, image)

			for _, flow := range flowCommands {
				startInSandbox(t, dc, containerID, FlowCommand(flow.command), flow.tty)
			}
			for _, other := range others {
				startInSandbox(t, dc, containerID, []string{"sh", "-c", other.command}, other.tty)
			}
			startInSandbox(t, dc, containerID, []string{"sleep", "908"}, false)
			runInSandbox(t, dc, containerID, `mkdir -p /work && echo kept >/work/keep.txt`)

			require.Eventually(t, func() bool {
				commands, ok := liveCommands(t.Context(), dc, containerID)
				if !ok {
					return false
				}
				for _, l := range flowCommands {
					if !strings.Contains(commands, l.listed) {
						return false
					}
				}
				for _, l := range others {
					if !strings.Contains(commands, l.listed) {
						return false
					}
				}

				return strings.Contains(commands, "sleep908")
			}, 15*time.Second, 200*time.Millisecond, "the probe commands never started")

			require.NoError(t, dc.KillFlowCommands(t.Context(), containerID))

			running, err := dc.IsContainerRunning(t.Context(), containerID)
			require.NoError(t, err)
			require.True(t, running, "stopping a flow must not stop its sandbox")

			commands := readLiveCommands(t, dc, containerID)
			for name, flow := range flowCommands {
				require.NotContains(t, commands, flow.listed, "%s outlived the stop", name)
			}
			for name, other := range others {
				require.Contains(t, commands, other.listed, "%s did not survive the stop", name)
			}
			require.Contains(t, commands, "sleep908", "a service exec from REST did not survive the stop")
			require.Equal(t, "kept", strings.TrimSpace(runInSandbox(t, dc, containerID, `cat /work/keep.txt`)))

			// the ended orphans linger as marked zombies under PID 1
			require.Empty(t, strings.TrimSpace(runInSandbox(t, dc, containerID, `set -- TERM; `+killFlowCommandsScript)),
				"a flow command already ended counted as one still to end")
		})
	}
}

func TestFlowCommands_WrapperAndSweepRunAtThePidsLimit(t *testing.T) {
	const pidsLimit = 32

	dc := newDaemonClient(t)
	containerID := startProbeSandbox(t, dc, probeImage)

	putSandboxFile(t, dc, containerID, "sweep.sh", killFlowCommandsScript)
	startInSandbox(t, dc, containerID, []string{"sh", "-c", "sleep 906"}, true)
	startInSandbox(t, dc, containerID,
		FlowCommand(`trap "" TERM; i=0; while [ $i -lt 8 ]; do sleep 600 & i=$((i+1)); done; wait`), false)

	wrapper := FlowCommand(`echo ran >/tmp/ran-at-limit; trap "" TERM; exec sleep 999`)
	occupancy := `read -r cur </sys/fs/cgroup/pids.current; read -r max </sys/fs/cgroup/pids.max`
	runners := map[string]string{
		"go-wrapper": `exec sh -c "$1" "$2"`,
		// an EXIT trap outlives the sourced sweep's `exit 0` without forking a subshell the limit would deny
		"go-kill": `trap '` + listCommandsToFile("/tmp/after-kill") + `' EXIT
` + occupancy + `; echo "$cur/$max" >/tmp/at-kill; set -- KILL; . /tmp/sweep.sh`,
		"go-fill": `read -r max </sys/fs/cgroup/pids.max; read -r cur </sys/fs/cgroup/pids.current
while [ "$cur" -lt "$max" ]; do sleep 700 & read -r cur </sys/fs/cgroup/pids.current; done
echo filled >/tmp/filled; exec tail -f /dev/null /tmp/filled`,
	}
	for trigger, body := range runners {
		startInSandbox(t, dc, containerID, []string{"sh", "-c",
			`while [ ! -e /tmp/` + trigger + ` ]; do :; done; ` + body, "sh", wrapper[2], wrapper[3]}, false)
	}
	startInSandbox(t, dc, containerID, []string{"sh", "-c",
		`( while [ ! -e /tmp/go-term ]; do :; done; ` + occupancy + `; echo "$cur/$max" >/tmp/at-term
set -- TERM; . /tmp/sweep.sh ) >/tmp/sweep-term.out
sleep 700 &
echo held >/tmp/term-held
exec tail -f /dev/null`}, false)

	require.Eventually(t, func() bool {
		commands, ok := liveCommands(t.Context(), dc, containerID)
		if !ok {
			return false
		}
		return strings.Count(commands, "sleep600") == 8 && strings.Contains(commands, "sleep906") &&
			strings.Count(commands, "while [ ! -e /tmp/go-") == len(runners)+2
	}, 15*time.Second, 100*time.Millisecond, "the probe commands never started")

	// Set only now: each exec's runc init counts its threads against pids.max while it starts.
	limit := int64(pidsLimit)
	_, err := dc.client.ContainerUpdate(t.Context(), containerID, client.ContainerUpdateOptions{
		Resources: &container.Resources{PidsLimit: &limit},
	})
	require.NoError(t, err)

	fire := func(trigger string, done func() bool, failure string) {
		putSandboxFile(t, dc, containerID, trigger, "")
		require.Eventually(t, done, 10*time.Second, 100*time.Millisecond, failure)
	}
	sandboxLine := func(path string) string {
		return strings.TrimSpace(sandboxFile(t.Context(), dc, containerID, path))
	}

	fire("go-fill", func() bool {
		return sandboxLine("/tmp/filled") == "filled"
	}, "the sandbox was never filled to its pids limit")
	fire("go-wrapper", func() bool {
		return sandboxLine("/tmp/ran-at-limit") == "ran"
	}, "a flow command did not run at the pids limit")
	fire("go-term", func() bool {
		return sandboxLine("/tmp/sweep-term.out") == "hit" && sandboxLine("/tmp/term-held") == "held"
	}, "the TERM pass did not finish at the pids limit")

	var commands string
	fire("go-kill", func() bool {
		commands = sandboxListing(t.Context(), dc, containerID, "/tmp/after-kill")

		return commands != ""
	}, "the KILL pass did not finish at the pids limit")

	atLimit := fmt.Sprintf("%d/%d", pidsLimit, pidsLimit)
	require.Equal(t, atLimit, sandboxLine("/tmp/at-term"), "the TERM pass ran below the pids limit")
	require.Equal(t, atLimit, sandboxLine("/tmp/at-kill"), "the KILL pass ran below the pids limit")
	require.NotContains(t, commands, "sleep600", "a flow command outlived the KILL pass at the pids limit")
	require.NotContains(t, commands, "sleep999", "the flow wrapper outlived the KILL pass at the pids limit")
	require.Contains(t, commands, "sleep906", "the assistant's command did not survive the sweep")
	require.Contains(t, commands, "sleep700", "a command outside the flow did not survive the sweep")
}

// Subtests are keyed by unit: FlowCommand and SandboxCommand.
func TestFlowCommands_BackgroundCommandsOutliveTheTerminalThatStartedThem(t *testing.T) {
	dc := newDaemonClient(t)

	tests := map[string]struct {
		launch  func(string) []string
		command string
		listed  string
	}{
		"a flow task's command":    {FlowCommand, "sleep 981", "sleep981"},
		"a command outside a flow": {SandboxCommand, "sleep 982", "sleep982"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			containerID := startProbeSandbox(t, dc, probeImage)

			select {
			case err := <-startInSandbox(t, dc, containerID, tc.launch(tc.command+` >/dev/null 2>&1 &`), true):
				require.NoError(t, err, "the terminal's stream broke instead of closing")
			case <-time.After(15 * time.Second):
				t.Fatal("the terminal never closed after its command exited")
			}

			require.Never(t, func() bool {
				commands, ok := liveCommands(t.Context(), dc, containerID)

				return ok && !strings.Contains(commands, tc.listed)
			}, 3*time.Second, 200*time.Millisecond,
				"the pty hung up when the command exited and took its background process with it")
		})
	}
}
