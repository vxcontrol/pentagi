package docker

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFlowCommands_FlowCommand_SweepEndsItsCommandsInThisPidNamespace(t *testing.T) {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		t.Skip("the sweep signals every marked process in the pid namespace; it runs only inside a container")
	}

	start := func(cmd []string) <-chan struct{} {
		c := exec.Command(cmd[0], cmd[1:]...)
		require.NoError(t, c.Start())

		done := make(chan struct{})
		go func() { _ = c.Wait(); close(done) }()
		t.Cleanup(func() {
			_ = c.Process.Kill()
			<-done
		})

		return done
	}

	flow := start(FlowCommand(`sleep 600`))
	stubborn := start(FlowCommand(`trap "" TERM; sleep 601 & wait`))
	other := start([]string{"sh", "-c", "sleep 602"})

	sweep := func(signal string) (string, error) {
		out, err := exec.Command("sh", "-c", killFlowCommandsScript, "sh", signal).Output()

		return strings.TrimSpace(string(out)), err
	}

	var termErr error
	require.Eventually(t, func() bool {
		var out string
		out, termErr = sweep("TERM")

		return termErr != nil || out == "hit"
	}, 5*time.Second, 50*time.Millisecond, "the TERM pass never found the flow's commands")
	require.NoError(t, termErr)

	time.Sleep(flowCommandsGracePeriod)
	_, err := sweep("KILL")
	require.NoError(t, err)

	for name, done := range map[string]<-chan struct{}{
		"the flow's command":          flow,
		"a command that ignores TERM": stubborn,
	} {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s outlived the sweep", name)
		}
	}
	require.False(t, runs("sleep 601"), "the child that inherited the mark outlived the sweep")

	select {
	case <-other:
		t.Fatal("a command outside the flow was ended by the sweep")
	case <-time.After(500 * time.Millisecond):
	}
	require.True(t, runs("sleep 602"), "a command outside the flow was ended by the sweep")
}

func runs(command string) bool {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		line, err := os.ReadFile("/proc/" + entry.Name() + "/cmdline")
		if err != nil {
			continue
		}
		if strings.Join(strings.Split(strings.TrimRight(string(line), "\x00"), "\x00"), " ") == command {
			return true
		}
	}

	return false
}
