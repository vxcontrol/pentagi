# Checker Test Scenarios

This document outlines test scenarios for the installer's system checking functionality, focusing on failure modes and their detection.

The scenarios fall into two groups. The first is what the checker observes about this machine — Docker, disk, memory, local connectivity. The second is what the PentAGI Cloud API answers, and there the interesting part is that a failed call is not one condition: "the server could not be reached", "the daily allowance is spent" and "this license tier has no access" need different things from the user, and each has its own reason code and its own test.

## Test Scenarios

### 1. Docker Not Installed
**Setup**: Remove Docker from the system
**Expected**:
- `DockerErrorType`: "not_installed"
- `DockerApiAccessible`: false
- `DockerInstalled`: false
- UI shows: "Docker Not Installed" with installation instructions

### 2. Docker Daemon Not Running
**Setup**: Install Docker but stop the daemon (e.g., quit Docker Desktop on macOS)
**Expected**:
- `DockerErrorType`: "not_running"
- `DockerApiAccessible`: false
- `DockerInstalled`: true
- UI shows: "Docker Daemon Not Running" with start instructions

### 3. Docker Permission Denied
**Setup**: Run installer as non-docker user on Linux
**Expected**:
- `DockerErrorType`: "permission"
- `DockerApiAccessible`: false
- `DockerInstalled`: true
- UI shows: "Docker Permission Denied" with usermod instructions

### 4. Remote Docker Connection Failed
**Setup**: Set DOCKER_HOST to invalid address
**Expected**:
- `DockerErrorType`: "api_error"
- `DockerApiAccessible`: false
- UI shows: "Docker API Connection Failed" with DOCKER_HOST troubleshooting

### 5. Write Permissions Denied
**Setup**: Run installer in read-only directory
**Expected**:
- `EnvDirWritable`: false
- UI shows: "Write Permissions Required" with chmod instructions

### 6. Network Issues - DNS Failure
**Setup**: Break name resolution itself — firewall port 53, or point the resolver at a dead server. Adding `127.0.0.1 docker.io` to /etc/hosts does **not** trigger this: the name resolves, and what fails is scenario 7.
**Expected**:
- `SysNetworkFailures`: ["• DNS resolution failed for docker.io"]
- `SysNetworkOK`: false
- UI shows specific DNS failure with resolution steps

The probe is a plain host lookup and it never goes through `PROXY_URL`. On a network where only the proxy can resolve names, this line appears while everything the installer actually does works — treat it as a hint, not a verdict.

### 7. Network Issues - HTTPS Blocked
**Setup**: Block outbound HTTPS (port 443), or point docker.io at a closed port
**Expected**:
- `SysNetworkFailures`: ["• Cannot reach external services via HTTPS"]
- UI shows HTTPS failure with proxy configuration info

The probe is a 5-second GET to https://docker.io through `PROXY_URL` when one is set, and **any answer below 500 counts as reachable** — an interception proxy replying 403 to the probe passes it. It says the path is open, not that the registry is usable.

### 8. Network Issues - Docker Registry Blocked
**Setup**: Block docker.io for the Docker daemon
**Expected**:
- `SysNetworkFailures`: ["• Cannot pull Docker images from registry"]
- UI shows registry access failure

The probe pulls `debian:latest` and succeeds if **either** the main daemon or the worker daemon can pull it, so a broken remote worker alone does not raise this line. It is skipped entirely when either client is absent — an unreachable daemon is scenarios 2-4, not this one.

### 9. Behind Corporate Proxy
**Setup**: Network requires a proxy; `PROXY_URL` unset, or set while the daemon has no proxy of its own
**Expected**:
- Multiple network failures
- UI shows proxy configuration instructions for HTTP_PROXY/HTTPS_PROXY

Only the HTTPS probe uses `PROXY_URL`. The lookup resolves locally, and the pull goes through the Docker daemon with the daemon's own proxy configuration — so a correctly configured `PROXY_URL` can still leave two of the three lines standing.

**None of these three probes touches update.pentagi.com.** `SysNetworkOK` is not evidence that the update check will work, and a failed update check is not evidence that the registry is blocked. The update server has its own scenarios below.

### 10. Low Memory
**Setup**: System with < 2GB available RAM
**Expected**:
- `SysMemoryOK`: false
- `SysMemoryAvailable`: < 2.0
- UI shows memory requirements with specific numbers

### 11. Low Disk Space
**Setup**: System with < 25GB free space
**Expected**:
- `SysDiskFreeSpaceOK`: false
- `SysDiskAvailable`: < 25.0
- UI shows disk requirements with cleanup suggestions

### 12. Worker Docker Environment Issues
**Setup**: Configure DOCKER_HOST for remote, but remote unavailable
**Expected**:
- `WorkerEnvApiAccessible`: false
- UI shows worker environment troubleshooting

## Update Server Scenarios

The installer talks to the PentAGI Cloud API over three calls — `POST /api/v1/proxy/updates/check`, `GET /api/v1/proxy/packages/info` and `GET /api/v1/proxy/packages/download`. Every one of them negotiates TLS and solves a proof-of-work challenge before the request is sent, so there is no local server to point the installer at: these scenarios are exercised in unit tests by substituting the call function with the error the transport would have produced (`cmd/installer/cloud/updates_test.go`, `cmd/installer/cloud/packages_test.go`, `cmd/installer/cloud/errors_test.go`), and observed in the field the way a support case arrives.

A call that produces no answer is classified into one of eight reasons. The distinction is the whole point — a spent allowance and a broken proxy call for opposite reactions, and collapsing them into one "update server unavailable" flag tells the user nothing they can act on.

| `reason` | What it means | `retryable` | `retry_after` |
|---|---|---|---|
| `unreachable` | never reached a verdict: no network, no DNS, refused proxy, connection died mid-flight | true | — |
| `timeout` | a deadline: the caller's, or the proof-of-work budget | true | — |
| `rate_limited` | too many requests in a rolling window | true | when the server sent one |
| `quota_exceeded` | the allowance for the period is spent | true | when the server sent one |
| `forbidden` | the license tier has no access to this call, or the client is blocked | false | — |
| `not_found` | no such package or version — an answer, not an outage | false | — |
| `rejected` | the server refused the request as malformed | false | — |
| `server_error` | the server failed on its side | true | — |

**What holds for every failed update check**, and is worth asserting in each scenario below:

- `UpdateServerAccessible`: false, `StackUpdates`: nil, and **all six** per-stack verdicts (`PentagiIsUpToDate`, `WorkerIsUpToDate`, `InstallerIsUpToDate`, `GraphitiIsUpToDate`, `LangfuseIsUpToDate`, `ObservabilityIsUpToDate`) reset to false. Leaving one behind is how a verdict from an earlier successful check gets presented as a current one, with nothing on screen to suggest it is old — the worker flag was the one that used to be missed.
- `update_failure`: `{reason, message, retry_after, retryable}` on the check result, plus a warning line `update check failed (<reason>): <error>` in the log.
- The gather continues. A check that could not be made is a fact about the installation, not a reason to abandon everything else the checker learned, so no other field is affected.
- The update overview screen says the check did not complete and that applying an update will pull whatever the compose files currently point at. It must **never** print "Everything is up to date" — that sentence belongs to an answer that arrived and listed nothing.
- The "Update Installer" entry disappears from Maintenance: it requires the server, and without a package description there is nothing that could be verified. The "Update PentAGI" entry stays **for any stack that is installed**, because none of them is known to be current — it leads to the overview, which says exactly that. On a machine with no compose stack installed the entry is absent whatever the check did: the offer is gated on a stack being installed *and* not up to date, so resetting the verdicts is only half of what makes it appear.
- The reason and the cooldown are what separate these scenarios; the overview's failure text is currently the same sentence for all of them. Assert the reason on the check result and in the log, not a per-reason wording that does not exist.

### 1. Update Server Unreachable
**Setup**: point `UPDATE_SERVER_HOST` at a name that does not resolve or a host that drops 443, or set `PROXY_URL` to a port nothing listens on
**Expected**:
- `update_failure.reason`: "unreachable", `retryable`: true, `retry_after`: 0
- `message` carries the transport error (`dial tcp ...`, proxy refused, TLS failure)
**Must not**:
- claim any stack is up to date, or keep the previous answer on screen
- be read as evidence about the Docker registry — that is scenarios 6-8, a different probe against a different host

A configuration the client refuses to build with never becomes a call at all: `UPDATE_SERVER_HOST=http://...` is an error rather than a silent upgrade to https (a user who wrote `http` meant it), as is a `PROXY_URL` with no host. It is still recorded as a failed check, but as an *unclassified* error — `reason` falls back to "unreachable" and `retryable` stays false, while the message names the setting. Read the message before blaming the network.

### 2. Proof-of-Work Challenge Times Out
**Setup**: in tests, return the SDK's challenge-timeout error from the substituted call. In the field it appears on CPU-limited VMs and throttled laptops — **do not manufacture load to reproduce it**; record the host it happened on instead.
**Expected**:
- `update_failure.reason`: "timeout", `retryable`: true
- The budget is 60 seconds per call, deliberately double the SDK default, because slow hardware genuinely needs the extra seconds
**Must not**:
- be presented as a network failure. The network was fine and the client ran out of budget; the two lead to opposite investigations, which is the entire reason this reason exists.
- be assumed to have been retried. A timed-out challenge is not retried by the transport — if a retry is wanted, the installer has to make it.

### 3. Request Rate Limited
**Setup**: several checks inside one rolling window. Note how easily this happens: each update step refreshes the answer in a deferred call, and updating everything recurses over the stacks — one press of "Update PentAGI" is therefore one check per stack touched, not one check. A limit sized for a single request breaks the client instead of throttling it.
**Expected**:
- `update_failure.reason`: "rate_limited", `retryable`: true
- `retry_after`: the server-advertised cooldown when it sent one, 0 otherwise. It is recorded on the check result and rendered nowhere — the overview prints its single failure sentence and the operation panel prints `failed to check updates: <message>`. Assert the value on the result, not a screen that tells the user when to come back: that screen does not exist yet.
**Must not**:
- retry in a tight loop. The per-minute and general limits are already retried by the transport (up to three attempts, waiting the advertised cooldown capped at 10 seconds); the per-hour and per-day ones are not retried at all, because that wait is too long to hold a screen — so what the installer sees for those is the final answer.
- report anything as up to date.

### 4. Daily Allowance Exhausted
**Setup**: spend the account's daily allowance (in tests: a quota error with the daily scope and a `Retry-After`)
**Expected**:
- `update_failure.reason`: "quota_exceeded", `retryable`: true
- `retry_after`: time until the allowance resets, taken from the server's answer — again recorded, not displayed
- On screen this is indistinguishable from any other failed check: the same overview sentence, the same `failed to check updates: <message>` line. What separates it is the reason and the cooldown on the check result and the warning in the log, so that is what the test asserts
**Must not**:
- be shown as a permission problem — the license is fine, the period's allowance is not
- be retried before the reset; the allowance does not return sooner for being asked
- report anything as up to date

Every check costs a slice of that allowance, which is why the answer is gathered once and shared, and why the package description is held rather than requested twice for the same build.

### 5. Access Forbidden for the License Tier
**Setup**: call an endpoint the tier does not carry, or run as a blocked client
**Expected**:
- `update_failure.reason`: "forbidden", `retryable`: false, `retry_after`: 0
- This includes a quota answer whose scope is "blocked": a tier that never had access is not an allowance that is spent, and there is nothing to wait for
- Nothing on screen names the license: the failure text is the same sentence as for every other reason. The distinction lives in `reason` and in the log line — and it is the one failure where waiting cannot help, because the license, not time, is what would have to change
**Must not**:
- offer "try again later", schedule a retry, or count down a cooldown that does not exist
- report anything as up to date

Related: a license key the client cannot introspect is dropped **before** the call and the request goes anonymously, with `license key is not usable, continuing without it` in the log. The client is still built — the update check is the one thing that works without a license, and refusing to run would leave the user with no way to learn that the key is the problem. The symptom of ignoring that warning is precisely this scenario: a tier nobody expected.

### 6. Package Not Published for This Platform (404)
**Setup**: ask `packages/info` for a component, version, os and arch combination that has no published package — a plugin version the server never published for one of the two plugin architectures, or an installer build that exists for another platform only
**Expected**:
- Reason "not_found", `retryable`: false. This arrives on the package calls, so the check result is untouched — the update check itself may well have succeeded.
- Installer self-update: the screen shows `Could not get the package description: ...` and Enter does nothing. Without the description there is no length, sha256 or signature to check a file against, so there is nothing that could be fetched safely.
- Jaeger plugin: the observability stack update aborts **before any image is pulled** — the plugin step runs first for exactly this reason — and the installation keeps the plugin it had.
**Must not**:
- retry. A 404 is an answer; asking again produces the same one.
- download anything without a description to verify it against.
- leave a staged `.new` file behind. Staging files are removed on failure, and again at the start of the next attempt, because nothing else ever looks for them.

An unfamiliar platform never produces this 404 on the plugin path: a component whose OS is not linux is dropped before anything is fetched, and an architecture with no plugin file defined for it fails locally with `no jaeger plugin file is defined for architecture <a>` before the call is made. A "not_found" here means the server publishes no package for a version and platform the installer legitimately asked about.

### 7. Malformed Request Refused (400)
**Setup**: in a test, hand the client a request the contract does not accept — an unknown update strategy, more than 30 reported artefacts, an image digest carrying its `sha256:` prefix, a file component reported under `images`, a version that is not semver
**Expected** — two distinct paths:
- **Caught locally.** The client validates before marshalling and does not send. Only the update check leaves a mark on the check result: `update check request is malformed: <field>` is recorded as a failed check with the fallback reason "unreachable" and `retryable` false, because nothing classified it. The package calls produce `package info request is malformed: ...` and `package download request is malformed: ...`, which are returned to their caller and surface as an operation error or as `Could not get the package description: ...` — the check result is untouched, exactly as in scenario 6. Either way the message, which names the offending field, is the useful half, and catching it here costs no call and no slice of the allowance.
- **Refused by the server.** Reason "rejected", `retryable`: false.
**Must not**:
- resend the identical request. It cannot start succeeding, and each attempt costs a challenge and a slice of the allowance.
- be presented to the user as a server outage. This is the installer's own bug; the field name in the message is what makes it fixable.

### 8. Download Does Not Match Its Description
**Setup**: serve a package that is truncated, padded, altered while keeping its length, or signed with a key that is not the release key
**Expected** — three checks, all describing the plain file (transport encryption is transparent and changes none of them), all decided only after the last byte has arrived:
- Length ≠ `size` → `package is <n> bytes, expected <m>`. The encrypted response carries no `Content-Length`, so the size from `packages/info` is the only length the client ever learns — this is a real check, not a formality, and it is also what the progress percentage is computed against.
- sha256 ≠ `hash` → `package sha256 is <x>, expected <y>`
- Ed25519 signature over the file's SHA-512 does not verify → `package signature does not verify: ...`
**Must not**:
- treat the destination as usable before the call returns nil. Bytes are written as they arrive, so a failed download leaves a partial or unverified file: the installer build is deleted, and the Jaeger plugin's `.new` is deleted with the target left untouched — which is the entire reason it is staged next to the target rather than written in place.
- keep using the description the failed download was checked against. A build republished under the same version since it was described fails all three checks identically forever; the held description is dropped so the next attempt asks again.
- report success on a matching length and hash alone. The signature is what makes the package ours rather than merely intact.

### 9. Server Failed on Its Side
**Setup**: the server answers an internal error or a bad gateway
**Expected**:
- `update_failure.reason`: "server_error", `retryable`: true
- This is one of the few failures the client retries on its own — up to three attempts, a few seconds apart — so what the installer sees is the state after them, not the first failure
**Must not**: report anything as up to date; suggest the user reconfigure anything — nothing on this machine caused it.

### 10. The Confirming Check After an Update Fails
**Setup**: apply an update, then make the check that follows it fail — any of scenarios 1-5 and 9 at that moment
**Expected**:
- The local comparison still runs and reports per component: `<component>: expected <a>, got <b>`, or `no published digest to compare against` when the answer named none
- The menus keep offering the update, because no stack is known to be current
**Must not**:
- report the stack as current on the strength of the local comparison alone. Only the server knows whether something newer appeared while the update was being applied.
- report a component that differs from its target as "newer than the offered build". That verdict — the registry legitimately moving on between the answer and the pull — is only available when a fresh answer says there is nothing left to apply; without one, a difference is a mismatch and is reported as such.

## Environment Variable Tests

### 1. HTTP_PROXY Auto-Detection
**Setup**: Set HTTP_PROXY before running installer
**Expected**: PROXY_URL in .env automatically populated

### 2. DOCKER_HOST Inheritance
**Setup**: Set DOCKER_HOST, DOCKER_TLS_VERIFY, DOCKER_CERT_PATH
**Expected**: 
- Values synchronized to .env on first run via DoSyncNetworkSettings()
- DOCKER_CERT_PATH migrated to PENTAGI_DOCKER_CERT_PATH (host path) + DOCKER_CERT_PATH set to /opt/pentagi/docker/ssl (container path)

## Edge Cases

### 1. Docker Version Too Old
**Setup**: Docker 19.x installed
**Expected**:
- `DockerVersionOK`: false
- UI shows version upgrade instructions

### 2. Docker Compose Missing
**Setup**: Docker installed without Compose
**Expected**:
- `DockerComposeInstalled`: false
- UI shows Compose installation instructions

### 3. Multiple Failures
**Setup**: No Docker + network issues + low resources
**Expected**: All issues shown in priority order:
1. Environment file
2. Write permissions
3. Docker issues
4. Resource issues
5. Network issues

## Testing Commands

```bash
# Simulate Docker not running (macOS)
osascript -e 'quit app "Docker"'

# Simulate permission issues (Linux)
sudo gpasswd -d $USER docker

# Simulate network issues
sudo iptables -A OUTPUT -p tcp --dport 443 -j DROP

# Simulate an unreachable registry — NOT a DNS failure: the name still resolves,
# so this trips the HTTPS probe (scenario 7), not the lookup (scenario 6)
echo "127.0.0.1 docker.io" | sudo tee -a /etc/hosts

# Test with proxy
export HTTP_PROXY=http://proxy:3128
export HTTPS_PROXY=http://proxy:3128

# Test remote Docker
export DOCKER_HOST=tcp://remote:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/path/to/certs  # auto-migrated to PENTAGI_DOCKER_CERT_PATH on startup

# Update server unreachable: a name that does not resolve
export UPDATE_SERVER_HOST=update.invalid

# Update server unreachable: a proxy nothing listens on
export PROXY_URL=http://127.0.0.1:1

# Refused before a call is ever made — http is an error, not a silent upgrade to https.
# Recorded as a failed check whose message names the setting.
export UPDATE_SERVER_HOST=http://update.pentagi.com
```

`UPDATE_STRATEGY` cannot be used to trigger the malformed-request path: an unrecognised value is validated, logged and replaced with `preview`, because a typo must not leave an installation unable to learn about security updates.

Everything the environment cannot produce — rate limits, exhausted allowances, a forbidden tier, a timed-out challenge, a 404, a tampered download — is driven in unit tests by substituting the call function with the error the transport would have returned:

```bash
go test ./cmd/installer/cloud/...    # classification, envelope, download verification
go test ./cmd/installer/checker/...  # what a failed check does to the check result
```

## Verification

Each scenario should:
1. Be detected correctly by the checker
2. Show appropriate error message in UI
3. Provide actionable fix instructions
4. Not block other checks unnecessarily
5. Work under both privileged and unprivileged users

For the update server scenarios, two more, and they are the ones that matter:

6. Never claim an installation is up to date on the strength of a check that produced no answer — "nothing to update" and "we never found out" are different sentences, and only the first may be printed as reassurance
7. Never retry what cannot succeed. `retryable` is the property to branch on: telling a user to wait out a forbidden endpoint wastes their time, and discarding a cooldown the server did advertise wastes the license. Both are carried on the check result for whoever branches on them — no screen reads either one today, so assert them there.
