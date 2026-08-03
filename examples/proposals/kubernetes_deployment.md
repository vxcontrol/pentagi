# Kubernetes Deployment RFC

## Summary

Issue [#324](https://github.com/vxcontrol/pentagi/issues/324) asks
whether PentAGI can run on Kubernetes. Today PentAGI is built and
documented around Docker Compose and the installer, and there is no
supported Kubernetes path. This RFC sketches what a future,
incremental Kubernetes-compatibility effort could look like and names
the parts of the current design that make it non-trivial.

This document does not implement runtime behavior. It does not add
Helm charts, Kubernetes manifests, Kustomize bases, an operator, or
CRDs. It does not change `docker-compose.yml`, the installer, the
backend, the database schema, or any environment variable. It does
not claim that PentAGI runs on Kubernetes today, because it does not.
It is a design surface for maintainers to push back on before any
deployment code lands.

The RFC is intentionally staged and docs-first. PentAGI's flow
executor currently talks to a Docker daemon over a bind-mounted
socket, and that single fact drives most of the difficulty below. A
naive "wrap the containers in a Deployment" approach would either
break flow execution or smuggle in implicit, hard-to-inspect
lifecycle behavior -- close to the patterns pushed back on during PR
[#268](https://github.com/vxcontrol/pentagi/pull/268) review. The
proposed path below stays explicit and reviewable: every Kubernetes
resource a future implementation would create should be something an
operator can see with `kubectl`, not a hidden background mechanism.

## Goals

- Capture, in one place, the concrete reasons PentAGI does not run on
  Kubernetes today, grounded in the current Compose and installer
  design rather than in guesswork.
- Map each Compose-era assumption (secrets, volumes, service
  discovery, TLS, health, networking, the container executor,
  observability, image selection, migrations) to its candidate
  Kubernetes equivalent.
- Identify the one genuinely hard problem -- the Docker-socket flow
  executor -- and lay out candidate approaches with their trade-offs,
  without choosing one.
- Propose an incremental, docs-first path so that any later
  implementation can be reviewed in small, self-contained slices.
- Keep operators in control of secrets, persistence, network reach,
  and the privilege level of flow execution at every step.
- Give maintainers a single artifact to accept, reject, or reshape
  before any chart, manifest, or operator code is written.

## Non-Goals

- This RFC does not add Helm charts, raw manifests, Kustomize
  overlays, an operator, or CRDs. No deployment artifact ships with
  this document.
- This RFC does not modify `docker-compose.yml`, the installer, or the
  current supported deployment path. Compose remains the only
  supported deployment model until maintainers decide otherwise.
- This RFC does not add, rename, or change any environment variable,
  and does not change any default in `.env.example` or the backend
  configuration.
- This RFC does not change the backend, the database schema, the
  generated code, the GraphQL or REST surface, or the flow executor.
- This RFC does not propose hidden background orchestration, an
  implicit queue, or out-of-band lifecycle state to make Kubernetes
  work. Carrying forward the explicit lesson from PR
  [#268](https://github.com/vxcontrol/pentagi/pull/268) review: any
  future Kubernetes resource (Pod, Job, PVC, Secret) must be visible
  and manageable through the standard Kubernetes API, not buried in
  process memory.
- This RFC does not claim parity with the Compose deployment. Some
  capabilities (notably the privileged Docker-socket executor) may
  never map cleanly, and this document does not promise that they
  will.
- This RFC does not pick a single executor strategy, a single
  ingress controller, a single storage class, or a single secret
  backend. Those are deferred to a later implementation RFC.

## Current Deployment Assumptions

This section describes how PentAGI is deployed today, because the
Kubernetes considerations only make sense against the current shape.
Everything here is drawn from `docker-compose.yml`, the installer
docs, and the backend, not from a hypothetical setup.

- **Compose-oriented topology.** The supported deployment is Docker
  Compose (directly or via the installer). The core stack is the
  `pentagi` backend, a `pgvector` PostgreSQL instance, a `pgexporter`
  metrics sidecar, and a `scraper` service. Optional stacks add
  Graphiti / Neo4j, Langfuse, and an observability bundle
  (OpenTelemetry collector, Grafana, VictoriaMetrics, and friends).
- **The flow executor uses the Docker socket.** This is the central
  fact for Kubernetes. The `pentagi` service bind-mounts the host
  Docker socket (`${PENTAGI_DOCKER_SOCKET:-/var/run/docker.sock}` to
  `/var/run/docker.sock`) and the backend's Docker client connects
  via `client.FromEnv`, honoring
  `DOCKER_HOST` (default `unix:///var/run/docker.sock`). During a
  flow, PentAGI creates and destroys terminal/worker containers
  against that daemon. The executor is effectively "talk to a Docker
  daemon and spawn sibling containers," not "run one long-lived
  process."
- **Elevated privileges by design.** Because the backend drives the
  Docker socket, the `pentagi` service runs `user: root:root`
  (commented in-file as "while using docker.sock") and carries
  Docker-related toggles (`DOCKER_INSIDE`, `DOCKER_NET_ADMIN`,
  `DOCKER_GID=998`, `DOCKER_WORK_DIR`). This privilege level is
  intrinsic to the current executor, not incidental.
- **Local-named volumes for state.** Persistent state uses Docker
  local volumes: `pentagi-data` mounted at `/opt/pentagi/data`,
  Postgres data in `pentagi-postgres-data`, plus `pentagi-ssl`,
  `scraper-ssl`, and `pentagi-ollama`. These assume a single host
  with local volume drivers.
- **Configuration and secrets via `.env`.** Provider keys, the
  database DSN, embedding settings, TLS material, and feature toggles
  are passed as environment variables sourced from `.env`. There is no
  externalized secret store in the default path; the env file is the
  source of truth.
- **Service discovery by Compose DNS.** Services find each other by
  Compose service name on user-defined bridge networks
  (`pentagi-network`, and the optional `observability-network` and
  `langfuse-network`). The backend reaches Postgres, the scraper, and
  optional services by name.
- **TLS terminates at the backend.** The backend listens on `8443`
  and is published to `${PENTAGI_LISTEN_IP:-127.0.0.1}:8443`,
  defaulting to loopback. There is no separate ingress or reverse
  proxy in the core stack; TLS is handled inside the container.
- **Health via Compose healthchecks.** Ordering uses Compose
  `depends_on` with `condition: service_healthy` (for example the
  backend waits on `pgvector`). Health is expressed as container
  healthchecks, not as orchestrator probes.
- **Database migrations on startup.** The backend embeds its SQL
  migrations and runs them with goose at process start
  (`goose.Up`). There is no separate migration step; the backend
  migrates itself when it boots.
- **Image selection via env override.** The backend image is
  `${PENTAGI_IMAGE:-vxcontrol/pentagi:latest}`, and worker/tool images
  are similarly overridable. Air-gapped and mirror setups already rely
  on these overrides (see the README's note on restricted networks,
  Docker mirrors, and proxies).

## Kubernetes Compatibility Considerations

For each Compose-era assumption above, this section names the
candidate Kubernetes equivalent and the friction. Nothing here is a
committed design; it is a map of the problem space.

- **Secrets and configuration.** The `.env` model maps to Kubernetes
  `Secret` objects (provider keys, DB credentials, TLS material) and
  `ConfigMap` objects (non-secret toggles). This is mostly mechanical.
  The open part is whether to keep a flat env-injection model
  (`envFrom` a Secret/ConfigMap) or move toward referenced secrets,
  and whether to integrate external secret managers. No change to the
  variable names themselves is needed.
- **Persistent volumes.** The local-named volumes map to
  `PersistentVolumeClaim`s backed by a cluster `StorageClass`. Postgres
  state in particular wants a `StatefulSet` with a stable claim, or an
  external managed Postgres. The friction is that several volumes today
  assume single-host locality and `ReadWriteOnce` semantics; a future
  design has to be explicit about access modes and about whether
  Postgres is in-cluster or external.
- **Service discovery.** Compose service-name DNS maps cleanly to
  Kubernetes `Service` objects and in-cluster DNS. This is among the
  lowest-friction items; the backend would address Postgres and the
  scraper by Service name instead of Compose name.
- **Ingress and TLS.** Today TLS terminates in the backend on `8443`
  bound to loopback. On Kubernetes the candidate is an `Ingress` (or
  Gateway API) with TLS via cert-manager, or preserving in-pod TLS and
  exposing it through a passthrough Service. The open question is
  whether to keep TLS in the backend or move termination to the edge;
  both are viable and have different operational profiles.
- **Health checks.** Compose healthchecks and `depends_on` map to
  `readinessProbe` and `livenessProbe`. Startup ordering that Compose
  expresses with `service_healthy` becomes readiness-gated rollout
  plus application-level retry, since Kubernetes does not block one
  workload's start on another's health the way Compose does.
- **Network policies.** The implicit isolation of Compose user-defined
  networks maps to Kubernetes `NetworkPolicy`. This is an opportunity
  to make the currently-implicit segmentation explicit, but it is also
  net-new surface that has to be designed rather than translated.
- **Flow / container execution model (the hard problem).** This is the
  item that does not translate mechanically. The backend expects a
  Docker daemon and spawns sibling containers over the socket.
  Kubernetes does not hand workloads a Docker socket, and modern
  clusters do not run Docker as the node runtime. Candidate
  approaches, each with real trade-offs and none free of cost:
  - **Kubernetes-native execution.** Teach the executor to create
    ephemeral `Pod`s or `Job`s through the Kubernetes API instead of
    Docker containers. Most idiomatic and the most inspectable
    (`kubectl get pods/jobs` shows exactly what a flow is running),
    but the largest backend change, and it requires an in-cluster
    `ServiceAccount` with pod-create RBAC, which is its own risk.
  - **Docker-in-Docker sidecar.** Run a DinD daemon next to the
    backend and keep the existing socket-based executor. Smallest
    backend change, but DinD typically needs a privileged container,
    has known stability and storage caveats, and concentrates risk in
    one privileged pod.
  - **Sandboxed runtimes.** Pair Kubernetes-native execution with a
    stronger isolation runtime (gVisor, Kata, sysbox, or similar) for
    the worker pods, since flow workers run untrusted, agent-driven
    commands. This is a hardening layer on top of native execution,
    not an alternative to it.
  Whatever is chosen, the PR
  [#268](https://github.com/vxcontrol/pentagi/pull/268) lesson
  applies: the running work must be visible and manageable through
  standard Kubernetes objects, not tracked only inside the backend
  process.
- **Observability.** The optional OpenTelemetry / Grafana /
  VictoriaMetrics stack maps to in-cluster deployments or, more
  likely, to whatever the operator's cluster already runs. The
  candidate direction is to make PentAGI emit to existing cluster
  observability rather than bundling its own, with the bundled stack
  as an opt-in for clusters that have none.
- **Image overrides.** The existing `PENTAGI_IMAGE` and related
  per-image overrides map directly to image fields in pod specs, which
  is helpful for air-gapped and mirror deployments. This is
  low-friction and reuses an existing mechanism rather than inventing
  one.
- **Upgrade and migration path.** Because the backend runs goose
  migrations on startup, a rolling update could run migrations from
  whichever replica starts first. On Compose with a single backend
  this is fine; on Kubernetes with multiple replicas it is not. A
  future design needs an explicit decision: a one-shot migration
  `Job` (or init container) gated ahead of the rollout, or an
  enforced single-writer constraint. This must be settled before any
  multi-replica backend deployment is suggested.

## Proposed Incremental Path

The path is deliberately docs-first so each step is small enough to
review and reject in isolation. No step below is started by this RFC;
this is the proposed sequence, not a commitment.

1. **This RFC.** Land the design surface, confirm the boundaries
   (docs-only, no charts, no executor change yet), and let maintainers
   accept, reshape, or decline the direction.
2. **Executor strategy decision.** Before any manifest exists, settle
   the single hardest question in a follow-up RFC: how flow workers
   run on Kubernetes (native Pods/Jobs vs DinD vs sandboxed runtime),
   and what privilege and RBAC that implies. Everything else depends
   on this.
3. **Stateless-core reference manifests.** Once the executor decision
   exists, a minimal, clearly-labeled reference for the parts that do
   translate cleanly -- backend Deployment/Service, Postgres via
   StatefulSet or external, Secrets/ConfigMaps, probes, a migration
   Job -- explicitly marked experimental and excluding flow execution.
4. **Flow execution on the chosen model.** Implement the executor
   decision from step 2 behind the existing Docker path, so Compose
   keeps working unchanged and Kubernetes execution is additive and
   opt-in.
5. **Packaging and operator guide.** Only after the above is proven,
   consider a Helm chart or operator and a Kubernetes operator guide
   (in the spirit of the existing `examples/` material), so packaging
   lands on top of a working deployment rather than ahead of it.

Each step is self-contained: maintainers can stop after any step
without leaving PentAGI in a half-migrated state, and Compose remains
the supported path throughout.

## Open Questions

- Which executor model should PentAGI target first -- Kubernetes-native
  Pods/Jobs, a DinD sidecar, or a sandboxed runtime -- and is it worth
  supporting more than one?
- Should Postgres (and pgvector) run in-cluster as a StatefulSet, or
  should the Kubernetes path assume an external managed database?
- Should TLS continue to terminate in the backend, or move to an
  Ingress / Gateway with cert-manager?
- How should the startup goose migration be handled under multiple
  backend replicas -- a gating migration Job, an init container, or an
  enforced single-writer?
- What RBAC is acceptable for the backend's ServiceAccount if it
  creates worker Pods/Jobs, and how is that least-privileged?
- Should the observability stack be bundled, or should PentAGI default
  to emitting into the operator's existing cluster observability?
- Is Helm, an operator, or plain manifests the right packaging once a
  working deployment exists, and which should ship first?
- How should air-gapped and mirror deployments be expressed on
  Kubernetes, reusing the existing image-override mechanism?
- What is the minimum Kubernetes version and feature set
  (StorageClass, Ingress/Gateway, NetworkPolicy support) a future
  reference deployment should assume?

## Security and Operational Considerations

Moving PentAGI onto Kubernetes changes its security posture, and the
changes should be designed in rather than discovered later.

- **Privilege of the executor.** The current model effectively grants
  the backend host-level container control via the Docker socket. Any
  Kubernetes equivalent (pod-create RBAC, a privileged DinD sidecar,
  or a sandboxed runtime) carries comparable or different risk. The
  privilege level must be explicit, least-privilege, and visible to
  operators -- not an unstated side effect of "making it work."
- **RBAC and namespacing.** If the backend creates worker Pods/Jobs,
  its ServiceAccount needs scoped permissions in a dedicated
  namespace, never cluster-admin. Flow workers should be confined to
  that namespace with their own constrained ServiceAccount.
- **Untrusted workloads.** Flow workers run agent-driven, untrusted
  commands. On Kubernetes that argues for pod security standards,
  seccomp/AppArmor profiles, dropped capabilities, and a sandboxed
  runtime for worker pods, rather than running them as ordinary
  privileged pods.
- **Secret handling.** Kubernetes `Secret`s are base64, not encrypted,
  at rest by default. A future design should call out
  encryption-at-rest, optional external secret managers, and the fact
  that provider keys and the DB DSN are sensitive. No secret should be
  baked into an image or committed to a manifest.
- **Network segmentation.** The implicit Compose-network isolation
  should be reproduced with explicit `NetworkPolicy`, defaulting to
  deny and opening only the required backend-to-Postgres,
  backend-to-scraper, and worker egress paths.
- **No unsafe defaults.** Any future reference deployment must not
  default to a privileged or host-network pod, must not expose the
  backend publicly without TLS, and must not widen RBAC for
  convenience. The Compose default already binds the backend to
  loopback; the Kubernetes default should be equally conservative.
- **Inspectable lifecycle.** Per the PR
  [#268](https://github.com/vxcontrol/pentagi/pull/268) lesson, flow
  execution state on Kubernetes should be representable as real
  objects an operator can list and delete, so a stuck or runaway flow
  is visible and stoppable through the cluster API rather than only
  through backend internals.

## Test and Validation Strategy

A future implementation should be validated against the points below
before being described as anything more than experimental. This RFC
itself is validated only as documentation.

- **Local clusters.** Bring-up and teardown on kind and minikube as
  the baseline developer-facing validation, since they need no cloud
  account.
- **Manifest and chart linting.** If/when manifests or a chart exist,
  `kubectl apply --dry-run=server`, `kubeconform` (or equivalent), and
  `helm lint` / `helm template` in CI before anything is published.
- **Migration validation.** Verify the chosen migration approach is
  safe under a rolling update with more than one backend replica, so
  goose does not run concurrently from multiple pods.
- **End-to-end flow test.** Run at least one real flow on the chosen
  executor model and confirm worker Pods/Jobs are created, complete,
  are cleaned up, and are visible via `kubectl` for their lifetime.
- **Security review.** Run pod security and RBAC checks (for example
  with a policy linter) to confirm least-privilege, deny-by-default
  network policies, and no privileged or host-network defaults.
- **Compose parity guard.** Confirm the existing Docker Compose path
  is unchanged and still the supported default, so the Kubernetes work
  remains additive and opt-in throughout.

## Recommendation and Requested Decision

This RFC is docs-only and additive: it asks maintainers to accept a direction and its boundaries, not to commit to any manifests. Recommended boundaries for v1:

- **Compose stays the supported default** and is unchanged; all Kubernetes work is additive and opt-in at every step.
- **Docs-first sequencing:** land this design surface (step 1), then settle the executor model (step 2) in a dedicated follow-up RFC before any manifest is written, since flow-worker execution is the hardest and most security-sensitive question. This RFC intentionally does not pre-decide it.
- **No charts or executor changes** are proposed here; reference manifests and packaging (Helm or an operator) are deferred to later, separately reviewable steps.

Decision requested from the maintainer: accept, reshape, or decline this docs-first direction and the Compose-stays-default boundary. If accepted, the immediate next step is a focused executor-strategy RFC (Kubernetes-native Pods/Jobs vs DinD vs sandboxed runtime), which is the blocker for everything downstream.

## References

- Issue [#324](https://github.com/vxcontrol/pentagi/issues/324):
  Kubernetes deployment request.
- PR [#268](https://github.com/vxcontrol/pentagi/pull/268): source of
  the explicit-lifecycle / no-hidden-state lesson carried forward
  here.
- `docker-compose.yml`: current service topology, the Docker-socket
  mount, the `root:root` executor, named volumes, and networks
  described in "Current Deployment Assumptions."
- The README sections on Docker image configuration and on restricted
  networks, Docker mirrors, and proxies: the existing image-override
  mechanism reused under "Image overrides."
