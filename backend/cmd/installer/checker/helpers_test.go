package checker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"pentagi/cmd/installer/loader"
	"pentagi/cmd/installer/state"

	"github.com/moby/moby/client"
)

// Covers checkFileExists and checkFileIsReadable on the same file before and after removal.
func TestHelpers_CheckFile_TellsAReadableFileFromAMissingOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), "testfile")
	if err := os.WriteFile(path, []byte("test content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !checkFileExists(path) || !checkFileIsReadable(path) {
		t.Errorf("an existing file: exists=%t readable=%t, want both", checkFileExists(path), checkFileIsReadable(path))
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	for _, missing := range []string{path, "", "/nonexistent/path/file.txt"} {
		if checkFileExists(missing) {
			t.Errorf("%q is reported to exist", missing)
		}
	}
	if checkFileIsReadable(path) {
		t.Error("a removed file is reported readable")
	}
}

func TestHelpers_GetEnvVar_FallsBackFromTheValueToTheDefaults(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]loader.EnvVar // nil means no state at all
		want string
	}{
		{"a set value", map[string]loader.EnvVar{"FOO": {Value: "bar"}}, "bar"},
		{"an absent variable", map[string]loader.EnvVar{}, "fallback"},
		{"an empty value", map[string]loader.EnvVar{"FOO": {Value: ""}}, "fallback"},
		{"an empty value with a config default", map[string]loader.EnvVar{"FOO": {Default: "configured"}}, "configured"},
		{"no state", nil, "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appState state.State
			if tt.vars != nil {
				appState = &mockState{vars: tt.vars}
			}
			if got := getEnvVar(appState, "FOO", "fallback"); got != tt.want {
				t.Errorf("getEnvVar = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHelpers_ExtractVersionFromOutput_FindsTheFirstSemanticVersion(t *testing.T) {
	for input, want := range map[string]string{
		"docker-compose version 1.29.2, build 5becea4c": "1.29.2",
		"Docker Compose version v2.12.2":                "2.12.2",
		"v1.0.0-alpha":                                  "1.0.0",
		"no version here":                               "",
		"":                                              "",
	} {
		if got := extractVersionFromOutput(input); got != want {
			t.Errorf("extractVersionFromOutput(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestHelpers_CheckDockerComposeVersionWithRunner_ReadsTheVersionFromStdout(t *testing.T) {
	tests := []struct {
		name   string
		output string
		err    error
		want   DockerVersion
	}{
		{"docker compose v2", "Docker Compose version v2.12.2", nil, DockerVersion{Version: "2.12.2", Valid: true}},
		{"a version on stdout survives a failing exit", "Docker Compose version v2.12.2", errors.New("exit status 1"),
			DockerVersion{Version: "2.12.2", Valid: true}},
		{"docker compose is unavailable", "", errors.New("executable file not found"), DockerVersion{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			got := checkDockerComposeVersionWithRunner(func(name string, args ...string) ([]byte, error) {
				calls = append(calls, strings.Join(append([]string{name}, args...), " "))
				return []byte(tt.output), tt.err
			})

			if len(calls) != 1 || calls[0] != "docker compose version" {
				t.Errorf("ran %q, want one docker compose version", calls)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestHelpers_CheckVersionCompatibility_ComparesPartByPart(t *testing.T) {
	tests := []struct {
		version, minVersion string
		want                bool
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
		{"1.2", "1.2.0", false},
		{"1.2.0", "1.2", true},
	}

	for _, tt := range tests {
		if got := checkVersionCompatibility(tt.version, tt.minVersion); got != tt.want {
			t.Errorf("checkVersionCompatibility(%q, %q) = %t, want %t", tt.version, tt.minVersion, got, tt.want)
		}
	}
}

func TestHelpers_ParseImageRef_SplitsTheReferenceTheUpdateServerMatches(t *testing.T) {
	tests := []struct {
		imageRef, imageID, wantName, wantTag, wantHash string
	}{
		{"alpine:3.18", "sha256:abc", "alpine", "3.18", "sha256:abc"},
		{"nginx", "", "nginx", "latest", ""},
		{"nginx", "sha256:def", "nginx", "latest", "sha256:def"},
		{"repo/nginx:1.2", "", "repo/nginx", "1.2", ""},
		{"docker.io/library/ubuntu:latest", "", "ubuntu", "latest", ""},
		{"nginx@sha256:deadbeef", "", "nginx", "latest", "sha256:deadbeef"},
		// Anywhere but Docker Hub the host is part of the name.
		{"myreg:5000/foo/bar:tag@sha256:beef", "", "myreg:5000/foo/bar", "tag", "sha256:beef"},
		{"localhost:5000/myapp:v1.0", "", "localhost:5000/myapp", "v1.0", ""},
		{"localhost:5000/myapp", "", "localhost:5000/myapp", "latest", ""},
		{"registry.example.com/team/app", "", "registry.example.com/team/app", "latest", ""},
		{"gcr.io/cadvisor/cadvisor:v0.51.0", "", "gcr.io/cadvisor/cadvisor", "v0.51.0", ""},
		{"index.docker.io/library/redis:7", "", "redis", "7", ""},
		{"registry-1.docker.io/prom/node-exporter:v1.8.2", "", "prom/node-exporter", "v1.8.2", ""},
		{"ubuntu:", "", "ubuntu", "latest", ""},
		{"ubuntu:@sha256:hash", "", "ubuntu", "latest", "sha256:hash"},
	}

	for _, tt := range tests {
		info := parseImageRef(tt.imageRef, tt.imageID)
		if info == nil {
			t.Errorf("parseImageRef(%q) = nil", tt.imageRef)
			continue
		}
		if *info != (ImageInfo{Name: tt.wantName, Tag: tt.wantTag, Hash: tt.wantHash}) {
			t.Errorf("parseImageRef(%q) = %+v, want %s:%s hash %q", tt.imageRef, *info, tt.wantName, tt.wantTag, tt.wantHash)
		}
	}
	if info := parseImageRef("", ""); info != nil {
		t.Errorf("parseImageRef of an empty reference = %+v, want nil", info)
	}
}

func TestHelpers_ParseImageRef_SplitsEveryShippedComposeImageIntoAMatchablePair(t *testing.T) {
	for _, path := range composeFiles {
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
			// The default inside `${X_IMAGE:-default}` is what a fresh installation pulls.
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
			if strings.Contains(info.Name, ":") {
				t.Errorf("%s: %q parsed to repository %q, the tag was swallowed into the name", filepath.Base(path), ref, info.Name)
			}
			if info.Tag == "latest" && !strings.HasSuffix(ref, ":latest") && strings.Contains(ref, ":") {
				t.Errorf("%s: %q parsed to tag %q, but the reference pins a different one", filepath.Base(path), ref, info.Tag)
			}
			if host, _, found := strings.Cut(ref, "/"); found && !isDockerHubHost(host) &&
				strings.ContainsAny(host, ".:") && !strings.HasPrefix(info.Name, host+"/") {
				t.Errorf("%s: %q parsed to repository %q, the registry host was dropped", filepath.Base(path), ref, info.Name)
			}
		}
	}
}

// Host-relative: on a host with N >= 2 CPUs only a threshold above N fails it.
func TestHelpers_CheckCPUResources_AgreesWithTheHostCPUCount(t *testing.T) {
	if got, want := checkCPUResources(), runtime.NumCPU() >= 2; got != want {
		t.Errorf("checkCPUResources = %t with %d CPUs, want %t", got, runtime.NumCPU(), want)
	}
}

func TestHelpers_CalculateRequiredMemoryGB_AddsWhatEachStartingComponentNeeds(t *testing.T) {
	tests := []struct {
		name                                       string
		pentagi, graphiti, langfuse, observability bool
		want                                       float64
	}{
		{"nothing to start", false, false, false, false, 0.5},
		{"pentagi", true, false, false, false, 1.0},
		{"graphiti", false, true, false, false, 2.5},
		{"langfuse", false, false, true, false, 2.0},
		{"observability", false, false, false, true, 2.0},
		{"everything", true, true, true, true, 6.0},
	}

	for _, tt := range tests {
		if got := calculateRequiredMemoryGB(tt.pentagi, tt.graphiti, tt.langfuse, tt.observability); got != tt.want {
			t.Errorf("%s: %v GB, want %v", tt.name, got, tt.want)
		}
	}
}

func TestHelpers_CheckMemoryResources_PassesWhenNothingStartsAndTheHostHasRoom(t *testing.T) {
	if !checkMemoryResources(false, false, false, false) {
		t.Error("nothing to start is refused")
	}
	check := map[string]func(float64) bool{"darwin": checkDarwinMemory, "linux": checkLinuxMemory}[runtime.GOOS]
	if check == nil {
		t.Skipf("no memory check on %s", runtime.GOOS)
	}
	// The platform comparison, both ways, with requirements no host can miss or meet.
	if !check(0) || check(1e6) {
		t.Errorf("0 GB passes: %t, a million GB passes: %t; want only the first", check(0), check(1e6))
	}
	// Host-relative and one way only: pentagi alone needs 1 GB, and the margin keeps a host near it from flaking.
	got, free := checkMemoryResources(true, false, false, false), getAvailableMemoryGB()
	if free > 1.5 && !got {
		t.Errorf("%.1f GB free refuses to start pentagi", free)
	}
}

func TestHelpers_CountLocalComponentsToInstall_CountsLocalStacksNotYetInstalled(t *testing.T) {
	tests := []struct {
		name                                             string
		pentagi                                          bool
		graphitiConnected, graphitiExternal, graphitiIns bool
		langfuseConnected, langfuseExternal, langfuseIns bool
		obsConnected, obsExternal, obsIns                bool
		want                                             int
	}{
		{"everything installed", true, true, false, true, true, false, true, true, false, true, 0},
		{"pentagi missing", false, false, false, false, false, false, false, false, false, false, 1},
		{"a local graphiti missing", true, true, false, false, false, false, false, false, false, false, 1},
		{"a local langfuse missing", true, false, false, false, true, false, false, false, false, false, 1},
		{"a local observability missing", true, false, false, false, false, false, false, true, false, false, 1},
		{"external stacks need no local space", true, true, true, false, true, true, false, true, true, false, 0},
		{"everything missing", false, true, false, false, true, false, false, true, false, false, 4},
	}

	for _, tt := range tests {
		got := countLocalComponentsToInstall(tt.pentagi,
			tt.graphitiConnected, tt.graphitiExternal, tt.graphitiIns,
			tt.langfuseConnected, tt.langfuseExternal, tt.langfuseIns,
			tt.obsConnected, tt.obsExternal, tt.obsIns)
		if got != tt.want {
			t.Errorf("%s: %d components, want %d", tt.name, got, tt.want)
		}
	}
}

func TestHelpers_CalculateRequiredDiskGB_ReservesForWorkerImagesFirst(t *testing.T) {
	tests := []struct {
		name              string
		workerImageExists bool
		localComponents   int
		want              float64
	}{
		{"no worker image yet", false, 2, 25},
		{"two local components to install", true, 2, 14},
		{"nothing to install", true, 0, 5},
	}

	for _, tt := range tests {
		if got := calculateRequiredDiskGB(tt.workerImageExists, tt.localComponents); got != tt.want {
			t.Errorf("%s: %v GB, want %v", tt.name, got, tt.want)
		}
	}
}

func TestHelpers_CheckDiskSpaceWithContext_PassesAnInstalledStackWhenTheHostHasRoom(t *testing.T) {
	check := map[string]func(context.Context, float64) bool{
		"darwin": checkDarwinDiskSpace, "linux": checkLinuxDiskSpace,
	}[runtime.GOOS]
	if check == nil {
		t.Skipf("no disk check on %s", runtime.GOOS)
	}
	// The platform comparison, both ways, with requirements no host can miss or meet.
	if !check(t.Context(), 0) || check(t.Context(), 1e9) {
		t.Errorf("0 GB passes: %t, a billion GB passes: %t; want only the first", check(t.Context(), 0), check(t.Context(), 1e9))
	}
	// Host-relative and one way only: everything installed needs 5 GB, and the margin keeps a host near it from flaking.
	free := getAvailableDiskGB(t.Context())
	if got := checkDiskSpaceWithContext(t.Context(), true, true,
		true, false, true, true, false, true, true, false, true); free > 5.5 && !got {
		t.Errorf("%.1f GB free refuses an installation that needs 5", free)
	}
}

func TestHelpers_CheckImageExists_FindsTheImageUnderItsFullReference(t *testing.T) {
	docker := checkerFakeDocker(t, map[string]http.HandlerFunc{"GET /images/json": checkerJSON(http.StatusOK,
		`[{"Id":"sha256:`+strings.Repeat("a", 64)+`","RepoTags":["vxcontrol/kali-linux:latest"]}]`)})
	tests := []struct {
		name  string
		cli   *client.Client
		image string
		want  bool
	}{
		{"the tag the daemon lists", docker, "vxcontrol/kali-linux:latest", true},
		{"the implicit latest tag", docker, "vxcontrol/kali-linux", true},
		{"an image the daemon does not have", docker, "debian:latest", false},
		{"no daemon", nil, "vxcontrol/kali-linux:latest", false},
	}

	for _, tt := range tests {
		if got := checkImageExists(t.Context(), tt.cli, tt.image); got != tt.want {
			t.Errorf("%s: checkImageExists(%q) = %t, want %t", tt.name, tt.image, got, tt.want)
		}
	}
}

func TestHelpers_CheckVolumesExist_MatchesANameOrItsComposeProjectPrefix(t *testing.T) {
	tests := []struct {
		name     string
		existing []string // nil makes the daemon refuse the listing
		search   []string
		want     bool
	}{
		{"an exact name", []string{"pentagi-data", "other-volume"}, []string{"pentagi-data"}, true},
		{"a compose project prefix", []string{"pentagi_pentagi-data", "pentagi_pentagi-ssl"}, []string{"pentagi-data"}, true},
		{"no such volume", []string{"other-volume", "another-volume"}, []string{"pentagi-data"}, false},
		{"a longer name is not a match", []string{"pentagi-data-backup", "my-pentagi-data"}, []string{"pentagi-data"}, false},
		{"any of the searched names", []string{"other-volume", "langfuse-data"}, []string{"pentagi-data", "langfuse-data"}, true},
		{"no volumes at all", []string{}, []string{"pentagi-data"}, false},
		{"nothing searched", []string{"pentagi-data"}, []string{}, false},
		{"the daemon refuses the listing", nil, []string{"pentagi-data"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listing := checkerJSON(http.StatusInternalServerError, `{"message":"daemon error"}`)
			if tt.existing != nil {
				volumes := make([]map[string]string, 0, len(tt.existing))
				for _, name := range tt.existing {
					volumes = append(volumes, map[string]string{"Name": name})
				}
				body, _ := json.Marshal(map[string]any{"Volumes": volumes})
				listing = checkerJSON(http.StatusOK, string(body))
			}
			cli := checkerFakeDocker(t, map[string]http.HandlerFunc{"GET /volumes": listing})

			if got := checkVolumesExist(t.Context(), cli, tt.search); got != tt.want {
				t.Errorf("checkVolumesExist(%q) = %t, want %t", tt.search, got, tt.want)
			}
		})
	}
	if checkVolumesExist(t.Context(), nil, []string{"pentagi-data"}) {
		t.Error("no daemon reports a volume")
	}
}
