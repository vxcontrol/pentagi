package checker

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"pentagi/cmd/installer/cloud"
	"pentagi/cmd/installer/loader"

	"github.com/vxcontrol/cloud/models"
)

// composeContainerNames reads every container_name declared across the compose files.
func composeContainerNames(t *testing.T) map[string]string {
	t.Helper()

	names := map[string]string{}
	for _, path := range composeFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, line := range strings.Split(string(content), "\n") {
			if name, found := strings.CutPrefix(strings.TrimSpace(line), "container_name:"); found {
				names[strings.TrimSpace(name)] = filepath.Base(path)
			}
		}
	}
	if len(names) == 0 {
		t.Fatal("no container names found — the compose files moved or changed shape")
	}
	return names
}

// inventoryContainers is the component-to-container mapping as the gathering holds it.
func inventoryContainers(t *testing.T) map[models.ComponentType]string {
	t.Helper()

	got := map[models.ComponentType]string{}
	for _, group := range [][]imageComponent{
		pentagiImageComponents, graphitiImageComponents,
		langfuseImageComponents, observabilityImageComponents,
	} {
		for _, source := range group {
			if previous, seen := got[source.Component]; seen {
				t.Errorf("component %s is listed twice (%s and %s)", source.Component, previous, source.Container)
			}
			got[source.Component] = source.Container
		}
	}
	return got
}

func TestComponents_ImageComponent_MatchesTheComposeContainersBothWays(t *testing.T) {
	declared := composeContainerNames(t)
	observed := map[string]models.ComponentType{}
	for component, container := range inventoryContainers(t) {
		observed[container] = component
		if _, exists := declared[container]; !exists {
			t.Errorf("component %s is observed through container %q, which no compose file declares", component, container)
		}
	}
	for container, file := range declared {
		if _, reported := observed[container]; !reported {
			t.Errorf("container %q in %s is not reported as a component, so it is never checked for updates", container, file)
		}
	}
}

func TestComponents_ImageComponent_CoversEveryImageComponentOfADeployedStack(t *testing.T) {
	deployed := map[models.ProductStack]bool{
		models.ProductStackPentagi:       true,
		models.ProductStackGraphiti:      true,
		models.ProductStackLangfuse:      true,
		models.ProductStackObservability: true,
	}

	reported := inventoryContainers(t)
	for _, component := range models.AllComponentTypes {
		if !component.IsImageComponent() || !deployed[component.GetProductStack()] {
			continue
		}
		if _, ok := reported[component]; !ok {
			t.Errorf("component %s of the %s stack is an image, but no source names its container",
				component, component.GetProductStack())
		}
	}
}

// Which list an artefact goes in is decided by the builder that constructs it; the contract decides it by kind.
func TestComponents_BuildsEachComponentIntoTheListForItsKind(t *testing.T) {
	for component, container := range inventoryContainers(t) {
		if !component.IsImageComponent() {
			t.Errorf("%s is built as an image component (container %q) but is delivered as a %s",
				component, container, component.ArtifactKind())
		}
	}
	if !models.ComponentTypeWorker.IsImageComponent() {
		t.Errorf("the worker is built as an image component but is delivered as a %s", models.ComponentTypeWorker.ArtifactKind())
	}
	for _, component := range []models.ComponentType{models.ComponentTypeInstaller, models.ComponentTypeJaegerClickhouse} {
		if !component.IsFileComponent() {
			t.Errorf("%s is built as a file component but is delivered as an %s", component, component.ArtifactKind())
		}
	}
}

func TestComponents_JaegerPluginDir_MatchesWhatTheContainerLoads(t *testing.T) {
	content, err := os.ReadFile("../files/fs/docker-compose-observability.yml")
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}
	compose := string(content)

	if !strings.Contains(compose, "./observability/jaeger:/etc/jaeger") {
		t.Fatal("the jaeger service no longer mounts ./observability/jaeger at /etc/jaeger")
	}
	if !strings.Contains(compose, "/etc/jaeger/bin/jaeger-clickhouse-linux-$$ARCH") {
		t.Fatal("the jaeger service no longer loads its plugin from /etc/jaeger/bin")
	}
	if JaegerPluginDir != "observability/jaeger/bin" {
		t.Errorf("plugin directory = %q, want observability/jaeger/bin", JaegerPluginDir)
	}
	for _, name := range JaegerPluginBinaries {
		if _, err := os.Stat(filepath.Join("../files/fs", JaegerPluginDir, name)); err != nil {
			t.Errorf("plugin %s is not where the installer ships it: %v", name, err)
		}
	}
}

func TestComponents_MaxReportedComponents_FitsTheWholeInventory(t *testing.T) {
	total := len(pentagiImageComponents) + len(graphitiImageComponents) +
		len(langfuseImageComponents) + len(observabilityImageComponents) +
		1 + // worker
		1 + // installer
		len(JaegerPluginBinaries)

	if total > maxReportedComponents {
		t.Fatalf("the inventory is %d components, above the %d a request carries", total, maxReportedComponents)
	}
}

func TestComponents_ImageComponents_DescribesOnlyTheImagesOfALocalStack(t *testing.T) {
	hash := strings.Repeat("a", 64)
	container := func(running bool, image, reference string) http.HandlerFunc {
		return checkerJSON(http.StatusOK, fmt.Sprintf(`{"State":{"Running":%t},"Image":%q,"Config":{"Image":%q}}`,
			running, image, reference))
	}
	docker := checkerFakeDocker(t, map[string]http.HandlerFunc{
		"GET /containers/langfuse-worker/json":     container(true, "sha256:"+hash, "langfuse/langfuse-worker:3"),
		"GET /containers/langfuse-web/json":        container(false, "sha256:"+hash, "docker.io/langfuse/langfuse:3"),
		"GET /containers/langfuse-postgres/json":   container(true, "sha256:gone", "postgres:16"),
		"GET /containers/langfuse-clickhouse/json": container(true, "sha256:s390x", "clickhouse/clickhouse-server:24"),
		"GET /containers/langfuse-redis/json":      container(true, "sha256:"+hash, ""),
		"GET /images/sha256:" + hash + "/json":     checkerJSON(http.StatusOK, `{"Id":"sha256:`+hash+`","Os":"linux","Architecture":"amd64"}`),
		"GET /images/sha256:s390x/json":            checkerJSON(http.StatusOK, `{"Id":"sha256:s390x","Os":"linux","Architecture":"s390x"}`),
	})
	// Minio is not deployed; postgres's image, clickhouse's platform and redis's reference cannot be described.
	described := []models.ImageComponentInfo{
		{Component: models.ComponentTypeLangfuseWorker, Status: models.ComponentStatusRunning, OS: models.OSTypeLinux,
			Arch: models.ArchTypeAMD64, Repository: "langfuse/langfuse-worker", Tag: "3", ImageHash: &hash},
		{Component: models.ComponentTypeLangfuseWeb, Status: models.ComponentStatusInstalled, OS: models.OSTypeLinux,
			Arch: models.ArchTypeAMD64, Repository: "langfuse/langfuse", Tag: "3", ImageHash: &hash},
	}
	tests := []struct {
		name string
		gate stackGate
		want []models.ImageComponentInfo
	}{
		{"a local stack reports every image it can describe", stackGate{connected: true}, described},
		{"a stack running elsewhere reports nothing", stackGate{connected: true, external: true}, nil},
		{"an unused stack reports nothing", stackGate{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &defaultCheckHandler{mx: &sync.Mutex{}, dockerClient: docker}

			got := handler.imageComponents(t.Context(), langfuseImageComponents, tt.gate)

			if len(got) != len(tt.want) || (len(got) > 0 && !reflect.DeepEqual(got, tt.want)) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestComponents_ImageReference_NamesEveryReferenceButAnAbsentOne(t *testing.T) {
	tests := []struct {
		reference, wantRepository, wantTag string
		wantOK                             bool
	}{
		{"vxcontrol/pentagi", "vxcontrol/pentagi", "latest", true},
		{":latest", ":latest", "latest", true},
		{"", "", "", false},
	}

	for _, tt := range tests {
		repository, tag, ok := imageReference(tt.reference)
		if ok != tt.wantOK || repository != tt.wantRepository || tag != tt.wantTag {
			t.Errorf("imageReference(%q) = %q, %q, %t; want %q, %q, %t",
				tt.reference, repository, tag, ok, tt.wantRepository, tt.wantTag, tt.wantOK)
			continue
		}
		component := models.ImageComponentInfo{
			Component: models.ComponentTypePentagi, Status: models.ComponentStatusRunning,
			OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64, Repository: repository, Tag: tag,
		}
		if err := component.Valid(); ok && err != nil {
			t.Errorf("a component built from %q would be rejected: %v", tt.reference, err)
		}
	}
}

func TestComponents_NormalizeDigest_KeepsOnlyABareLowercaseSHA256(t *testing.T) {
	valid := strings.Repeat("ab", 32)
	tests := []struct {
		name, raw, want string
	}{
		{"the daemon prefix is stripped", "sha256:" + valid, valid},
		{"already bare", valid, valid},
		{"uppercase is lowered", strings.ToUpper(valid), valid},
		{"surrounding whitespace", "  sha256:" + valid + "  ", valid},
		{"truncated digests are refused", valid[:12], ""},
		{"non-hex is refused", strings.Repeat("z", 64), ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		if got := normalizeDigest(tt.raw); got != tt.want {
			t.Errorf("%s: normalizeDigest(%q) = %q, want %q", tt.name, tt.raw, got, tt.want)
		}
	}
}

func TestComponents_JaegerPluginComponents_ReportsBothArchitecturesOfALocalPlugin(t *testing.T) {
	envDir := t.TempDir()
	pluginDir := filepath.Join(envDir, "observability", "jaeger", "bin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	sums := map[string]string{}
	for _, name := range []string{"jaeger-clickhouse-linux-amd64", "jaeger-clickhouse-linux-arm64"} {
		content := []byte("plugin " + name)
		if err := os.WriteFile(filepath.Join(pluginDir, name), content, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		sum := sha256.Sum256(content)
		sums[name] = hex.EncodeToString(sum[:])
	}
	handler := &defaultCheckHandler{
		mx:       &sync.Mutex{},
		appState: &mockState{vars: map[string]loader.EnvVar{}, envPath: filepath.Join(envDir, ".env")},
	}
	type archHash struct {
		arch models.ArchType
		hash string
	}
	tests := []struct {
		name string
		gate stackGate
		want []archHash
	}{
		{"a local observability stack", stackGate{connected: true}, []archHash{
			{"amd64", sums["jaeger-clickhouse-linux-amd64"]},
			{"arm64", sums["jaeger-clickhouse-linux-arm64"]},
		}},
		{"an unused observability stack", stackGate{}, nil},
		{"observability running elsewhere", stackGate{connected: true, external: true}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []archHash
			for _, component := range handler.jaegerPluginComponents(tt.gate) {
				if component.Component != models.ComponentTypeJaegerClickhouse || component.OS != models.OSTypeLinux ||
					component.FileHash == nil || component.Valid() != nil {
					t.Fatalf("%s: %+v is not a linux jaeger plugin the contract accepts", component.Arch, component)
				}
				got = append(got, archHash{component.Arch, *component.FileHash})
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("plugin components = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponents_InstallerComponent_ReportsAContractVersion(t *testing.T) {
	handler := &defaultCheckHandler{mx: &sync.Mutex{}}

	component := handler.installerComponent()
	if component == nil {
		t.Skip("this platform has no contract name, which the gathering handles by skipping")
	}
	if err := component.Valid(); err != nil {
		t.Fatalf("the installer component would be rejected: %v", err)
	}
	if component.Version == nil {
		t.Fatal("the installer reports no version")
	}
	if component.FileHash != nil && len(*component.FileHash) != 64 {
		t.Errorf("file hash is %d characters, want 64", len(*component.FileHash))
	}
}

func TestComponents_EnforceComponentLimit_KeepsTheImagesAndTrimsTheFiles(t *testing.T) {
	handler := &defaultCheckHandler{mx: &sync.Mutex{}}
	oversized := reportedComponents{}
	for range models.MaxReportedComponents {
		oversized.images = append(oversized.images, models.ImageComponentInfo{Component: models.ComponentTypePentagi})
	}
	for range 3 {
		oversized.files = append(oversized.files, models.FileComponentInfo{Component: models.ComponentTypeInstaller})
	}

	trimmed := handler.enforceComponentLimit(oversized)

	if len(trimmed.images) != models.MaxReportedComponents || len(trimmed.files) != 0 {
		t.Errorf("kept %d images and %d files, want the contract's %d images and no files",
			len(trimmed.images), len(trimmed.files), models.MaxReportedComponents)
	}
}

func TestComponents_GatherComponents_BuildsARequestTheContractAccepts(t *testing.T) {
	handler := &defaultCheckHandler{
		mx:       &sync.Mutex{},
		appState: &mockState{vars: map[string]loader.EnvVar{}, envPath: t.TempDir() + "/.env"},
	}
	result := &CheckResult{GraphitiConnected: true, LangfuseConnected: true, ObservabilityConnected: true}

	components := handler.gatherComponents(t.Context(), result)

	request := models.CheckUpdatesRequest{
		InstallerVersion: cloud.NormalizeVersion(InstallerVersion),
		InstallerOS:      models.OSTypeLinux,
		InstallerArch:    models.ArchTypeAMD64,
		Strategy:         models.UpdateStrategyStable,
		Images:           components.images,
		Files:            components.files,
	}
	if err := request.Valid(); err != nil {
		t.Fatalf("the gathered request would be rejected: %v", err)
	}
}
