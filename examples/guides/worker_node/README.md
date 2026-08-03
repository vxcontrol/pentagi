# Worker Node Configuration Bundle

Configuration files for the hardened Docker-in-Docker (dind) sandbox described in
[worker_node.md](../worker_node.md). Deploy them with the steps in that guide — it
covers certificates, paths, and the `dind.env` settings these files read.

| File | Installed as | Purpose |
|---|---|---|
| `authz.rego` | `/etc/docker/dind/authz/authz.rego` | OPA authorization policy — fail-closed gate on every dind API call |
| `daemon.json` | `/etc/docker/dind/authz/daemon.json` | dind daemon config; sets `seccomp.json` as the daemon-wide default profile |
| `seccomp.json` | `/etc/docker/dind/authz/seccomp.json` | Seccomp profile for all nested containers; blocks block-device `mknod` |
| `dockerd-entrypoint.sh` | `/etc/docker/dind/scripts/dockerd-entrypoint.sh` | dind entrypoint with `DOCKER_API_HOST` / `DOCKER_DNS_SERVERS` support |
| `run-dind.sh` | `/usr/local/bin/run-dind` | Creates the dind container, bootstraps the OPA plugin, purges on policy change |
| `docker-dind-tls.sh` | `/usr/local/bin/docker-dind-tls` | Docker CLI wrapper for the dind TLS endpoint |
| `docker-dind-sock.sh` | `/usr/local/bin/docker-dind-sock` | Docker CLI wrapper for the dind Unix socket |
| `dind-cleanup.sh` | `/usr/local/bin/dind-cleanup` | Removes nested containers older than `MAX_AGE_HOURS` |
| `dind-cleanup.service` / `.timer` | `/etc/systemd/system/` | Hourly systemd timer running the cleanup script |
| `policy-tests.sh` | `/usr/local/bin/dind-policy-tests` | Isolation test suite for the policy and seccomp profile |

The three files under `/etc/docker/dind/authz/` are mounted read-only as `/etc/docker`
inside the dind container: that directory is both the daemon config directory and the
policy source the OPA plugin reads.

Every script sources `/etc/docker/dind/dind.env` (overridable via the `DIND_ENV`
environment variable), so per-host settings live in one place.
