# =============================================================================
# Docker authorization policy for the DinD daemon (opa-docker-authz).
#
# Goal: nested containers launched THROUGH DinD must not be able to escape to
# the DinD sandbox layer or reach the physical host. Three classes of
# primitives are denied:
#
#   1. Capability / device / namespace escalation: obtaining SYS_ADMIN or raw
#      block-device access enables mounting the disk and reading the host FS.
#
#   2. Host bind-mount escalation: mounting ANY host path (including writable
#      directories like /work) allows filesystem pivots, symlink escapes, and
#      credential leakage. ALL host bind-mounts are denied; use Docker volumes.
#
#   3. Host namespace sharing: --pid=host, --ipc=host, etc. expose the DinD
#      host's process tree and IPC namespace, enabling credential extraction.
#
# IMPORTANT: this is a deny-list. It is only as strong as its completeness.
# Any new Docker API field that can grant SYS_ADMIN / device access / namespace
# escape or filesystem access must be added here. For a kernel-enforced
# guarantee, use dind-rootless.
#
# The opa-docker-authz plugin evaluates data.docker.authz.allow (boolean).
#
# Controlled request types:
#   - containers/create  (POST)     — checked for capability/mount/namespace violations
#   - containers/{id}/exec (POST)   — checked for privileged flag
#   - containers/{id}/update (POST) — checked for DeviceCgroupRules injection
#   - containers/{id}/archive (GET) — unconditionally denied (PKI extraction risk)
#   - commit (POST)                 — unconditionally denied (state snapshot risk)
#   - build  (POST)                 — unconditionally denied (supply chain injection risk, legacy API)
#   - session (POST, BuildKit)      — unconditionally denied (bypasses /build in Docker 23+)
#   - images/load (POST)            — unconditionally denied (arbitrary tar with device nodes / setuid)
#   - plugins/* (POST/PUT/DELETE)   — unconditionally denied (arbitrary plugin = OPA bypass)
#   - swarm/services/nodes/configs/secrets (POST/PUT/DELETE) — unconditionally denied
#                                     (Swarm creates containers bypassing HostConfig validation)
#   - volumes/create (POST)         — checked for bind-emulation driver options
#   - networks/create (POST)        — checked for L2 drivers (macvlan/ipvlan)
# All other Docker API requests are allowed without further inspection.
# =============================================================================
package docker.authz

default allow = false

# Allow anything that is neither a container-create nor a container-exec request
# nor any other controlled request type.
allow {
    not is_container_create
    not is_container_exec
    not is_container_archive
    not is_container_commit
    not is_image_build
    not is_buildkit_session
    not is_container_update
    not is_image_load
    not is_plugin_mutate
    not is_swarm_mutate
    not is_volume_create
    not is_network_create
}

# Allow container creation only when it triggers no violation.
allow {
    is_container_create
    count(violation) == 0
}

# Allow exec only when it triggers no exec_violation.
allow {
    is_container_exec
    count(exec_violation) == 0
}

# Allow container update only when it triggers no update_violation.
allow {
    is_container_update
    count(update_violation) == 0
}

# Allow volume creation only when driver options are safe (no bind-emulation).
allow {
    is_volume_create
    count(volume_violation) == 0
}

# Allow network creation only when the driver is not an L2 passthrough type.
allow {
    is_network_create
    count(network_violation) == 0
}

# Archive, commit, and build are unconditionally denied.
# These endpoints bypass the containers/create gate and can be used to:
#   - archive: extract PKI material from containers with bind-mounts or
#              sensitive volumes (GET /containers/{id}/archive?path=/hostfs/certs)
#   - commit:  snapshot running container state including cached credentials
#   - build:   inject malicious image layers bypassing OPA create restrictions

# ---------------------------------------------------------------------------
# Request classification
# ---------------------------------------------------------------------------

is_container_create {
    input.Method == "POST"
    regex.match(`/containers/create`, input.Path)
}

is_container_exec {
    input.Method == "POST"
    regex.match(`/containers/[^/]+/exec`, input.Path)
}

# Container archive (GET) — reading container filesystems; gated to prevent
# extraction of sensitive paths such as /proc/1/root (host FS pivot),
# /certs, and any other path that could expose PKI or host data.
is_container_archive {
    input.Method == "GET"
    regex.match(`/containers/[^/]+/archive`, input.Path)
}

# Container commit — committing a running container to an image; gated to
# prevent capturing sensitive state and exfiltrating it as an image layer.
is_container_commit {
    input.Method == "POST"
    regex.match(`/commit`, input.Path)
}

# Image build — building a new image from a Dockerfile context; gated to
# prevent supply-chain attacks embedding malicious layers or escape tooling.
is_image_build {
    input.Method == "POST"
    regex.match(`/build`, input.Path)
}

# BuildKit session — Docker's default builder (DOCKER_BUILDKIT=1, default since
# Docker 23) bypasses the legacy POST /build endpoint and instead opens a
# WebSocket session via POST /session before issuing gRPC build commands.
# Denied alongside is_image_build to close the build surface completely.
is_buildkit_session {
    input.Method == "POST"
    regex.match(`/session`, input.Path)
}

# Container update — updating resource limits and runtime config of a running
# container. Most fields (memory, CPU) are safe; DeviceCgroupRules is not:
# injecting "b 8:* rwm" grants the container permission to open any block
# device, which bypasses the device-cgroup eBPF filter set at create-time.
# Combined with seccomp.json this is not immediately exploitable
# (mknod for S_IFBLK is blocked and /dev/ contains no block device nodes),
# but the gap should be closed at the policy layer for defense-in-depth.
is_container_update {
    input.Method == "POST"
    regex.match(`/containers/[^/]+/update`, input.Path)
}

# Image load — POST /images/load accepts an arbitrary tar archive and unpacks
# it into the daemon image store. Unlike registry pulls, tar archives are not
# signed or scanned and can contain block device nodes or setuid binaries.
# Pulling from registries (POST /images/create?fromImage=...) is NOT blocked.
is_image_load {
    input.Method == "POST"
    regex.match(`/images/load`, input.Path)
}

# Plugin management (mutating) — installing, enabling, disabling, or removing
# plugins grants the plugin process daemon-level privileges and can implement
# arbitrary volume, network, or authz logic. A malicious authz plugin would
# unconditionally approve all subsequent requests, bypassing this policy.
# Plugins are managed out-of-band via daemon.json; runtime management is off.
is_plugin_mutate {
    input.Method != "GET"
    regex.match(`/plugins`, input.Path)
}

# Swarm / Services (mutating) — Swarm service specs go through /services/create
# rather than /containers/create, and the daemon may schedule containers
# without applying the HostConfig validation enforced above. Blocking all
# Swarm mutating operations ensures every container goes through the guarded
# containers/create path.
is_swarm_mutate {
    input.Method != "GET"
    regex.match(`/swarm`, input.Path)
}

is_swarm_mutate {
    input.Method != "GET"
    regex.match(`/services`, input.Path)
}

is_swarm_mutate {
    input.Method != "GET"
    regex.match(`/nodes`, input.Path)
}

is_swarm_mutate {
    input.Method != "GET"
    regex.match(`/configs`, input.Path)
}

is_swarm_mutate {
    input.Method != "GET"
    regex.match(`/secrets`, input.Path)
}

# Volume creation — POST /volumes/create is safe for named volumes with the
# local driver and default options. However, local-driver DriverOpts of the
# form {type:none, o:bind, device:/host/path} emulate a bind-mount from the
# DinD container's own filesystem into nested containers, potentially exposing
# TLS certs or DinD daemon state. Checked via volume_violation below.
is_volume_create {
    input.Method == "POST"
    regex.match(`/volumes/create`, input.Path)
}

# Network creation — POST /networks/create is safe for bridge/null/overlay.
# macvlan and ipvlan attach containers directly to the host's L2 segment,
# bypassing iptables rules and allowing access to physical network interfaces.
# Checked via network_violation below.
is_network_create {
    input.Method == "POST"
    regex.match(`/networks/create`, input.Path)
}

# ---------------------------------------------------------------------------
# Capability whitelist: the set of capabilities a nested container may request
# via --cap-add. Mirrors Docker's default cap set with two intentional deltas:
#   - MKNOD excluded  (block device escape vector; still in Docker default set)
#   - NET_ADMIN added (NOT a Docker default; safe for pentest tooling, see below)
# Never add SYS_ADMIN / SYS_MODULE / SYS_PTRACE / SYS_RAWIO / SYS_BOOT.
# ---------------------------------------------------------------------------
allowed_caps = {
    "CHOWN",
    "DAC_OVERRIDE",
    "FSETID",
    "FOWNER",
    # MKNOD intentionally omitted from the --cap-add whitelist: CAP_MKNOD allows
    # creation of block device nodes (mknod /dev/sda b 8 1), which combined with
    # tools like debugfs enables raw ext4 filesystem reads WITHOUT SYS_ADMIN.
    # This was the primary confirmed escape vector in pentest.
    #
    # Dual enforcement against MKNOD block-device escape:
    #   1. authz blocks explicit "--cap-add MKNOD" here.
    #   2. daemon.json sets seccomp.json as the daemon-wide default seccomp profile.
    #      seccomp.json blocks mknod/mknodat with S_IFBLK at the kernel level.
    #      authz blocks "--security-opt" overrides (see below), so the seccomp
    #      profile cannot be replaced or disabled by nested containers.
    "NET_RAW",
    "SETGID",
    "SETUID",
    "SETFCAP",
    "SETPCAP",
    "NET_BIND_SERVICE",
    "SYS_CHROOT",
    "KILL",
    "AUDIT_WRITE",
    # NET_ADMIN is not in Docker's default set, but pentest tooling commonly
    # needs it (raw sockets, VPN, interface config). It is network-namespace
    # scoped and does NOT grant mount / CAP_SYS_ADMIN, so it is safe re: host
    # filesystem escape.
    "NET_ADMIN",
}

# Normalize a capability name: upper-case and strip an optional "CAP_" prefix.
normalize_cap(c) = n {
    u := upper(c)
    n := trim_prefix(u, "CAP_")
}

# ---------------------------------------------------------------------------
# Exec violations: privileged exec sessions grant CAP_SYS_ADMIN + CAP_NET_ADMIN
# to an already-running container, bypassing container-create checks. Blocked
# here to maintain policy consistency with the --privileged create restriction.
# ---------------------------------------------------------------------------

exec_violation["privileged exec sessions are not allowed"] {
    is_container_exec
    input.Body.Privileged == true
}

# ---------------------------------------------------------------------------
# Container-create violations (additional fields)
# ---------------------------------------------------------------------------

# Sysctls: with --network=host + CAP_NET_ADMIN, container-level sysctl writes
# go directly into the host network namespace (e.g. net.ipv4.ip_forward).
# Blocking all Sysctls prevents unintended host kernel parameter changes.
# Pentest tools (nmap, masscan) do not require custom sysctls.
violation["sysctls are not allowed"] {
    is_container_create
    count(input.Body.HostConfig.Sysctls) > 0
}

# DeviceRequests: GPU / accelerator access via device plugins. Not needed for
# pentest workloads; device plugins can expose DMA or sensitive hardware paths.
violation["device requests are not allowed"] {
    is_container_create
    count(input.Body.HostConfig.DeviceRequests) > 0
}

# Runtime: non-standard runtimes (e.g. runsc, crun) may have different
# seccomp defaults or capability handling. Enforce runc only.
violation[msg] {
    is_container_create
    rt := input.Body.HostConfig.Runtime
    rt != ""
    rt != "runc"
    msg := sprintf("runtime not allowed: %v", [rt])
}

# ---------------------------------------------------------------------------
# Update violations: DeviceCgroupRules injection bypasses the device-cgroup
# eBPF filter and grants the container access to block devices. While not
# immediately exploitable (mknod S_IFBLK is seccomp-blocked and /dev/ has no
# block nodes), it is a direct policy bypass and must be denied here.
# Safe update fields (Memory, NanoCpus, CpusetCpus, PidsLimit, RestartPolicy,
# BlkioWeight, etc.) are explicitly allowed when DeviceCgroupRules is absent.
# ---------------------------------------------------------------------------

update_violation["DeviceCgroupRules injection via update is not allowed"] {
    is_container_update
    count(input.Body.DeviceCgroupRules) > 0
}

# ---------------------------------------------------------------------------
# Volume violations: local-driver bind-emulation
# ---------------------------------------------------------------------------

# The local volume driver accepts DriverOpts {type:none, o:bind, device:/path}
# which makes Docker treat /path inside DinD as the volume backing store —
# effectively a bind-mount from DinD's own filesystem into nested containers.
# This could expose /etc/docker/dind/certs, /var/run/dind/docker.sock, etc.
volume_violation["bind-emulation via volume driver option 'o=bind' is not allowed"] {
    is_volume_create
    opt := input.Body.DriverOpts["o"]
    contains(lower(opt), "bind")
}

volume_violation["bind-emulation via volume driver option 'type=none' is not allowed"] {
    is_volume_create
    input.Body.DriverOpts["type"] == "none"
}

# ---------------------------------------------------------------------------
# Network violations: L2 driver restriction
# ---------------------------------------------------------------------------

# macvlan and ipvlan attach containers directly to the host's L2 network
# segment, bypassing Docker's iptables rules and allowing access to all
# hosts on the physical subnet. bridge, null, and overlay are safe.
_unsafe_net_drivers := {"macvlan", "ipvlan"}

network_violation[msg] {
    is_network_create
    drv := lower(input.Body.Driver)
    _unsafe_net_drivers[drv]
    msg := sprintf("network driver not allowed: %v", [input.Body.Driver])
}

# ---------------------------------------------------------------------------
# Violations
# ---------------------------------------------------------------------------

# --privileged
violation["privileged containers are not allowed"] {
    is_container_create
    input.Body.HostConfig.Privileged == true
}

# --cap-add of anything outside the safe whitelist (also catches "ALL")
violation[msg] {
    is_container_create
    some i
    cap := input.Body.HostConfig.CapAdd[i]
    not allowed_caps[normalize_cap(cap)]
    msg := sprintf("capability not allowed: %v", [cap])
}

# --device (block/char device passthrough)
violation["device passthrough is not allowed"] {
    is_container_create
    count(input.Body.HostConfig.Devices) > 0
}

# --device-cgroup-rule
violation["device cgroup rules are not allowed"] {
    is_container_create
    count(input.Body.HostConfig.DeviceCgroupRules) > 0
}

# --security-opt (seccomp/apparmor/systempaths=unconfined, etc.)
violation["custom security options are not allowed"] {
    is_container_create
    count(input.Body.HostConfig.SecurityOpt) > 0
}

# Host namespaces (pid/ipc/userns/cgroup/uts).
#
# NOTE on --network=host: host *network* is intentionally NOT denied here.
# Nested containers need it for DNS resolution and raw socket access from
# the DinD bridge network. However, it is a documented HIGH risk: DinD itself
# runs with --network=host, so nested containers with --network=host can reach
# every service on the physical host's network stack — including the outer
# Docker daemon's TLS port (3376). This was the primary Phase 3 escape vector.
# Mitigations:
#   - Docker-ca TLS client certificates must never be accessible inside DinD.
#   - Outer daemon must deploy its own authorization plugin.
#   - If network isolation is more important than raw-socket pentesting,
#     remove --network=host from this allowlist and configure explicit DNS.
violation["host PID namespace is not allowed"] {
    is_container_create
    input.Body.HostConfig.PidMode == "host"
}

# --pid=container:<id>: shares another container's PID namespace; blocked for
# defense-in-depth even though the kernel currently rejects it at mount namespace
# boundaries. Keeping the block here ensures policy consistency if the downstream
# check is ever relaxed or a different runtime is used.
violation["container PID namespace sharing is not allowed"] {
    is_container_create
    startswith(lower(input.Body.HostConfig.PidMode), "container:")
}

violation["host IPC namespace is not allowed"] {
    is_container_create
    input.Body.HostConfig.IpcMode == "host"
}

# --ipc=shareable: marks a container as a donor that other containers can join.
# Blocked to prevent IPC namespace sharing chains that bypass the host IPC denial.
violation["shareable IPC mode is not allowed"] {
    is_container_create
    input.Body.HostConfig.IpcMode == "shareable"
}

# --ipc=container:<id>: joins another container's IPC namespace.
violation["container IPC namespace sharing is not allowed"] {
    is_container_create
    startswith(lower(input.Body.HostConfig.IpcMode), "container:")
}

violation["host user namespace is not allowed"] {
    is_container_create
    input.Body.HostConfig.UsernsMode == "host"
}

violation["host cgroup namespace is not allowed"] {
    is_container_create
    input.Body.HostConfig.CgroupnsMode == "host"
}

violation["host UTS namespace is not allowed"] {
    is_container_create
    input.Body.HostConfig.UTSMode == "host"
}

# ---------------------------------------------------------------------------
# MaskedPaths and ReadonlyPaths overrides
#
# CONFIRMED ESCAPE VECTOR
#   Setting ReadonlyPaths:[] removes Docker's default /proc/sys read-only mount.
#   Since /proc/sys/kernel is NOT a PID namespace path — it is global to the
#   host kernel — writing /proc/sys/kernel/modprobe from inside a nested DinD
#   container changes the HOST kernel's usermode helper path. When any package
#   is installed with `apk add` (or equivalent), the busybox post-install
#   trigger calls request_module(), the host kernel executes modprobe_path,
#   and the payload runs as root in the init mount namespace = physical host.
#
#   Setting MaskedPaths:[] exposes /proc/kcore (raw kernel memory),
#   /proc/keys (kernel keyring), /sys/firmware/efi/efivars (UEFI variables —
#   writing here can permanently brick the system), and /proc/sched_debug.
#
# This is NOT a CVE in runc. It is legitimate Docker API behaviour: empty
# arrays [] explicitly replace the default lists, whereas an absent field
# (null / not present) keeps Docker's built-in safe defaults. The same
# mechanism is used by Kubernetes procMount:Unmasked (a privileged feature
# gated behind PodSecurity admission).
#
# NOTE: --security-opt systempaths=unconfined achieves the same effect as
# MaskedPaths:[]+ReadonlyPaths:[] and is already denied by the security-opt
# violation above. This adds API-level protection for direct JSON injection.
#
# Safe rule: absent field (null/undefined) = Docker applies built-in defaults.
#            is_array([]) or is_array([...]) = DENIED.
# ---------------------------------------------------------------------------

violation["explicit ReadonlyPaths override is not allowed — absent field uses Docker defaults; any explicit array (including []) removes /proc/sys read-only protection and enables modprobe_path host escape"] {
    is_container_create
    is_array(input.Body.HostConfig.ReadonlyPaths)
}

violation["explicit MaskedPaths override is not allowed — absent field uses Docker defaults; any explicit array (including []) unmasks /proc/kcore, /proc/keys, and /sys/firmware/efi/efivars"] {
    is_container_create
    is_array(input.Body.HostConfig.MaskedPaths)
}

# ---------------------------------------------------------------------------
# CgroupParent override
#
# Placing a container in a non-default cgroup tree changes which eBPF device
# filter programs are active for the container's cgroup. A container placed
# in the root cgroup ("/") or an ancestor cgroup may inherit a more permissive
# device policy than DinD's isolated cgroup subtree would provide, potentially
# bypassing the device-cgroup restrictions that protect block device reads.
# Also breaks PID/resource accounting isolation.
# ---------------------------------------------------------------------------

violation["CgroupParent override is not allowed"] {
    is_container_create
    count(input.Body.HostConfig.CgroupParent) > 0
}

# ---------------------------------------------------------------------------
# Bind-mount restrictions: ALL host filesystem bind-mounts are denied.
#
# Rationale: any bind-mount of a host path into a nested container can be
# used to escape the DinD sandbox — either directly (sensitive dirs) or via
# symlink indirection from a writable directory. Rather than maintaining an
# ever-growing deny-list, we take the safe default: deny ALL bind-mounts.
#
# Containers that need persistent storage should use Docker volumes (--volume
# or --mount type=volume), which are managed by the DinD daemon and stored
# under /var/lib/docker/volumes — fully isolated from the DinD host FS.
# Ephemeral scratch space should use --tmpfs or named volumes.
#
# What is allowed:
#   --mount type=volume,...  – Docker-managed named/anonymous volumes
#   --mount type=tmpfs,...   – in-memory tmpfs mounts
#
# What is denied:
#   -v /any/host/path:/dest  – any host bind-mount (legacy -v syntax)
#   --mount type=bind,...    – any host bind-mount (structured syntax)
# ---------------------------------------------------------------------------

# Block ALL legacy -v SOURCE:DEST bind-mounts (source starts with "/").
violation[msg] {
    is_container_create
    some i
    bind := input.Body.HostConfig.Binds[i]
    # Binds format is "SOURCE:DEST" or "SOURCE:DEST:opts".
    # Named volumes have no leading "/" in the source (e.g. "myvolume:/data").
    # Anonymous volumes have an empty or absent source.
    parts := split(bind, ":")
    src := parts[0]
    # Block only absolute paths (host bind-mounts).
    startswith(src, "/")
    msg := sprintf("host bind-mount not allowed (use Docker volumes instead): %v", [src])
}

# Block structured --mount type=bind mounts.
violation[msg] {
    is_container_create
    some i
    mount := input.Body.HostConfig.Mounts[i]
    mount.Type == "bind"
    msg := sprintf("host bind-mount not allowed (use Docker volumes instead): %v", [mount.Source])
}

# Block BindMounts enriched field (opa-docker-authz resolved symlinks).
# Any entry here is a host bind-mount that was resolved by the plugin.
violation[msg] {
    is_container_create
    some i
    bm := input.BindMounts[i]
    msg := sprintf("host bind-mount not allowed (use Docker volumes instead): %v", [bm.Source])
}
