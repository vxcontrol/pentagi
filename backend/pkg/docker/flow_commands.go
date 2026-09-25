package docker

import "time"

const (
	flowCommandsGracePeriod = time.Second
	flowCommandsSweepLimit  = 30 * time.Second

	ignoreHangupSh   = `trap "" HUP`
	sandboxCommandSh = ignoreHangupSh + `; exec sh -c "$0"`

	markFlowCommandSh = ignoreHangupSh + `; read -r adj 2>/dev/null </proc/self/oom_score_adj && ` +
		`echo $((adj + 1)) 2>/dev/null >/proc/self/oom_score_adj; exec sh -c "$0"`
)

// FlowCommand builds a flow task's sandbox command; Stop ends only commands launched through it.
func FlowCommand(command string) []string {
	return []string{"sh", "-c", markFlowCommandSh, command}
}

// SandboxCommand builds a sandbox command that belongs to no flow task; Stop leaves it running.
func SandboxCommand(command string) []string {
	return []string{"sh", "-c", sandboxCommandSh, command}
}

const killFlowCommandsScript = `sig=$1
read -r base </proc/self/oom_score_adj || exit 0
mark=$((base + 1))

sweep() {
  hits=0
  for p in /proc/[0-9]*; do
    read -r adj 2>/dev/null <"$p/oom_score_adj" || continue
    [ "$adj" = "$mark" ] || continue
    read -r state 2>/dev/null <"$p/stat" || continue
    kill -"$sig" "${p#/proc/}" 2>/dev/null || continue
    case "${state##*") "}" in Z*) ;; *) hits=$((hits + 1)) ;; esac
  done
}

sweep
if [ "$sig" = TERM ]; then
  [ "$hits" -gt 0 ] && echo hit
  exit 0
fi

passes=1
while [ "$hits" -gt 0 ] && [ "$passes" -lt 10 ]; do
  sweep
  passes=$((passes + 1))
done
exit 0`
