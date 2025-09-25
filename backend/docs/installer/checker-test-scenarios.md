# Checker Test Scenarios

This document outlines test scenarios for the installer's system checking functionality, focusing on failure modes and their detection.

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
**Setup**: Block DNS resolution (modify /etc/hosts or firewall)
**Expected**:
- `SysNetworkFailures`: ["• DNS resolution failed for docker.io"]
- UI shows specific DNS failure with resolution steps

### 7. Network Issues - HTTPS Blocked
**Setup**: Block outbound HTTPS (port 443)
**Expected**:
- `SysNetworkFailures`: ["• Cannot reach external services via HTTPS"]
- UI shows HTTPS failure with proxy configuration info

### 8. Network Issues - Docker Registry Blocked
**Setup**: Block docker.io specifically
**Expected**:
- `SysNetworkFailures`: ["• Cannot pull Docker images from registry"]
- UI shows registry access failure

### 9. Behind Corporate Proxy
**Setup**: Network requires proxy, but not configured
**Expected**:
- Multiple network failures
- UI shows proxy configuration instructions for HTTP_PROXY/HTTPS_PROXY

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

## Environment Variable Tests

### 1. HTTP_PROXY Auto-Detection
**Setup**: Set HTTP_PROXY before running installer
**Expected**: PROXY_URL in .env automatically populated

### 2. DOCKER_HOST Inheritance
**Setup**: Set DOCKER_HOST, DOCKER_TLS_VERIFY, DOCKER_CERT_PATH
**Expected**: Values synchronized to .env on first run

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

# Simulate DNS issues
echo "127.0.0.1 docker.io" | sudo tee -a /etc/hosts

# Test with proxy
export HTTP_PROXY=http://proxy:3128
export HTTPS_PROXY=http://proxy:3128

# Test remote Docker
export DOCKER_HOST=tcp://remote:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/path/to/certs
```

## Verification

Each scenario should:
1. Be detected correctly by the checker
2. Show appropriate error message in UI
3. Provide actionable fix instructions
4. Not block other checks unnecessarily
5. Work under both privileged and unprivileged users
