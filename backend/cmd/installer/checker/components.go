package checker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"pentagi/cmd/installer/cloud"
	"pentagi/cmd/installer/wizard/logger"

	"github.com/moby/moby/client"
	"github.com/vxcontrol/cloud/models"
)

// This file turns a running installation into the component list the update server is
// asked about. The list has to be COMPLETE: the server answers only about what it was
// told, so a component left out comes back as "nothing to report" — which reads exactly
// like "up to date" and is the more dangerous of the two mistakes.

// maxReportedComponents is the server-side ceiling on one request, and it IS the
// contract's constant rather than a number that agrees with it today. A local literal
// can only be wrong in two ways, and both are silent: too high and the service refuses
// the whole request — losing the answer for every component, not just the ones past the
// bound — too low and the tail is trimmed here while the service would have taken it.
//
// The inventory is well under it, but the list grows with every stack, so the limit is
// enforced before sending and what gets dropped is written to the log: a silently
// truncated list would report the remaining components as complete.
const maxReportedComponents = models.MaxReportedComponents

// JaegerPluginDir holds the Jaeger storage plugin, relative to the directory the
// environment file lives in.
//
// Exported because the processor writes into this directory when it downloads a newer
// plugin: the path, the file names and the digests reported for them have to be one
// table, or an update writes a file nothing reads.
const JaegerPluginDir = "observability/jaeger/bin"

// JaegerPluginBinaries maps the architecture a plugin file is built for to its name.
//
// Both files are on disk regardless of the host: the choice happens inside the Jaeger
// container, by the architecture of the Docker VM rather than of this machine. So both are
// reported, as two components.
var JaegerPluginBinaries = map[models.ArchType]string{
	models.ArchTypeAMD64: "jaeger-clickhouse-linux-amd64",
	models.ArchTypeARM64: "jaeger-clickhouse-linux-arm64",
}

// imageComponent names the container an image component is observed through.
//
// The table is explicit because nothing here can be derived. Four of the sixteen observed
// components run in a container whose name is not the component's: postgres, clickhouse,
// redis and minio all carry the `langfuse-` prefix of the stack that owns them. Guessing
// the container from the component would silently drop those four — a name matching no
// container is indistinguishable from a container that is not deployed — so the mapping
// lives in exactly one place, and TestComponents_ImageComponent_MatchesTheComposeContainersBothWays
// keeps it honest against the compose files.
type imageComponent struct {
	Component models.ComponentType
	Container string
}

// The four groups below are also the order components are reported in, which is the order
// the ceiling above cuts from: the stacks whose updates matter most come first.
var (
	pentagiImageComponents = []imageComponent{
		{models.ComponentTypePentagi, PentagiContainerName},
		{models.ComponentTypeScraper, ScraperContainerName},
		{models.ComponentTypePgvector, PgvectorContainerName},
		// Last in its group on purpose: the ceiling cuts from the tail, and an
		// exporter is the one component of this stack an installation can live
		// without hearing about.
		{models.ComponentTypePgexporter, PgexporterContainerName},
	}
	graphitiImageComponents = []imageComponent{
		{models.ComponentTypeGraphiti, GraphitiContainerName},
		{models.ComponentTypeNeo4j, Neo4jContainerName},
	}
	langfuseImageComponents = []imageComponent{
		{models.ComponentTypeLangfuseWorker, LangfuseWorkerContainerName},
		{models.ComponentTypeLangfuseWeb, LangfuseWebContainerName},
		{models.ComponentTypePostgres, LangfusePostgresContainerName},
		{models.ComponentTypeClickhouse, LangfuseClickhouseContainerName},
		{models.ComponentTypeRedis, LangfuseRedisContainerName},
		{models.ComponentTypeMinio, LangfuseMinioContainerName},
	}
	observabilityImageComponents = []imageComponent{
		{models.ComponentTypeGrafana, GrafanaContainerName},
		{models.ComponentTypeOtel, OpenTelemetryContainerName},
		{models.ComponentTypeVictoriametrics, VictoriaMetricsContainerName},
		{models.ComponentTypeLoki, LokiContainerName},
		{models.ComponentTypeJaeger, JaegerContainerName},
		// clickstore is the ClickHouse server Jaeger stores traces in, and it is a
		// different component from `clickhouse`, which is Langfuse's. The two run
		// the same image and belong to different stacks.
		{models.ComponentTypeClickstore, ClickstoreContainerName},
		{models.ComponentTypeCadvisor, CadvisorContainerName},
		{models.ComponentTypeNodeExporter, NodeExporterContainerName},
	}
)

// stackGate decides how a stack's components are reported.
//
// A stack pointing at an external address has no local images to describe, so its
// components are reported as connected and carry no image data. A stack that is not
// configured at all is not reported: "unused" would spend one of the thirty slots saying
// nothing.
type stackGate struct {
	connected bool
	external  bool
}

func (g stackGate) skip() bool       { return !g.connected }
func (g stackGate) isExternal() bool { return g.external }

// gatherComponents builds the component list for an update check.
//
// Anything that cannot be described completely is left out rather than guessed at: a
// component with an unknown architecture, or an image the daemon will not describe, would
// fail the contract's validation and take the whole request down with it — costing the
// answer for every other component too.
// reportedComponents is what one update check says about this machine. Images and
// files are separate all the way through: an image is identified by a registry
// reference and a digest, a file by a version and a hash, and no artefact is both.
type reportedComponents struct {
	images []models.ImageComponentInfo
	files  []models.FileComponentInfo
}

func (r reportedComponents) total() int { return len(r.images) + len(r.files) }

func (h *defaultCheckHandler) gatherComponents(ctx context.Context, c *CheckResult) reportedComponents {
	var reported reportedComponents

	// The pentagi stack is always reported: it is the product itself, with no external
	// alternative to point at.
	reported.images = append(reported.images,
		h.imageComponents(ctx, pentagiImageComponents, stackGate{connected: true})...)

	if worker := h.workerComponent(ctx, c); worker != nil {
		reported.images = append(reported.images, *worker)
	}
	if installer := h.installerComponent(); installer != nil {
		reported.files = append(reported.files, *installer)
	}

	reported.images = append(reported.images, h.imageComponents(ctx, graphitiImageComponents,
		stackGate{connected: c.GraphitiConnected, external: c.GraphitiExternal})...)
	reported.images = append(reported.images, h.imageComponents(ctx, langfuseImageComponents,
		stackGate{connected: c.LangfuseConnected, external: c.LangfuseExternal})...)

	observability := stackGate{connected: c.ObservabilityConnected, external: c.ObservabilityExternal}
	reported.images = append(reported.images, h.imageComponents(ctx, observabilityImageComponents, observability)...)
	reported.files = append(reported.files, h.jaegerPluginComponents(observability)...)

	return h.enforceComponentLimit(reported)
}

// enforceComponentLimit trims the report to what the server accepts, naming what it
// dropped.
//
// The limit is on the SUM of both lists — that is how the server counts it, and
// bounding each separately would build a request the server rejects whole.
func (h *defaultCheckHandler) enforceComponentLimit(reported reportedComponents) reportedComponents {
	if reported.total() <= maxReportedComponents {
		return reported
	}

	var dropped []string
	// Files are trimmed first only if images alone already fill the budget; otherwise
	// the tail of the files is what goes, so the images — which are the product — keep
	// their slots.
	if len(reported.images) > maxReportedComponents {
		for _, component := range reported.images[maxReportedComponents:] {
			dropped = append(dropped, component.Component.String())
		}
		for _, component := range reported.files {
			dropped = append(dropped, component.Component.String())
		}
		reported.images = reported.images[:maxReportedComponents]
		reported.files = nil
	} else {
		room := maxReportedComponents - len(reported.images)
		for _, component := range reported.files[room:] {
			dropped = append(dropped, component.Component.String())
		}
		reported.files = reported.files[:room]
	}

	logger.Warnf("update check covers %d of %d components, no answer for: %s",
		maxReportedComponents, maxReportedComponents+len(dropped), strings.Join(dropped, ", "))

	return reported
}

// imageComponents describes one stack's images.
func (h *defaultCheckHandler) imageComponents(
	ctx context.Context, sources []imageComponent, gate stackGate,
) []models.ImageComponentInfo {
	if gate.skip() {
		return nil
	}

	components := make([]models.ImageComponentInfo, 0, len(sources))
	for _, source := range sources {
		if gate.isExternal() {
			// An external stack runs somewhere we cannot inspect. There is no reference
			// to report and therefore nothing the server can resolve, so the component
			// is left out rather than sent as an image without an image.
			continue
		}

		if component := h.describeContainerImage(ctx, source); component != nil {
			components = append(components, *component)
		}
	}

	return components
}

// describeContainerImage inspects one container and the image behind it.
//
// The image, not the host, decides os and arch: a container image is linux even on a
// darwin host, and on Apple silicon it may well be an amd64 image running under emulation.
// Taking either from runtime would describe the wrong artefact.
func (h *defaultCheckHandler) describeContainerImage(
	ctx context.Context, source imageComponent,
) *models.ImageComponentInfo {
	if h.dockerClient == nil {
		return nil
	}

	container, err := h.dockerClient.ContainerInspect(ctx, source.Container, client.ContainerInspectOptions{})
	if err != nil {
		// No container means the component is not deployed here, which is not worth a
		// slot in the request.
		return nil
	}

	status := models.ComponentStatusInstalled
	if container.Container.State != nil && container.Container.State.Running {
		status = models.ComponentStatusRunning
	}

	reference := ""
	if container.Container.Config != nil {
		reference = container.Container.Config.Image
	}

	image, err := h.dockerClient.ImageInspect(ctx, container.Container.Image)
	if err != nil {
		logger.Warnf("component %s: image %s cannot be inspected: %v",
			source.Component, container.Container.Image, err)
		return nil
	}

	osType := models.OSType(strings.ToLower(image.Os))
	archType := models.ArchType(strings.ToLower(image.Architecture))
	if osType.Valid() != nil || archType.Valid() != nil {
		// A platform the contract has no name for would fail validation and take the
		// whole request with it, so this one component is dropped instead.
		logger.Warnf("component %s: unsupported image platform %s/%s, not reported",
			source.Component, image.Os, image.Architecture)
		return nil
	}

	repository, tag, ok := imageReference(reference)
	if !ok {
		logger.Warnf("component %s: container image reference %q cannot be parsed, not reported",
			source.Component, reference)
		return nil
	}

	component := models.ImageComponentInfo{
		Component:  source.Component,
		Status:     status,
		OS:         osType,
		Arch:       archType,
		Repository: repository,
		Tag:        tag,
	}
	if digest := normalizeDigest(image.ID); digest != "" {
		component.ImageHash = &digest
	}

	return &component
}

// imageReference splits a container's image reference into the pair the server
// matches on — repository as compose spells it, with the registry prefix and the
// implicit "library/" removed — and reports whether it can be named at all.
//
// A reference that cannot be named has no channel to resolve, so the component is
// dropped rather than sent incomplete: an update is the newest image UNDER a
// reference, and the contract requires both halves for exactly that reason.
//
// There is no tag default here because parseImageRef already applies the one
// `docker pull` applies, and never returns an empty tag or an empty name for a
// non-empty reference — it falls back to the whole original. So the only
// reference this rejects is the absent one, which is a container whose config
// carries no image at all.
func imageReference(reference string) (repository, tag string, ok bool) {
	info := parseImageRef(reference, "")
	if info == nil || info.Name == "" {
		return "", "", false
	}
	return info.Name, info.Tag, true
}

// workerComponent describes the pentest worker image.
//
// It has no container and no compose service: the image is pulled into the worker
// environment on demand. What is reported is what is actually there — substituting the
// built-in default would claim an image the machine may not have.
func (h *defaultCheckHandler) workerComponent(ctx context.Context, c *CheckResult) *models.ImageComponentInfo {
	if h.workerClient == nil || !c.WorkerImageExists {
		return nil
	}

	configured := getEnvVar(h.appState, "DOCKER_DEFAULT_IMAGE_FOR_PENTEST", DefaultImageForPentest)
	info := getImageInfo(ctx, h.workerClient, configured)
	if info == nil || info.Name == "" {
		return nil
	}

	component := models.ImageComponentInfo{
		Component: models.ComponentTypeWorker,
		// No container exists for it, so the image being present is the whole story.
		Status: models.ComponentStatusInstalled,
		OS:     models.OSTypeLinux,
		Arch:   models.ArchType(runtime.GOARCH),
	}

	// The image itself is authoritative about its platform when the daemon will describe
	// it; the host architecture is only a fallback for a daemon that will not.
	if image, err := h.workerClient.ImageInspect(ctx, info.Name+":"+info.Tag); err == nil {
		osType := models.OSType(strings.ToLower(image.Os))
		archType := models.ArchType(strings.ToLower(image.Architecture))
		if osType.Valid() == nil && archType.Valid() == nil {
			component.OS, component.Arch = osType, archType
		}
		if digest := normalizeDigest(image.ID); digest != "" {
			component.ImageHash = &digest
		}
	} else if digest := normalizeDigest(info.Hash); digest != "" {
		component.ImageHash = &digest
	}

	if component.Arch.Valid() != nil {
		return nil
	}

	// getImageInfo answers through parseImageRef, so the tag is already the one
	// `docker pull` would apply.
	component.Repository = info.Name
	component.Tag = info.Tag

	return &component
}

// installerComponent describes the running installer binary — the one component that is a
// file on this host rather than an image, and so the only one whose os and arch are the
// host's own.
func (h *defaultCheckHandler) installerComponent() *models.FileComponentInfo {
	osType := models.OSType(runtime.GOOS)
	archType := models.ArchType(runtime.GOARCH)
	if osType.Valid() != nil || archType.Valid() != nil {
		logger.Warnf("installer runs on unsupported platform %s/%s, not reported",
			runtime.GOOS, runtime.GOARCH)
		return nil
	}

	version := cloud.NormalizeVersion(InstallerVersion)
	component := models.FileComponentInfo{
		Component: models.ComponentTypeInstaller,
		Status:    models.ComponentStatusInstalled,
		OS:        osType,
		Arch:      archType,
		Version:   &version,
	}

	// A hash of ourselves is a nice-to-have: the server identifies installer builds by
	// version. Failing to read our own path is no reason to drop the component.
	if path, err := os.Executable(); err == nil {
		if digest, err := fileDigest(path); err == nil {
			component.FileHash = &digest
		}
	}

	return &component
}

// jaegerPluginComponents describes the Jaeger storage plugin, which ships as one file per
// architecture. Both are reported: which one runs is decided inside the Jaeger container.
func (h *defaultCheckHandler) jaegerPluginComponents(gate stackGate) []models.FileComponentInfo {
	if gate.skip() || gate.isExternal() || h.appState == nil {
		return nil
	}

	base := filepath.Join(filepath.Dir(h.appState.GetEnvPath()), JaegerPluginDir)

	components := make([]models.FileComponentInfo, 0, len(JaegerPluginBinaries))
	// Fixed order, so two runs on the same machine produce the same request.
	for _, arch := range []models.ArchType{models.ArchTypeAMD64, models.ArchTypeARM64} {
		digest, err := fileDigest(filepath.Join(base, JaegerPluginBinaries[arch]))
		if err != nil {
			continue
		}
		components = append(components, models.FileComponentInfo{
			Component: models.ComponentTypeJaegerClickhouse,
			Status:    models.ComponentStatusInstalled,
			// The plugin runs inside the Jaeger container, so it is a linux binary
			// whatever this host is.
			OS:       models.OSTypeLinux,
			Arch:     arch,
			FileHash: &digest,
		})
	}

	return components
}

// normalizeDigest converts a digest as the Docker daemon reports it into the form the
// contract accepts: bare lowercase hex, no "sha256:" prefix. An unrecognisable value
// becomes empty, so it is omitted rather than sent and rejected.
func normalizeDigest(raw string) string {
	digest := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(raw), "sha256:"))
	if len(digest) != 64 {
		return ""
	}
	for _, r := range digest {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return ""
		}
	}
	return digest
}

// fileDigest returns the sha256 of a file in the same bare-hex form.
func fileDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(digest.Sum(nil)), nil
}
