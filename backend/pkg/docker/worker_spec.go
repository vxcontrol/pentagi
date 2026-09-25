package docker

import (
	"pentagi/pkg/config"

	"github.com/moby/moby/api/types/container"
)

// WorkerSpec is the container a sandbox worker is made from, before
// RunContainer adds what every container of this daemon gets: the injected
// DOCKER_* environment, the certificate mount, the network, the pids limit.
//
// It lives here rather than beside its caller in pkg/tools because the startup
// sandbox check has to launch a worker that is identical to a flow's in every
// way but the image. Two copies of this would agree on the day they were
// written and drift afterwards, and the check would then be measuring a
// container no agent ever runs in.
func WorkerSpec(cfg *config.Config, image string) (*container.Config, *container.HostConfig) {
	return &container.Config{
			Image:      image,
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		}, &container.HostConfig{
			CapDrop: []string{"ALL"},
			CapAdd:  workerCapabilities(cfg),
		}
}

// workerCapabilities is the explicit allow-list that goes with CapDrop: ALL --
// Docker's default 14 minus MKNOD (block-device escape vector), plus SYS_PTRACE
// (debugging, not a Docker default) and NET_ADMIN when configured. Never add
// SYS_ADMIN, SYS_MODULE, SYS_RAWIO, SYS_BOOT. See "Capability Management" in
// docker.md for the full per-capability rationale.
func workerCapabilities(cfg *config.Config) []string {
	capAdd := []string{
		"CHOWN", "DAC_OVERRIDE", "FSETID", "FOWNER",
		"NET_RAW", "SETGID", "SETUID", "SETFCAP", "SETPCAP",
		"NET_BIND_SERVICE", "SYS_CHROOT", "KILL", "AUDIT_WRITE", "SYS_PTRACE",
	}
	if cfg != nil && cfg.DockerNetAdmin {
		capAdd = append(capAdd, "NET_ADMIN")
	}

	return capAdd
}
