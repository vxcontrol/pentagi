package checker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/state"
)

type mockState struct {
	vars    map[string]loader.EnvVar
	envPath string
}

func (m *mockState) GetVar(key string) (loader.EnvVar, bool) {
	if val, exists := m.vars[key]; exists {
		return val, true
	}
	return loader.EnvVar{}, false
}

func (m *mockState) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
	return m.vars, make(map[string]bool, len(names))
}

func (m *mockState) GetEnvPath() string {
	return m.envPath
}

func (m *mockState) Exists() bool                         { return true }
func (m *mockState) Reset() error                         { return nil }
func (m *mockState) Commit() error                        { return nil }
func (m *mockState) IsDirty() bool                        { return false }
func (m *mockState) GetEulaConsent() bool                 { return true }
func (m *mockState) SetEulaConsent() error                { return nil }
func (m *mockState) SetStack(stack []string) error        { return nil }
func (m *mockState) GetStack() []string                   { return []string{} }
func (m *mockState) SetVar(name, value string) error      { return nil }
func (m *mockState) ResetVar(name string) error           { return nil }
func (m *mockState) SetVars(vars map[string]string) error { return nil }
func (m *mockState) ResetVars(names []string) error       { return nil }
func (m *mockState) GetAllVars() map[string]loader.EnvVar { return m.vars }

func TestCheckFileExistsAndReadable(t *testing.T) {
	f, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if !checkFileExists(f.Name()) {
		t.Errorf("file should exist")
	}
	if !checkFileIsReadable(f.Name()) {
		t.Errorf("file should be readable")
	}

	os.Remove(f.Name())
	if checkFileExists(f.Name()) {
		t.Errorf("file should not exist")
	}
	if checkFileIsReadable(f.Name()) {
		t.Errorf("removed file should not be readable")
	}

	if checkFileExists("") {
		t.Errorf("empty path should not exist")
	}
	if checkFileExists("/nonexistent/path/file.txt") {
		t.Errorf("nonexistent file should not exist")
	}
}

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		name         string
		vars         map[string]loader.EnvVar
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "existing variable",
			vars:         map[string]loader.EnvVar{"FOO": {Value: "bar"}},
			key:          "FOO",
			defaultValue: "default",
			expected:     "bar",
		},
		{
			name:         "non-existing variable",
			vars:         map[string]loader.EnvVar{},
			key:          "MISSING",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty variable value",
			vars:         map[string]loader.EnvVar{"EMPTY": {Value: ""}},
			key:          "EMPTY",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "nil state",
			vars:         nil,
			key:          "ANY",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appState state.State
			if tt.vars != nil {
				appState = &mockState{vars: tt.vars}
			}

			result := getEnvVar(appState, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvVar() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractVersionFromOutput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"docker-compose version 1.29.2, build 5becea4c", "1.29.2"},
		{"Docker Compose version v2.12.2", "2.12.2"},
		{"Docker version 20.10.8, build 3967b7d", "20.10.8"},
		{"no version here", ""},
		{"v1.0.0-alpha", "1.0.0"},
		{"version: 3.14.159", "3.14.159"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input_%s", tt.input), func(t *testing.T) {
			result := extractVersionFromOutput(tt.input)
			if result != tt.expected {
				t.Errorf("extractVersionFromOutput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckDockerComposeVersionWithRunner(t *testing.T) {
	t.Run("uses docker compose v2 output", func(t *testing.T) {
		calls := 0
		result := checkDockerComposeVersionWithRunner(func(name string, args ...string) ([]byte, error) {
			calls++
			if name != "docker" {
				t.Fatalf("unexpected command %q", name)
			}
			if len(args) != 2 || args[0] != "compose" || args[1] != "version" {
				t.Fatalf("unexpected args: %v", args)
			}

			return []byte("Docker Compose version v2.12.2"), nil
		})

		if calls != 1 {
			t.Fatalf("expected 1 command invocation, got %d", calls)
		}
		if result.Version != "2.12.2" {
			t.Fatalf("expected version 2.12.2, got %q", result.Version)
		}
		if !result.Valid {
			t.Fatal("expected docker compose version to be valid")
		}
	})

	t.Run("parses version from stdout even when error is returned", func(t *testing.T) {
		calls := 0
		result := checkDockerComposeVersionWithRunner(func(name string, args ...string) ([]byte, error) {
			calls++
			return []byte("Docker Compose version v2.12.2"), errors.New("exit status 1")
		})

		if calls != 1 {
			t.Fatalf("expected 1 command invocation, got %d", calls)
		}
		if result.Version != "2.12.2" {
			t.Fatalf("expected version 2.12.2, got %q", result.Version)
		}
		if !result.Valid {
			t.Fatal("expected docker compose version to remain valid when stdout is parseable")
		}
	})

	t.Run("fails when docker compose is unavailable", func(t *testing.T) {
		calls := 0
		result := checkDockerComposeVersionWithRunner(func(name string, args ...string) ([]byte, error) {
			calls++
			return nil, errors.New("executable file not found")
		})

		if calls != 1 {
			t.Fatalf("expected 1 command invocation, got %d", calls)
		}
		if result.Version != "" {
			t.Fatalf("expected empty version, got %q", result.Version)
		}
		if result.Valid {
			t.Fatal("expected docker compose check to be invalid")
		}
	})
}

func TestCheckVersionCompatibility(t *testing.T) {
	tests := []struct {
		version    string
		minVersion string
		expected   bool
	}{
		{"1.2.3", "1.2.0", true},
		{"1.2.0", "1.2.0", true},
		{"1.1.9", "1.2.0", false},
		{"2.0.0", "1.9.9", true},
		{"1.2.3", "1.2.4", false},
		{"", "1.0.0", false},
		{"1.0.0", "", false},
		{"invalid", "1.0.0", false},
		{"1.0.0", "invalid", false},
		{"1.2", "1.2.0", false}, // fewer parts should fail
		{"1.2.0", "1.2", true},  // more parts should pass
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.version, tt.minVersion), func(t *testing.T) {
			result := checkVersionCompatibility(tt.version, tt.minVersion)
			if result != tt.expected {
				t.Errorf("checkVersionCompatibility(%q, %q) = %v, want %v",
					tt.version, tt.minVersion, result, tt.expected)
			}
		})
	}
}

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		imageRef string
		imageID  string
		wantName string
		wantTag  string
		wantHash string
	}{
		{"alpine:3.18", "sha256:abc", "alpine", "3.18", "sha256:abc"},
		{"nginx", "", "nginx", "latest", ""},
		{"nginx", "sha256:def", "nginx", "latest", "sha256:def"},
		{"repo/nginx:1.2", "", "repo/nginx", "1.2", ""},
		// Docker Hub spells an official image in full; everywhere else it is just "ubuntu",
		// and that is the name the update server is asked about.
		{"docker.io/library/ubuntu:latest", "", "ubuntu", "latest", ""},
		{"nginx@sha256:deadbeef", "", "nginx", "latest", "sha256:deadbeef"},

		// Everywhere that is not Docker Hub, the host is PART of the name and stays.
		//
		// These three used to expect the opposite, and that expectation is what
		// made the defect invisible: the service stores a reference the way
		// compose writes it — `gcr.io/cadvisor/cadvisor`, host and all — and
		// matches on exact equality. A client reporting `cadvisor/cadvisor` names
		// a different image (a Docker Hub repository that is not cAdvisor), so the
		// component matched nothing, came back `repository_not_tracked`, and took
		// its whole stack's resolution down to `not_tracked` with it. Two tests
		// were green the whole time, one on each side, pinning opposite rules.
		{"myreg:5000/foo/bar:tag@sha256:beef", "", "myreg:5000/foo/bar", "tag", "sha256:beef"},
		{"localhost:5000/myapp:v1.0", "", "localhost:5000/myapp", "v1.0", ""},
		{"registry.example.com/team/app", "", "registry.example.com/team/app", "latest", ""},

		// The two components that are not on Docker Hub, exactly as their compose
		// defaults spell them and exactly as the service has them stored.
		{"gcr.io/cadvisor/cadvisor:v0.51.0", "", "gcr.io/cadvisor/cadvisor", "v0.51.0", ""},
		{"quay.io/prometheuscommunity/postgres-exporter:v0.16.0", "",
			"quay.io/prometheuscommunity/postgres-exporter", "v0.16.0", ""},

		// Docker Hub under each of the names it answers to: here the host IS
		// dropped, because that is the form compose writes and the service stores.
		{"index.docker.io/library/redis:7", "", "redis", "7", ""},
		{"registry-1.docker.io/prom/node-exporter:v1.8.2", "", "prom/node-exporter", "v1.8.2", ""},
		{"", "", "", "", ""},
		{"ubuntu:", "", "ubuntu", "latest", ""},
		{"ubuntu:@sha256:hash", "", "ubuntu", "latest", "sha256:hash"},

		// Every version-numbered tag the shipped compose files use. These are the cases the
		// old dot-means-a-port rule got wrong: the tag was swallowed into the repository
		// name and replaced with "latest", so the pair the server matches on named nothing
		// and the component silently got no answer.
		{"neo4j:5.26.2", "", "neo4j", "5.26.2", ""},
		{"grafana/grafana:11.4.0", "", "grafana/grafana", "11.4.0", ""},
		{"grafana/loki:3.3.2", "", "grafana/loki", "3.3.2", ""},
		{"jaegertracing/all-in-one:1.56.0", "", "jaegertracing/all-in-one", "1.56.0", ""},
		{"victoriametrics/victoria-metrics:v1.108.1", "", "victoriametrics/victoria-metrics", "v1.108.1", ""},
		{"otel/opentelemetry-collector-contrib:0.116.1", "", "otel/opentelemetry-collector-contrib", "0.116.1", ""},
		{"minio/minio:RELEASE.2025-07-23T15-54-02Z", "", "minio/minio", "RELEASE.2025-07-23T15-54-02Z", ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("parse_%s", tt.imageRef), func(t *testing.T) {
			if tt.imageRef == "" {
				info := parseImageRef(tt.imageRef, tt.imageID)
				if info != nil {
					t.Errorf("parseImageRef(%q) should return nil for empty input", tt.imageRef)
				}
				return
			}

			info := parseImageRef(tt.imageRef, tt.imageID)
			if info == nil {
				t.Errorf("parseImageRef(%q) = nil, want non-nil", tt.imageRef)
				return
			}

			if info.Name != tt.wantName {
				t.Errorf("parseImageRef(%q).Name = %q, want %q", tt.imageRef, info.Name, tt.wantName)
			}
			if info.Tag != tt.wantTag {
				t.Errorf("parseImageRef(%q).Tag = %q, want %q", tt.imageRef, info.Tag, tt.wantTag)
			}
			if info.Hash != tt.wantHash {
				t.Errorf("parseImageRef(%q).Hash = %q, want %q", tt.imageRef, info.Hash, tt.wantHash)
			}
		})
	}
}

// TestEveryComposeImageParsesIntoAMatchablePair reads the shipped compose files and checks
// that every image reference in them splits into a repository and tag the update server can
// match. It reads the files rather than a copy, so an image pinned to a new kind of tag is
// caught here instead of quietly costing that component its answer.
func TestEveryComposeImageParsesIntoAMatchablePair(t *testing.T) {
	for _, path := range []string{
		"../files/fs/docker-compose.yml",
		"../files/fs/docker-compose-graphiti.yml",
		"../files/fs/docker-compose-langfuse.yml",
		"../files/fs/docker-compose-observability.yml",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		for _, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "image:") {
				continue
			}
			ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "image:"))
			// Every image line is `${<COMPONENT>_IMAGE:-<default>}` now — the server chooses
			// the tag and the installer only writes what it was given — so the default inside
			// the wrapper is what a fresh installation actually pulls, and it is the thing
			// worth checking. Stripping only PENTAGI_IMAGE was right when it was the only
			// parameterised line and became wrong the day the rest followed.
			//
			// The wrapper is never handed to parseImageRef in production: the checker parses
			// what the daemon reports for a running container, which is already resolved. A
			// template reaching the parser produces nonsense either way — `${NEO4J_IMAGE:-neo4j`
			// as a repository, or the quieter `latest}` as a tag when the default has a slash
			// and the wrapper's own colon gets mistaken for a registry port.
			if inner, found := strings.CutPrefix(ref, "${"); found {
				if _, def, ok := strings.Cut(strings.TrimSuffix(inner, "}"), ":-"); ok {
					ref = def
				}
			}

			info := parseImageRef(ref, "")
			if info == nil {
				t.Errorf("%s: %q did not parse at all", filepath.Base(path), ref)
				continue
			}
			// A tag glued onto the repository is the exact shape of the old defect, and it
			// produces a pair that matches nothing on the server.
			if strings.Contains(info.Name, ":") {
				t.Errorf("%s: %q parsed to repository %q — the tag was swallowed into the name",
					filepath.Base(path), ref, info.Name)
			}
			// Only a reference that genuinely omits its tag should come out as "latest".
			if info.Tag == "latest" && !strings.HasSuffix(ref, ":latest") && strings.Contains(ref, ":") {
				t.Errorf("%s: %q parsed to tag %q, but the reference pins a different one",
					filepath.Base(path), ref, info.Tag)
			}
			// A reference that names a registry must still name it after parsing.
			//
			// The service stores what compose writes and matches on exact equality,
			// so a dropped host is a repository that names a DIFFERENT image —
			// `cadvisor/cadvisor` on Docker Hub instead of the one on gcr.io. The
			// component then resolves to nothing and is answered
			// `repository_not_tracked` for the life of the installation. Checking
			// every compose line rather than the two that have a host today is the
			// point: the next third-party image added to these files gets the same
			// guarantee without anybody remembering this.
			if host, _, found := strings.Cut(ref, "/"); found && !isDockerHubHost(host) {
				if strings.ContainsAny(host, ".:") && !strings.HasPrefix(info.Name, host+"/") {
					t.Errorf("%s: %q parsed to repository %q — the registry host was dropped, "+
						"which names a different image than the one running",
						filepath.Base(path), ref, info.Name)
				}
			}
		}
	}
}

func TestCheckCPUResources(t *testing.T) {
	result := checkCPUResources()
	// assuming test machine has at least 2 CPUs, this is reasonable for CI/dev environments
	if !result {
		t.Logf("CPU check returned false - this is expected on machines with < 2 CPUs")
	}
}

func TestCheckMemoryResources(t *testing.T) {
	tests := []struct {
		name                     string
		needsForPentagi          bool
		needsForGraphiti         bool
		needsForLangfuse         bool
		needsForObservability    bool
		expectMinimumRequirement bool
	}{
		{
			name:                     "no components needed",
			needsForPentagi:          false,
			needsForGraphiti:         false,
			needsForLangfuse:         false,
			needsForObservability:    false,
			expectMinimumRequirement: true,
		},
		{
			name:                     "pentagi only",
			needsForPentagi:          true,
			needsForGraphiti:         false,
			needsForLangfuse:         false,
			needsForObservability:    false,
			expectMinimumRequirement: false, // requires actual memory check
		},
		{
			name:                     "all components",
			needsForPentagi:          true,
			needsForGraphiti:         true,
			needsForLangfuse:         true,
			needsForObservability:    true,
			expectMinimumRequirement: false, // requires actual memory check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkMemoryResources(tt.needsForPentagi, tt.needsForGraphiti, tt.needsForLangfuse, tt.needsForObservability)
			if tt.expectMinimumRequirement && !result {
				t.Errorf("checkMemoryResources() should return true when no components are needed")
			}
			// note: we can't reliably test memory checks across different environments
			// the function will work correctly based on actual system memory
		})
	}
}

func TestCheckDiskSpaceWithContext(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		workerImageExists bool
		pentagiInstalled  bool
		graphitiConnected bool
		graphitiExternal  bool
		graphitiInstalled bool
		langfuseConnected bool
		langfuseExternal  bool
		langfuseInstalled bool
		obsConnected      bool
		obsExternal       bool
		obsInstalled      bool
		expectHighSpace   bool // whether we expect it to require more disk space
	}{
		{
			name:              "all installed and running",
			workerImageExists: true,
			pentagiInstalled:  true,
			graphitiConnected: true,
			graphitiExternal:  false,
			graphitiInstalled: true,
			langfuseConnected: true,
			langfuseExternal:  false,
			langfuseInstalled: true,
			obsConnected:      true,
			obsExternal:       false,
			obsInstalled:      true,
			expectHighSpace:   false, // minimal space needed
		},
		{
			name:              "no worker images",
			workerImageExists: false,
			pentagiInstalled:  true,
			expectHighSpace:   true, // needs to download images
		},
		{
			name:              "pentagi not installed",
			workerImageExists: true,
			pentagiInstalled:  false,
			expectHighSpace:   false, // moderate space for components
		},
		{
			name:              "langfuse local not installed",
			workerImageExists: true,
			pentagiInstalled:  true,
			langfuseConnected: true,
			langfuseExternal:  false,
			langfuseInstalled: false,
			expectHighSpace:   false, // moderate space for components
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkDiskSpaceWithContext(
				ctx,
				tt.workerImageExists,
				tt.pentagiInstalled,
				tt.graphitiConnected,
				tt.graphitiExternal,
				tt.graphitiInstalled,
				tt.langfuseConnected,
				tt.langfuseExternal,
				tt.langfuseInstalled,
				tt.obsConnected,
				tt.obsExternal,
				tt.obsInstalled,
			)
			// note: actual disk space check depends on OS and available space
			// we mainly test that the function doesn't panic and returns a boolean
			_ = result
		})
	}
}

func TestCreateTempFileForTesting(t *testing.T) {
	// helper test to ensure temp file creation works for other tests
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "checker_test_file")

	// create test file
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	// verify it exists and is readable
	if !checkFileExists(testFile) {
		t.Error("test file should exist")
	}
	if !checkFileIsReadable(testFile) {
		t.Error("test file should be readable")
	}

	// note: directory readability behavior is platform-dependent
	// so we skip this assertion
}

func TestConstants(t *testing.T) {
	// test that critical constants are defined
	if InstallerVersion == "" {
		t.Error("InstallerVersion should not be empty")
	}
	if UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
	if !strings.Contains(UserAgent, InstallerVersion) {
		t.Error("UserAgent should contain InstallerVersion")
	}
	if DefaultUpdateServerEndpoint == "" {
		t.Error("DefaultUpdateServerEndpoint should not be empty")
	}

	// test memory and disk constants are reasonable
	if MinFreeMemGB <= 0 {
		t.Error("MinFreeMemGB should be positive")
	}
	if MinFreeMemGBForPentagi <= 0 {
		t.Error("MinFreeMemGBForPentagi should be positive")
	}
	if MinFreeDiskGB <= 0 {
		t.Error("MinFreeDiskGB should be positive")
	}
	if MinFreeDiskGBForWorkerImages <= MinFreeDiskGB {
		t.Error("MinFreeDiskGBForWorkerImages should be larger than MinFreeDiskGB")
	}
}

func TestCheckImageExistsEdgeCases(t *testing.T) {
	ctx := context.Background()

	// test with nil client
	result := checkImageExists(ctx, nil, "nginx:latest")
	if result {
		t.Error("checkImageExists should return false for nil client")
	}

	// test with empty image name
	// note: we can't test with real Docker client in unit tests
	// but we can test that the function handles edge cases gracefully
}

func TestGetImageInfoEdgeCases(t *testing.T) {
	ctx := context.Background()

	// test with nil client
	result := getImageInfo(ctx, nil, "nginx:latest")
	if result != nil {
		t.Error("getImageInfo should return nil for nil client")
	}

	// test with empty image name
	// again, testing without real Docker client
}

func TestImageInfoStructure(t *testing.T) {
	// test ImageInfo struct
	info := &ImageInfo{
		Name: "nginx",
		Tag:  "latest",
		Hash: "sha256:abc123",
	}

	if info.Name != "nginx" {
		t.Error("ImageInfo.Name should be set correctly")
	}
	if info.Tag != "latest" {
		t.Error("ImageInfo.Tag should be set correctly")
	}
	if info.Hash != "sha256:abc123" {
		t.Error("ImageInfo.Hash should be set correctly")
	}
}

func TestCheckVolumesExist(t *testing.T) {
	// note: this test uses a mock volume list since we can't rely on real Docker client in unit tests
	// in real scenarios, checkVolumesExist is called with actual Docker API client

	// test with nil client
	t.Run("nil_client", func(t *testing.T) {
		ctx := context.Background()
		volumeNames := []string{"test-volume"}
		result := checkVolumesExist(ctx, nil, volumeNames)
		if result {
			t.Error("checkVolumesExist should return false for nil client")
		}
	})

	// test with empty volume list
	t.Run("empty_volume_list", func(t *testing.T) {
		ctx := context.Background()
		// we can't create a real client in unit tests, so we pass nil
		// the function should handle empty list gracefully
		result := checkVolumesExist(ctx, nil, []string{})
		if result {
			t.Error("checkVolumesExist should return false for empty volume list")
		}
	})

	// note: testing actual volume matching requires Docker integration tests
	// the function logic handles:
	// 1. Exact match: "pentagi-data" matches "pentagi-data"
	// 2. Compose prefix match: "pentagi-data" matches "pentagi_pentagi-data"
	// 3. Compose prefix match: "pentagi-postgres-data" matches "myproject_pentagi-postgres-data"
	//
	// This ensures compatibility with Docker Compose project prefixes
}

// mockDockerVolume simulates Docker API volume structure for testing
type mockDockerVolume struct {
	Name string
}

func TestCheckVolumesExist_MatchingLogic(t *testing.T) {
	// unit test for the matching logic without Docker client
	// simulates what checkVolumesExist does internally

	tests := []struct {
		name            string
		existingVolumes []string
		searchVolumes   []string
		expected        bool
		description     string
	}{
		{
			name:            "exact match",
			existingVolumes: []string{"pentagi-data", "other-volume"},
			searchVolumes:   []string{"pentagi-data"},
			expected:        true,
			description:     "should match exact volume name",
		},
		{
			name:            "compose prefix match",
			existingVolumes: []string{"pentagi_pentagi-data", "pentagi_pentagi-ssl"},
			searchVolumes:   []string{"pentagi-data"},
			expected:        true,
			description:     "should match volume with compose project prefix",
		},
		{
			name:            "arbitrary prefix match",
			existingVolumes: []string{"myproject_pentagi-postgres-data", "other_volume"},
			searchVolumes:   []string{"pentagi-postgres-data"},
			expected:        true,
			description:     "should match volume with any compose prefix",
		},
		{
			name:            "no match",
			existingVolumes: []string{"other-volume", "another-volume"},
			searchVolumes:   []string{"pentagi-data"},
			expected:        false,
			description:     "should not match when volume doesn't exist",
		},
		{
			name:            "partial name should not match",
			existingVolumes: []string{"pentagi-data-backup", "my-pentagi-data"},
			searchVolumes:   []string{"pentagi-data"},
			expected:        false,
			description:     "should not match partial names without underscore separator",
		},
		{
			name:            "match multiple search volumes",
			existingVolumes: []string{"proj_pentagi-data", "langfuse-data"},
			searchVolumes:   []string{"pentagi-data", "langfuse-data", "missing-volume"},
			expected:        true,
			description:     "should return true if any search volume matches",
		},
		{
			name:            "empty existing volumes",
			existingVolumes: []string{},
			searchVolumes:   []string{"pentagi-data"},
			expected:        false,
			description:     "should return false when no volumes exist",
		},
		{
			name:            "multiple compose prefixes",
			existingVolumes: []string{"proj1_vol1", "proj2_vol2", "pentagi_pentagi-ssl"},
			searchVolumes:   []string{"pentagi-ssl"},
			expected:        true,
			description:     "should find volume among multiple compose projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// simulate the matching logic from checkVolumesExist
			result := false
			for _, volumeName := range tt.searchVolumes {
				for _, existingVolume := range tt.existingVolumes {
					if existingVolume == volumeName || strings.HasSuffix(existingVolume, "_"+volumeName) {
						result = true
						break
					}
				}
				if result {
					break
				}
			}

			if result != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func (m *mockState) WriteVars(vars map[string]string) error { return nil }
