package checker

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"time"

	"pentagi/cmd/installer/cloud"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/logger"

	"github.com/vxcontrol/cloud/models"
)

// StackUpdate is what the update server said about one product stack.
//
// HasUpdate is the stack-level verdict and the only thing the menu needs. Everything else
// is for the screen that explains the update before it is applied: which version is
// current, which is offered, and what the release says about the difference.
type StackUpdate struct {
	Stack          string `json:"stack" yaml:"stack"`
	HasUpdate      bool   `json:"has_update" yaml:"has_update"`
	CurrentVersion string `json:"current_version,omitempty" yaml:"current_version,omitempty"`
	LatestVersion  string `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`
	// CurrentVersionMixed says the stack's components came from DIFFERENT releases, so
	// CurrentVersion names the oldest of them rather than a version this installation as
	// a whole ever was. Legitimate — a component nobody rebuilt stays on its old release
	// — but "you are on 2.1.0" and "the oldest thing you have is from 2.1.0" are
	// different sentences and the overview must not print the first when it means the
	// second.
	CurrentVersionMixed bool `json:"current_version_mixed,omitempty" yaml:"current_version_mixed,omitempty"`
	// Changelog and ReleaseNotes carry the TARGET release's text only. They are the
	// fallback for a server that predates Releases; when Releases is present its last
	// entry carries the same text.
	Changelog    string `json:"changelog,omitempty" yaml:"changelog,omitempty"`
	ReleaseNotes string `json:"release_notes,omitempty" yaml:"release_notes,omitempty"`
	// Releases is every curated release this update crosses, oldest first, the last one
	// being the target. Empty when the server did not send any — either it is older than
	// the field, or there is genuinely nothing to cross.
	Releases []ReleaseSummary `json:"releases,omitempty" yaml:"releases,omitempty"`
	// ReleasesTruncated says older entries were dropped by the server.
	ReleasesTruncated bool `json:"releases_truncated,omitempty" yaml:"releases_truncated,omitempty"`
	// Resolution says HOW the server arrived at this answer — release, channel,
	// ahead_of_release, no_artifact_for_tag, not_tracked. It is DIAGNOSTIC ONLY:
	// nothing here may branch on it, and it exists so `has_update: false` can be
	// read in a log instead of guessed at.
	Resolution string            `json:"resolution,omitempty" yaml:"resolution,omitempty"`
	Components []ComponentUpdate `json:"components,omitempty" yaml:"components,omitempty"`
}

// ReleaseSummary is one curated release crossed by applying an update.
type ReleaseSummary struct {
	Version      string `json:"version" yaml:"version"`
	IsStable     bool   `json:"is_stable" yaml:"is_stable"`
	ReleasedAt   string `json:"released_at,omitempty" yaml:"released_at,omitempty"`
	Changelog    string `json:"changelog,omitempty" yaml:"changelog,omitempty"`
	ReleaseNotes string `json:"release_notes,omitempty" yaml:"release_notes,omitempty"`
}

// ComponentUpdate is one artefact inside a stack, and what would change for it.
//
// Being listed does NOT mean the artefact is outdated. Under the stable strategy the
// answer names every artefact of the release that matches something reported, so the
// installation can be attributed to a version even when nothing needs doing. Outdated is
// the conclusion drawn here by comparing digests.
type ComponentUpdate struct {
	Component  string `json:"component" yaml:"component"`
	OS         string `json:"os" yaml:"os"`
	Arch       string `json:"arch" yaml:"arch"`
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Tag        string `json:"tag,omitempty" yaml:"tag,omitempty"`

	// Action is what the SERVER says to do — current, install, upgrade,
	// downgrade or unknown — and it is the authority. Membership in the answer
	// is not: under a curated release every artefact matching a reported
	// component is listed, whether or not it differs from what is installed.
	//
	// Reason explains an `unknown`, and is empty for every other action. It is
	// the difference between "nothing to do" and "we publish nothing for what you
	// are running", which are opposite situations that used to arrive identically.
	Action string `json:"action,omitempty" yaml:"action,omitempty"`
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
	// PullReference is what to write into the compose variable before pulling,
	// ready to use. It is NOT always `repository:tag` — which tag to pull is the
	// cloud's decision and varies with the update strategy, so composing one here
	// out of the repository and tag fields would override the answer with a guess.
	PullReference string `json:"pull_reference,omitempty" yaml:"pull_reference,omitempty"`

	// Outdated is true only when the installed artefact is known AND differs from the
	// offered one.
	Outdated bool `json:"outdated" yaml:"outdated"`
	// Verifiable is false when there is nothing on THIS side to compare: nothing is
	// installed yet, or the server answered `unknown` and there is no artefact at all. An
	// unverifiable component is not a mismatch, and saying so is the difference between
	// "this will change" and "we cannot tell".
	//
	// It says nothing about whether the ANSWER carried a digest we can match — the server
	// compared against what we reported and its action is the authority either way.
	Verifiable     bool   `json:"verifiable" yaml:"verifiable"`
	CurrentVersion string `json:"current_version,omitempty" yaml:"current_version,omitempty"`

	// TargetVersion is the VERSION of the offered artefact, and only that. Files have
	// one; images do not (the cloud identifies their builds by digest), so it stays empty
	// for them. It is what the self-update download asks for as `?version=`.
	TargetVersion string `json:"target_version,omitempty" yaml:"target_version,omitempty"`

	// TargetDigest is the sha256 an updated artefact should end up carrying, in the same
	// identity the installed side reports: the config digest for an image, the file hash
	// for a file.
	//
	// Separate from TargetVersion because the two are not interchangeable and were once
	// the same field. Post-update verification compares this against what is on disk, and
	// with a version in it the Jaeger plugin — which has a hash and no version at all —
	// reported a mismatch after every successful update: "expected 0.13.0, got a423c6…".
	//
	// Empty means the cloud did not name a digest, and verification degrades to "nothing
	// to compare" rather than inventing a mismatch.
	TargetDigest string `json:"target_digest,omitempty" yaml:"target_digest,omitempty"`
}

// InstalledComponent is what the last update check reported about one artefact on this
// machine. It is kept because verifying an update afterwards needs to know what was
// installed, and re-inspecting every container a second time would ask Docker the same
// questions the check already asked.
type InstalledComponent struct {
	Component  string `json:"component" yaml:"component"`
	OS         string `json:"os" yaml:"os"`
	Arch       string `json:"arch" yaml:"arch"`
	Status     string `json:"status" yaml:"status"`
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Tag        string `json:"tag,omitempty" yaml:"tag,omitempty"`
	// Digest is the image config digest, or the file digest for components that ship as
	// files. Empty when nothing observable was installed.
	Digest  string `json:"digest,omitempty" yaml:"digest,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// UpdateCheckFailure explains why the update check produced no answer, in terms the
// interface can act on. A bare "server unavailable" tells the user nothing they can do
// about it — a spent daily quota and a broken proxy call for opposite reactions.
type UpdateCheckFailure struct {
	Reason  string `json:"reason" yaml:"reason"`
	Message string `json:"message" yaml:"message"`
	// RetryAfter is the server-advertised cooldown; zero when there is none, either
	// because the server did not say or because waiting will not help.
	RetryAfter time.Duration `json:"retry_after" yaml:"retry_after"`
	// Retryable separates "come back later" from "this needs a different license".
	Retryable bool `json:"retryable" yaml:"retryable"`
}

// updateStackFlags maps a contract stack onto the per-stack flag the menus already read,
// so the answer changes what those flags say without the screens having to change.
type stackFlag struct {
	stack models.ProductStack
	set   func(c *CheckResult, upToDate bool)
}

var updateStackFlags = []stackFlag{
	{models.ProductStackPentagi, func(c *CheckResult, v bool) { c.PentagiIsUpToDate = v }},
	{models.ProductStackWorker, func(c *CheckResult, v bool) { c.WorkerIsUpToDate = v }},
	{models.ProductStackInstaller, func(c *CheckResult, v bool) { c.InstallerIsUpToDate = v }},
	{models.ProductStackGraphiti, func(c *CheckResult, v bool) { c.GraphitiIsUpToDate = v }},
	{models.ProductStackLangfuse, func(c *CheckResult, v bool) { c.LangfuseIsUpToDate = v }},
	{models.ProductStackObservability, func(c *CheckResult, v bool) { c.ObservabilityIsUpToDate = v }},
}

// GatherUpdatesInfo asks the update server what is available for this installation.
//
// It never fails the surrounding gather: an update check that cannot be made is a fact
// about the installation, recorded in the result, not a reason to abandon everything else
// the checker learned.
func (h *defaultCheckHandler) GatherUpdatesInfo(ctx context.Context, c *CheckResult) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	client, err := h.updateClient()
	if err != nil {
		c.recordUpdateFailure(err)
		return nil
	}
	if warning := client.LicenseWarning(); warning != nil {
		// The check still runs, just without the license — say so, because otherwise the
		// only symptom is a tier the user did not expect.
		logger.Warnf("update check: %v", warning)
	}

	components := h.gatherComponents(ctx, c)
	// Recorded before the call: what is installed is a local fact, true whether or not the
	// server answers, and the verification after an update needs it either way.
	c.recordInstalledComponents(components)

	request := models.CheckUpdatesRequest{
		// The installer is a binary for THIS machine, so its platform is the host's.
		// Every other component describes an image and carries its own.
		InstallerVersion: cloud.NormalizeVersion(InstallerVersion),
		InstallerOS:      models.OSType(runtime.GOOS),
		InstallerArch:    models.ArchType(runtime.GOARCH),
		Strategy:         updateStrategy(h.appState),
		Images:           components.images,
		Files:            components.files,
		// Reported for its own sake: nothing in the answer depends on it. A stack
		// deployed elsewhere contributes no components at all — there is no
		// reference to resolve — so without this the service cannot tell "runs
		// Langfuse on their own cluster" from "does not use Langfuse".
		Stacks: stackStatuses(c),
		// What the product says about itself, carried through untouched. Absent
		// whenever the product is not running or cannot be asked, which is why it
		// is gathered last and why nothing below depends on it.
		Info: h.gatherProductInfo(ctx),
	}

	response, err := client.CheckUpdates(ctx, request)
	if err != nil {
		c.recordUpdateFailure(err)
		return nil
	}

	c.applyUpdateAnswer(response, components)

	return nil
}

// updateClient builds the client for this installation's configuration.
func (h *defaultCheckHandler) updateClient() (*cloud.Client, error) {
	defaultUpdateServerHost := strings.TrimPrefix(strings.TrimSuffix(DefaultUpdateServerEndpoint, "/"), "https://")
	return cloud.New(cloud.Config{
		Host:             getEnvVar(h.appState, "UPDATE_SERVER_HOST", defaultUpdateServerHost),
		ProxyURL:         getProxyURL(h.appState),
		LicenseKey:       getEnvVar(h.appState, "LICENSE_KEY", ""),
		InstallerVersion: InstallerVersion,
	})
}

// DefaultUpdateStrategy is the channel an installation follows when it has not
// been told otherwise.
//
// `preview` rather than `stable`, and the difference is not caution versus
// eagerness. Under `stable` the answer is measured against curated releases, and
// an installation whose stack has no published release is offered nothing at all
// — which reads to the user as "no updates exist" rather than "we are not
// publishing releases for this yet". `preview` follows the channels the client is
// already on, which is what an installation actually wants until releases are a
// thing it can rely on.
const DefaultUpdateStrategy = models.UpdateStrategyPreview

// updateStrategy reads which release channel to follow. An unrecognised value falls back
// to the default rather than failing the check, and says so: a typo should not leave an
// installation unable to learn about security updates.
func updateStrategy(appState state.State) models.UpdateStrategy {
	configured := strings.ToLower(strings.TrimSpace(getEnvVar(appState, "UPDATE_STRATEGY", "")))
	if configured == "" {
		return DefaultUpdateStrategy
	}

	strategy := models.UpdateStrategy(configured)
	if err := strategy.Valid(); err != nil {
		logger.Warnf("unknown UPDATE_STRATEGY %q, falling back to %s", configured, DefaultUpdateStrategy)
		return DefaultUpdateStrategy
	}

	return strategy
}

// recordInstalledComponents stores what the gathering observed on this machine.
func (c *CheckResult) recordInstalledComponents(components reportedComponents) {
	c.InstalledComponents = make([]InstalledComponent, 0, components.total())
	for _, component := range components.images {
		c.InstalledComponents = append(c.InstalledComponents, InstalledComponent{
			Component:  component.Component.String(),
			OS:         component.OS.String(),
			Arch:       component.Arch.String(),
			Status:     component.Status.String(),
			Repository: component.Repository,
			Tag:        component.Tag,
			Digest:     derefString(component.ImageHash),
		})
	}
	for _, component := range components.files {
		c.InstalledComponents = append(c.InstalledComponents, InstalledComponent{
			Component: component.Component.String(),
			OS:        component.OS.String(),
			Arch:      component.Arch.String(),
			Status:    component.Status.String(),
			Digest:    derefString(component.FileHash),
			Version:   derefString(component.Version),
		})
	}
}

// recordUpdateFailure stores why the check produced nothing and resets every per-stack
// flag.
//
// Resetting ALL of them matters: leaving one behind means a stack keeps whatever verdict
// an earlier, successful check left in it, and the interface presents a stale answer as a
// current one. The worker flag used to be the one left behind.
func (c *CheckResult) recordUpdateFailure(err error) {
	c.UpdateServerAccessible = false
	c.StackUpdates = nil
	for _, flag := range updateStackFlags {
		flag.set(c, false)
	}

	failure := &UpdateCheckFailure{
		Reason:  string(cloud.FailureUnreachable),
		Message: err.Error(),
	}
	var classified *cloud.Failure
	if errors.As(err, &classified) {
		failure.Reason = string(classified.Reason)
		failure.RetryAfter = classified.RetryAfter
		failure.Retryable = classified.Reason.Retryable()
	}
	c.UpdateFailure = failure

	logger.Warnf("update check failed (%s): %v", failure.Reason, err)
}

// applyUpdateAnswer turns the server's answer into the per-stack verdicts the menus read
// and the detail the update screen shows.
//
// A stack missing from the answer stays "up to date": the server reports on what it was
// asked about, and no news is not the same as bad news. Offering an update nobody has
// evidence for is the worse failure.
func (c *CheckResult) applyUpdateAnswer(response *models.CheckUpdatesResponse, reported reportedComponents) {
	c.UpdateServerAccessible = true
	c.UpdateFailure = nil
	for _, flag := range updateStackFlags {
		flag.set(c, true)
	}

	installed := make(map[string]installedArtefact, reported.total())
	for _, component := range reported.images {
		installed[componentKey(component.Component, component.OS, component.Arch)] = installedArtefact{
			digest: component.ImageHash,
		}
	}
	for _, component := range reported.files {
		installed[componentKey(component.Component, component.OS, component.Arch)] = installedArtefact{
			digest:  component.FileHash,
			version: component.Version,
		}
	}

	c.StackUpdates = make([]StackUpdate, 0, len(response.Updates))
	for _, update := range response.Updates {
		for _, flag := range updateStackFlags {
			if flag.stack == update.Stack {
				flag.set(c, !update.HasUpdate)
			}
		}
		c.StackUpdates = append(c.StackUpdates, buildStackUpdate(update, installed))
	}
}

// installedArtefact is what is known about an artefact on this machine, in the one shape
// both kinds share: something to compare an offer against.
type installedArtefact struct {
	digest  *string
	version *string
}

func buildStackUpdate(update models.UpdateInfo, installed map[string]installedArtefact) StackUpdate {
	stack := StackUpdate{
		Stack:               update.Stack.String(),
		HasUpdate:           update.HasUpdate,
		CurrentVersion:      derefString(update.CurrentVersion),
		CurrentVersionMixed: update.CurrentVersionMixed,
		LatestVersion:       derefString(update.LatestVersion),
		Changelog:           derefString(update.Changelog), //nolint:staticcheck
		ReleaseNotes:        derefString(update.ReleaseNotes),
		ReleasesTruncated:   update.ReleasesTruncated,
		Resolution:          string(update.Resolution),
	}

	// Copied rather than referenced: CheckResult is serialised to the state file, and a
	// time.Time there would round-trip through a format nothing else in this struct uses.
	// The overview only ever prints the date.
	for _, release := range update.Releases {
		summary := ReleaseSummary{
			Version:      release.Version,
			IsStable:     release.IsStable,
			Changelog:    release.Changelog,
			ReleaseNotes: release.ReleaseNotes,
		}
		if release.ReleasedAt != nil {
			summary.ReleasedAt = release.ReleasedAt.UTC().Format(time.RFC3339)
		}
		stack.Releases = append(stack.Releases, summary)
	}

	stack.Components = make([]ComponentUpdate, 0, len(update.Images)+len(update.Files))
	for _, offered := range update.Images {
		current := installed[componentKey(offered.Component, offered.OS, offered.Arch)]
		stack.Components = append(stack.Components, buildImageComponentUpdate(offered, current))
	}
	for _, offered := range update.Files {
		current := installed[componentKey(offered.Component, offered.OS, offered.Arch)]
		stack.Components = append(stack.Components, buildFileComponentUpdate(offered, current))
	}

	// Logged and nothing more. `resolution` explains an answer — "you are ahead
	// of the curated set", "we track nothing under that reference" — and the one
	// thing it must never do is drive a decision here: it is diagnostic by
	// contract, and a client branching on it would couple itself to wording the
	// server is free to extend.
	if stack.Resolution != "" {
		logger.Debugf("update check: stack %s resolved as %s (has_update=%t)",
			stack.Stack, stack.Resolution, stack.HasUpdate)
	}
	for _, component := range stack.Components {
		if component.Reason != "" {
			logger.Debugf("update check: %s is %s — %s",
				component.Component, component.Action, component.Reason)
		}
	}

	return stack
}

// buildComponentUpdate decides whether one artefact would actually change.
//
// Images are compared by digest, not by tag: a moving tag such as "latest" points at a
// different image over time, so the tag alone says nothing. Binaries are compared by
// version, which is how the server identifies their builds.
//
// A comparison that cannot be made is reported as unverifiable rather than as a
// difference. The server omits a digest it does not know, and treating an absent value as
// a mismatch would offer an update on no evidence at all.
func buildImageComponentUpdate(offered models.ImageUpdate, current installedArtefact) ComponentUpdate {
	component := ComponentUpdate{
		Component:     offered.Component.String(),
		OS:            offered.OS.String(),
		Arch:          offered.Arch.String(),
		Repository:    offered.Repository,
		Tag:           offered.Tag,
		Action:        string(offered.Action),
		PullReference: offered.PullReference,
	}
	if offered.Reason != nil {
		component.Reason = string(*offered.Reason)
	}

	// The target an updated image should end up carrying is the offered CONFIG
	// digest: the installed side of every comparison is normalizeDigest(image.ID),
	// which is `docker inspect` → Id, and ConfigHash is the same identity as the
	// server resolved it (cloud/models: "this is the value to verify a pulled
	// image against"). Without this the post-update verification had a target of
	// "" for every image, routed each one to "not verifiable", and then printed
	// the pass line for an update that may not have taken.
	//
	// When the server does not know the config digest the target stays empty and
	// verification degrades honestly to "cannot verify" — never to a mismatch.
	//
	// TargetVersion is deliberately left empty here: an image has no version of its
	// own in the answer, and the field means a version everywhere else.
	if offered.ConfigHash != nil {
		component.TargetDigest = *offered.ConfigHash
	}

	// Nothing installed means nothing to compare, which is not a mismatch: it is the
	// state of a component that has never been pulled, and the answer names what to pull
	// rather than a difference to reconcile.
	if current.digest != nil {
		component.CurrentVersion = *current.digest
		component.Verifiable = true
		component.Outdated = !offered.CarriesDigest(*current.digest)
	}

	// The server's own verdict wins where it is more informed than a digest
	// comparison can be, and it always is: it compared against the digest THIS
	// client reported, knowing every identity the artefact has.
	//
	// `install` is the case that matters most: nothing is pulled, so there is no
	// digest to compare and the loop above leaves `Outdated` false — which reads
	// as "nothing to do" for the one component that most needs doing. `unknown`
	// is the mirror image: there is no artefact at all, so whatever a comparison
	// produced means nothing.
	//
	// `current` is here because the comparison above can disagree with it and be
	// wrong. Docker exposes three sha256 identities per image and this client
	// reports the config digest; an answer that names only the manifest digest
	// therefore matches nothing locally, and `CarriesDigest` returns false for an
	// installation that is exactly up to date. Without this branch the overview
	// tells the user a component will be updated while the server has just said
	// it will not.
	switch offered.Action {
	case models.ComponentActionInstall, models.ComponentActionUpgrade, models.ComponentActionDowngrade:
		component.Outdated = true
	case models.ComponentActionCurrent:
		component.Outdated = false
	case models.ComponentActionUnknown:
		component.Outdated = false
		component.Verifiable = false
	}

	return component
}

// buildFileComponentUpdate is the file half. Files are compared by version, which is how
// the server identifies their builds, and by hash when no version was reported — which is
// the Jaeger plugin's case, since it ships without one.
func buildFileComponentUpdate(offered models.FileUpdate, current installedArtefact) ComponentUpdate {
	component := ComponentUpdate{
		Component: offered.Component.String(),
		OS:        offered.OS.String(),
		Arch:      offered.Arch.String(),
		Action:    string(offered.Action),
		// A file carries both, and they describe different sides. The version is the
		// OFFERED build's, and it is always present — a published binary cannot exist
		// without one — and it is what the download asks for, since the package request
		// requires a version and has no field for a hash. The hash is what the result is
		// verified against, and it is the only identity the INSTALLED side has for a
		// plugin: a file on disk carries no version metadata. Conflating the two is what
		// made a correctly updated plugin report "expected 0.13.0, got a423c6…".
		TargetVersion: offered.Version,
		TargetDigest:  offered.FileHash,
	}
	if offered.Reason != nil {
		component.Reason = string(*offered.Reason)
	}

	switch {
	case current.version != nil:
		component.CurrentVersion = *current.version
		component.Verifiable = true
		component.Outdated = offered.Version != *current.version

	case current.digest != nil:
		component.CurrentVersion = *current.digest
		component.Verifiable = true
		component.Outdated = offered.FileHash != *current.digest
	}

	// Same rule as the image half, for the same reason: the server's action is the
	// authority, because it compared against what this client reported. A file that
	// was never installed reports neither a version nor a digest, so both branches
	// above are skipped; and a `current` verdict has to survive a version string
	// that merely spells the same build differently.
	switch offered.Action {
	case models.ComponentActionInstall, models.ComponentActionUpgrade, models.ComponentActionDowngrade:
		component.Outdated = true
	case models.ComponentActionCurrent:
		component.Outdated = false
	case models.ComponentActionUnknown:
		component.Outdated = false
		component.Verifiable = false
	}

	return component
}

// componentKey identifies an artefact across the request and the answer. The platform is
// part of it: the Jaeger plugin is reported twice, once per architecture, and matching on
// the component name alone would collapse the two.
func componentKey(component models.ComponentType, os models.OSType, arch models.ArchType) string {
	return component.String() + "/" + os.String() + "/" + arch.String()
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// stackStatuses says how each product stack is deployed on this machine.
//
// Every stack is reported, including the unused ones: silence about a stack is
// indistinguishable from an installer too old to speak about stacks at all, and
// the difference matters to whoever reads the numbers.
//
// The precedence between installed and external is deliberate. `external` is the
// more specific statement — it says who OPERATES the stack — and an installation
// pointing at somebody else's Langfuse may well also have the compose file on
// disk.
func stackStatuses(c *CheckResult) []models.StackInfo {
	state := func(connected, installed, external bool) models.StackStatus {
		switch {
		case !connected:
			return models.StackStatusUnused
		case external:
			return models.StackStatusExternal
		case installed:
			return models.StackStatusInstalled
		default:
			return models.StackStatusConnected
		}
	}
	worker := models.StackStatusUnused
	if c.WorkerImageExists {
		worker = models.StackStatusInstalled
	}
	return []models.StackInfo{
		// The product itself and the installer asking the question: both present
		// by definition, or this check would not be running.
		{Stack: models.ProductStackPentagi, Status: models.StackStatusInstalled},
		{Stack: models.ProductStackInstaller, Status: models.StackStatusInstalled},
		{Stack: models.ProductStackWorker, Status: worker},
		{Stack: models.ProductStackGraphiti, Status: state(
			c.GraphitiConnected, c.GraphitiInstalled, c.GraphitiExternal)},
		{Stack: models.ProductStackLangfuse, Status: state(
			c.LangfuseConnected, c.LangfuseInstalled, c.LangfuseExternal)},
		{Stack: models.ProductStackObservability, Status: state(
			c.ObservabilityConnected, c.ObservabilityInstalled, c.ObservabilityExternal)},
		// Not deployed by compose, so the installer has nothing to observe about
		// them — and `undefined` is the value that says exactly that.
		//
		// They used to be reported as `unused`, on the reasoning that saying
		// something beats saying nothing. That reasoning held only while silence
		// and `unused` were recorded identically: it made a placeholder for "we
		// cannot tell" indistinguishable from a measured negative, and every count
		// of engine adoption was reading a hardcoded constant.
		{Stack: models.ProductStackEngine, Status: models.StackStatusUndefined},
		{Stack: models.ProductStackBrowser, Status: models.StackStatusUndefined},
	}
}
