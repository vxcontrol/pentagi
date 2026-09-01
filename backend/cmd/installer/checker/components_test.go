package checker

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"pentagi/cmd/installer/loader"

	"github.com/vxcontrol/cloud/models"
)

// composeFiles are the deployment descriptors the installer ships and extracts. The tests
// below read them rather than a copy, so a container renamed in a compose file fails here
// instead of silently making a component unobservable.
var composeFiles = []string{
	"../files/fs/docker-compose.yml",
	"../files/fs/docker-compose-graphiti.yml",
	"../files/fs/docker-compose-langfuse.yml",
	"../files/fs/docker-compose-observability.yml",
}

// containersNotReported are the containers deliberately left out of update checks.
//
// It is empty, and keeping it empty is the point. It once held pgexporter, node-exporter,
// cadvisor and clickstore under the reason "no component type in the contract" — which was
// simply untrue: all four are declared component types and the server has always been able
// to answer for them. The exclusion list turned a wrong belief into an enforced rule, so
// TestEveryComposeContainerIsAccountedFor passed while four running containers could never
// be offered an update. A CVE in cAdvisor or in the ClickHouse behind Jaeger would have
// reached no installation.
//
// So an entry here now needs a reason that survives being checked against
// models.AllComponentTypes, and TestEveryImageComponentOfADeployedStackIsReported below
// closes the same hole from the other side.
var containersNotReported = map[string]string{}

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
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "container_name:") {
				continue
			}
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "container_name:"))
			names[name] = filepath.Base(path)
		}
	}

	if len(names) == 0 {
		t.Fatal("no container names found — the compose files moved or changed shape")
	}

	return names
}

// inventoryContainers is the mapping as the gathering actually holds it.
func inventoryContainers(t *testing.T) map[models.ComponentType]string {
	t.Helper()

	got := map[models.ComponentType]string{}
	for _, group := range [][]imageComponent{
		pentagiImageComponents, graphitiImageComponents,
		langfuseImageComponents, observabilityImageComponents,
	} {
		for _, source := range group {
			if previous, seen := got[source.Component]; seen {
				t.Errorf("component %s is listed twice (%s and %s)",
					source.Component, previous, source.Container)
			}
			got[source.Component] = source.Container
		}
	}

	return got
}

// TestEveryObservedContainerExistsInCompose checks the mapping against the compose files
// themselves.
//
// Four components run in a container whose name is not their own — postgres, clickhouse,
// redis and minio all carry the `langfuse-` prefix of the stack that owns them — so the
// mapping cannot be derived and has to be written down. Writing it down by hand is exactly
// how it goes stale, and the failure is invisible: a container name that matches nothing
// looks identical to a container that is not deployed, so the component is quietly dropped
// from the request and the server never reports on it.
func TestEveryObservedContainerExistsInCompose(t *testing.T) {
	declared := composeContainerNames(t)

	for component, container := range inventoryContainers(t) {
		if _, exists := declared[container]; !exists {
			t.Errorf("component %s is observed through container %q, which no compose file declares",
				component, container)
		}
	}
}

// TestEveryComposeContainerIsAccountedFor is the other direction, and the one that catches
// growth: a service added to a compose file must either become a reported component or be
// listed as deliberately excluded. Without this, a new container simply never gets checked
// for updates and nothing anywhere says so.
func TestEveryComposeContainerIsAccountedFor(t *testing.T) {
	observed := map[string]models.ComponentType{}
	for component, container := range inventoryContainers(t) {
		observed[container] = component
	}

	for container, file := range composeContainerNames(t) {
		if _, reported := observed[container]; reported {
			continue
		}
		if _, excluded := containersNotReported[container]; excluded {
			continue
		}
		t.Errorf("container %q in %s is neither reported as a component nor listed as "+
			"deliberately excluded — it would never be checked for updates", container, file)
	}
}

// TestJaegerPluginPathMatchesWhatTheContainerLoads pins the location against the compose
// file that mounts it. The Jaeger service mounts ./observability/jaeger at /etc/jaeger and
// loads the plugin from /etc/jaeger/bin/jaeger-clickhouse-linux-$ARCH, so the host-side
// path is fixed by that pair — and hashing a file that is not the one running would report
// a plugin nobody uses.
func TestJaegerPluginPathMatchesWhatTheContainerLoads(t *testing.T) {
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

	// The mount plus the load path fix the host-side directory.
	if JaegerPluginDir != "observability/jaeger/bin" {
		t.Errorf("plugin directory = %q, want observability/jaeger/bin", JaegerPluginDir)
	}

	// Both files ship, because the architecture is chosen inside the container by uname.
	for _, name := range JaegerPluginBinaries {
		if _, err := os.Stat(filepath.Join("../files/fs", JaegerPluginDir, name)); err != nil {
			t.Errorf("plugin %s is not where the installer ships it: %v", name, err)
		}
	}
}

// TestInventoryFitsTheRequestCeiling: the whole observable inventory has to fit in one
// request, or some of it is never asked about. Sixteen images, plus the worker, the
// installer and the plugin's two architectures.
func TestInventoryFitsTheRequestCeiling(t *testing.T) {
	total := len(pentagiImageComponents) + len(graphitiImageComponents) +
		len(langfuseImageComponents) + len(observabilityImageComponents) +
		1 + // worker
		1 + // installer
		len(JaegerPluginBinaries)

	if total > maxReportedComponents {
		t.Fatalf("the inventory is %d components, above the %d a request carries: "+
			"some would be dropped and never reported on", total, maxReportedComponents)
	}

	// The ceiling has to BE the contract's, not a number that matches it today. Too
	// high and the service refuses the whole request, costing the answer for every
	// component; too low and the tail is trimmed here while the service would have
	// accepted it. Both are invisible from inside this package, because every other
	// test in this file measures against this same constant.
	if maxReportedComponents != models.MaxReportedComponents {
		t.Errorf("the installer caps a request at %d artefacts, the contract at %d",
			maxReportedComponents, models.MaxReportedComponents)
	}
}

// TestEveryReportedComponentIsInTheListForItsKind.
//
// Which list an artefact goes in is decided here by WHICH BUILDER constructs it
// — the image tables above versus installerComponent and jaegerPluginComponents
// — and nothing related that choice to the component itself. The delivery kind
// is a property of the component (models.ArtifactKind), the service validates
// the request against exactly that table, and a component built into the wrong
// list is now a rejected request rather than a silent misresolution.
//
// The service side rejects it at the wire; this catches it here, where the
// mistake would actually be made.
// TestEveryImageComponentOfADeployedStackIsReported walks the contract's vocabulary and
// demands a source for each entry, which is the direction the compose-file tests cannot
// see.
//
// Those tests compare the inventory against the compose files, so a component the contract
// declares but no source table mentions is invisible to them — the container is simply
// never looked at, and "not deployed" and "not looked at" produce identical requests. That
// is exactly how pgexporter, node-exporter, cadvisor and clickstore stayed unreportable
// while every existing test was green: four running containers the server knew about and
// was never told to answer for.
//
// Only the four stacks the installer deploys through compose are demanded here. The worker
// has its own gathering (no container exists for it) and engine and browser are not
// deployed by the installer at all, so a source for them would describe nothing.
func TestEveryImageComponentOfADeployedStackIsReported(t *testing.T) {
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
			t.Errorf("component %s belongs to the %s stack and is delivered as an image, but no "+
				"source names the container it runs in — it can never be offered an update",
				component, component.GetProductStack())
		}
	}
}

func TestEveryReportedComponentIsInTheListForItsKind(t *testing.T) {
	for component, container := range inventoryContainers(t) {
		if !component.IsImageComponent() {
			t.Errorf("%s is built as an image component (container %q) but is delivered as a %s",
				component, container, component.ArtifactKind())
		}
	}

	if !models.ComponentTypeWorker.IsImageComponent() {
		t.Errorf("the worker is built as an image component but is delivered as a %s",
			models.ComponentTypeWorker.ArtifactKind())
	}

	for _, component := range []models.ComponentType{
		models.ComponentTypeInstaller,
		models.ComponentTypeJaegerClickhouse,
	} {
		if !component.IsFileComponent() {
			t.Errorf("%s is built as a file component but is delivered as an %s",
				component, component.ArtifactKind())
		}
	}
}

// TestExternalStackIsNotReportedAsAnImage: a stack pointing somewhere we cannot inspect
// has no reference we can name, and an image component that names none cannot be resolved
// — the contract requires the repository and the tag precisely because an update is the
// newest image UNDER a reference.
//
// It used to be reported as `connected` with every image field empty. That worked while
// the fields were optional and is inexpressible now.
//
// What this costs is an analytics signal, not an answer: nothing was ever offered for an
// external stack, but its presence in the list is what raised has_langfuse /
// has_graphiti / has_observability in the cloud's view of this installation. Restoring
// that needs each component to carry its compose default reference, which is a separate
// piece of inventory work.
func TestExternalStackIsNotReportedAsAnImage(t *testing.T) {
	handler := &defaultCheckHandler{mx: &sync.Mutex{}}

	components := handler.imageComponents(t.Context(), langfuseImageComponents,
		stackGate{connected: true, external: true})

	if len(components) != 0 {
		t.Fatalf("got %d components for a stack running elsewhere, want none: %+v",
			len(components), components)
	}
}

// TestAnUnnameableReferenceIsNotReported.
//
// The contract requires an image to name the reference it follows, because an update
// IS the newest image under that reference. A container whose image reference cannot
// be parsed therefore has nothing that can be resolved, and reporting it with an empty
// repository would not fail quietly — it would fail the request's validation and cost
// the answer for every other component too.
//
// A tag is different: `docker pull` supplies one for an omitted tag, and so does this.
//
// Note what is NOT covered: the wiring from container inspection into this decision.
// describeContainerImage needs a Docker daemon and there is no fake for one here, so
// what is pinned is the decision itself.
func TestAnUnnameableReferenceIsNotReported(t *testing.T) {
	tests := []struct {
		reference      string
		wantOK         bool
		wantRepository string
		wantTag        string
	}{
		{"vxcontrol/pentagi:latest", true, "vxcontrol/pentagi", "latest"},
		{"vxcontrol/pentagi", true, "vxcontrol/pentagi", "latest"},
		{"docker.io/library/redis:7", true, "redis", "7"},
		// parseImageRef never returns an empty name for a non-empty reference: it
		// falls back to the whole original. So the only reference that cannot be
		// named is the absent one — which is the real case, a container whose
		// config carries no image at all.
		{"", false, "", ""},
		{":latest", true, ":latest", "latest"},
	}

	for _, tt := range tests {
		t.Run(tt.reference, func(t *testing.T) {
			repository, tag, ok := imageReference(tt.reference)
			if ok != tt.wantOK {
				t.Fatalf("ok = %t, want %t (repository %q, tag %q)", ok, tt.wantOK, repository, tag)
			}
			if !ok {
				return
			}
			if repository != tt.wantRepository || tag != tt.wantTag {
				t.Errorf("got %q:%q, want %q:%q", repository, tag, tt.wantRepository, tt.wantTag)
			}
			component := models.ImageComponentInfo{
				Component: models.ComponentTypePentagi, Status: models.ComponentStatusRunning,
				OS: models.OSTypeLinux, Arch: models.ArchTypeAMD64,
				Repository: repository, Tag: tag,
			}
			if err := component.Valid(); err != nil {
				t.Errorf("the component built from this reference would be rejected: %v", err)
			}
		})
	}
}

// TestUnusedStackIsNotReportedAtAll: a stack nobody configured has nothing to say, and one
// of the thirty slots is worth more than an entry saying so.
func TestUnusedStackIsNotReportedAtAll(t *testing.T) {
	handler := &defaultCheckHandler{mx: &sync.Mutex{}}

	components := handler.imageComponents(t.Context(), graphitiImageComponents,
		stackGate{connected: false})

	if len(components) != 0 {
		t.Errorf("an unconfigured stack contributed %d components", len(components))
	}
}

func TestNormalizeDigest(t *testing.T) {
	valid := strings.Repeat("ab", 32)

	tests := []struct {
		name string
		raw  string
		want string
	}{
		// The daemon always reports the prefix; the contract's validator rejects it,
		// and on this path a rejected field is answered as a server error rather than a
		// complaint about the field.
		{name: "the daemon prefix is stripped", raw: "sha256:" + valid, want: valid},
		{name: "already bare", raw: valid, want: valid},
		{name: "uppercase is lowered", raw: strings.ToUpper(valid), want: valid},
		{name: "surrounding whitespace", raw: "  sha256:" + valid + "  ", want: valid},

		// Anything unrecognisable is dropped rather than sent: an omitted digest means
		// "cannot verify", while a malformed one takes the whole request down.
		{name: "truncated digests are refused", raw: valid[:12], want: ""},
		{name: "non-hex is refused", raw: strings.Repeat("z", 64), want: ""},
		{name: "empty", raw: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDigest(tt.raw); got != tt.want {
				t.Errorf("normalizeDigest(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// TestJaegerPluginIsReportedForBothArchitectures: both files sit on disk whatever the host
// is, because which one runs is decided inside the Jaeger container by the architecture of
// the Docker VM. Reporting only the host's would leave the other unverified forever.
func TestJaegerPluginIsReportedForBothArchitectures(t *testing.T) {
	envDir := t.TempDir()
	pluginDir := filepath.Join(envDir, JaegerPluginDir)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for arch, name := range JaegerPluginBinaries {
		if err := os.WriteFile(filepath.Join(pluginDir, name), []byte("plugin for "+arch.String()), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	handler := &defaultCheckHandler{
		mx:       &sync.Mutex{},
		appState: &mockState{vars: map[string]loader.EnvVar{}, envPath: filepath.Join(envDir, ".env")},
	}

	components := handler.jaegerPluginComponents(stackGate{connected: true})

	if len(components) != 2 {
		t.Fatalf("got %d plugin components, want 2", len(components))
	}

	seen := map[models.ArchType]string{}
	for _, component := range components {
		if component.Component != models.ComponentTypeJaegerClickhouse {
			t.Errorf("component = %q, want %q", component.Component, models.ComponentTypeJaegerClickhouse)
		}
		// The plugin runs inside the container, so it is a linux binary whatever this
		// host is.
		if component.OS != models.OSTypeLinux {
			t.Errorf("%s: os = %q, want linux", component.Arch, component.OS)
		}
		if component.FileHash == nil {
			t.Fatalf("%s: no digest", component.Arch)
		}
		if err := component.Valid(); err != nil {
			t.Errorf("%s: would be rejected by the contract: %v", component.Arch, err)
		}
		seen[component.Arch] = *component.FileHash
	}

	if seen[models.ArchTypeAMD64] == seen[models.ArchTypeARM64] {
		t.Error("both architectures report the same digest — one file was hashed twice")
	}
}

// TestJaegerPluginIsSkippedWhenObservabilityIsNotOurs: the plugin serves the local Jaeger.
// With observability pointing elsewhere there is no local Jaeger for it to serve.
func TestJaegerPluginIsSkippedWhenObservabilityIsNotOurs(t *testing.T) {
	// The files have to be on disk, or their absence — not the gate — is what produces the
	// empty list, and this test would pass with the gate removed entirely.
	envDir := t.TempDir()
	pluginDir := filepath.Join(envDir, JaegerPluginDir)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for arch, name := range JaegerPluginBinaries {
		if err := os.WriteFile(filepath.Join(pluginDir, name), []byte("plugin for "+arch.String()), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	handler := &defaultCheckHandler{
		mx:       &sync.Mutex{},
		appState: &mockState{vars: map[string]loader.EnvVar{}, envPath: filepath.Join(envDir, ".env")},
	}

	for _, gate := range []stackGate{
		{connected: false},
		{connected: true, external: true},
	} {
		if components := handler.jaegerPluginComponents(gate); len(components) != 0 {
			t.Errorf("gate %+v: reported %d plugin components", gate, len(components))
		}
	}
}

// TestInstallerReportsItselfWithAContractVersion: development builds are named after their
// branch, which is not a version. Sending one gets the whole request refused — and with it
// the answer for every other component.
func TestInstallerReportsItselfWithAContractVersion(t *testing.T) {
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

// TestComponentLimitNamesWhatItDropped: a silently truncated list reports the components
// that survived as if they were the whole installation.
func TestComponentLimitNamesWhatItDropped(t *testing.T) {
	handler := &defaultCheckHandler{mx: &sync.Mutex{}}

	// Over the ceiling with BOTH kinds present: the limit is on the sum, because that
	// is how the server counts it. Bounding each list separately would build a request
	// the server rejects whole.
	oversized := reportedComponents{}
	for i := 0; i < maxReportedComponents; i++ {
		oversized.images = append(oversized.images,
			models.ImageComponentInfo{Component: models.ComponentTypePentagi})
	}
	for i := 0; i < 3; i++ {
		oversized.files = append(oversized.files,
			models.FileComponentInfo{Component: models.ComponentTypeInstaller})
	}

	trimmed := handler.enforceComponentLimit(oversized)

	if trimmed.total() != maxReportedComponents {
		t.Errorf("got %d components, want the ceiling of %d", trimmed.total(), maxReportedComponents)
	}
	if len(trimmed.images) != maxReportedComponents {
		t.Errorf("the images are the product and must keep their slots, got %d", len(trimmed.images))
	}
}
