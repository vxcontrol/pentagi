package update

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/version"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/cloud/models"
)

// A development build is named after its branch, and the contract validates this
// field as a version. Sending "develop" would have the ENTIRE request refused
// over a field that is decoration — and refused silently, because the failure
// arrives as a validation error nobody on this side can see. The build itself is
// not lost: it travels in the User-Agent.
func TestABranchBuildIsStillAcceptedByTheContract(t *testing.T) {
	for name, raw := range map[string]string{
		"a branch name": "develop",
		"a commit hash": "a1b2c3d",
		"nothing":       "",
		"a tag with v":  "v2.1.0",
	} {
		normalised := normalizeVersion(raw)
		if err := models.GetValidator().Var(normalised, "semver"); err != nil {
			t.Errorf("%s (%q) normalised to %q, which the contract refuses: %v",
				name, raw, normalised, err)
		}
	}

	// A real version passes through untouched — the normalisation must not be a
	// blanket replacement, or every report would claim the same build.
	if got := normalizeVersion("2.1.0"); got != "2.1.0" {
		t.Errorf("a valid version was rewritten to %q", got)
	}
	if got := normalizeVersion("2.1.0-93e99748"); got != "2.1.0-93e99748" {
		t.Errorf("a branch build with a version core was rewritten to %q", got)
	}
}

// TestAReportFitsWhatTheContractAccepts.
//
// The document is bounded by the contract, and the bound is not generous next to
// a summary that grows with the provider list. Exceeding it does not truncate
// anything — the service refuses the whole request — so the size is checked
// before sending rather than discovered afterwards.
func TestAReportFitsWhatTheContractAccepts(t *testing.T) {
	summary := database.InstanceSummary{
		Schema:  database.InstanceSummarySchema,
		Version: "2.1.0",
	}
	// Every provider type there is, each with usage: the widest a real document
	// gets today.
	for _, provider := range []string{
		"openai", "anthropic", "gemini", "bedrock", "ollama", "custom",
		"deepseek", "glm", "kimi", "qwen", "minimax",
	} {
		summary.Providers = append(summary.Providers,
			database.InstanceProvider{Type: provider, Count: 3})
		summary.Usage = append(summary.Usage, database.InstanceUsage{
			Provider: provider, Chains: 99999, TokensIn: 123456789, TokensOut: 987654321,
			CacheIn: 1234567, CacheOut: 7654321, CostIn: 12345.67, CostOut: 7654.32,
		})
	}
	summary.OAuth = []string{"google", "github"}
	summary.Skipped = []string{"toolcalls", "usage"}

	document, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("the description must serialise: %v", err)
	}
	if len(document) > models.MaxProductInfoBytes {
		t.Errorf("a full description is %d bytes, over the %d the contract accepts — "+
			"every report from a fully configured installation would be refused whole",
			len(document), models.MaxProductInfoBytes)
	}
}

// TestTheDescriptionNamesNoOne is the last gate before the document leaves the
// process. The summary is built elsewhere and tested there; this checks the thing
// that actually goes on the wire, because that is what cannot be taken back.
func TestTheDescriptionNamesNoOne(t *testing.T) {
	summary := database.InstanceSummary{
		Schema:    database.InstanceSummarySchema,
		Version:   "2.1.0",
		Providers: []database.InstanceProvider{{Type: "openai", Count: 2}},
		OAuth:     []string{"google"},
		Usage:     []database.InstanceUsage{{Provider: "openai", TokensIn: 10}},
	}
	summary.Counts.Users = 3
	summary.Counts.Flows = 12

	document, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("serialise: %v", err)
	}

	// Keys, not values: a value could legitimately be any string, but a key names
	// what kind of thing the document is prepared to carry.
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(document, &shape); err != nil {
		t.Fatalf("the document must be an object: %v", err)
	}

	allowed := map[string]bool{
		"schema": true, "version": true, "develop": true, "features": true,
		"providers": true, "oauth": true, "counts": true, "usage": true, "skipped": true,
		// The provider type embeddings actually go to, as the product resolves it.
		// A closed vocabulary and a deliberate choice: an "openai" embedder pointed
		// at a custom URL is not talking to OpenAI, and the base URL — which would
		// name somebody's infrastructure — is never sent.
		"embedding_provider": true,
		// What this process is, and it has to be in the document rather than only in
		// the request envelope: whoever forwards a description cannot produce these.
		// A caller running this binary in a container sees the platform of the HOST,
		// and the image digest it can see is not the file that is executing.
		//
		// Neither says anything about the machine or the people on it. `os`/`arch`
		// are the two-value vocabularies we compile for, and `binary_hash` is the
		// sha256 of an executable we published — it identifies OUR build, not this
		// installation, and it is the only thing that tells two builds carrying the
		// same version string apart.
		"os": true, "arch": true, "binary_hash": true,
	}
	for key := range shape {
		if !allowed[key] {
			t.Errorf("the document carries %q, which nobody decided to send — every key here "+
				"has to be a deliberate choice, because this leaves the machine", key)
		}
	}

	// And the one thing no key list can catch: an identifier smuggled inside a
	// value. The installation is identified by the SDK headers, never by content.
	if strings.Contains(strings.ToLower(string(document)), "@") {
		t.Errorf("the document contains an address-shaped value: %s", document)
	}
}

// TestAMisconfiguredStrategyIsDroppedRatherThanForwarded.
//
// UPDATE_STRATEGY is an operator-supplied string and the field is validated on both
// sides, so forwarding a typo would have the whole check refused — and a check that
// is refused is an installation that stops being told about updates AND stops being
// visible, over one misspelt variable. Omitting the field has the service answer
// under its own default, which is the same channel this product defaults to.
func TestAMisconfiguredStrategyIsDroppedRatherThanForwarded(t *testing.T) {
	for _, valid := range []string{"preview", "stable", "nightly"} {
		if got := updateStrategy(&config.Config{UpdateStrategy: valid}); string(got) != valid {
			t.Errorf("%q must be forwarded, got %q", valid, got)
		}
	}
	// Whitespace is the shape a .env value actually arrives in.
	if got := updateStrategy(&config.Config{UpdateStrategy: "  stable  "}); got != models.UpdateStrategyStable {
		t.Errorf("a padded value must still be recognised, got %q", got)
	}
	for _, invalid := range []string{"", "Stable", "latest", "beta"} {
		if got := updateStrategy(&config.Config{UpdateStrategy: invalid}); got != "" {
			t.Errorf("%q must be dropped, got %q", invalid, got)
		}
	}

	// And an omitted strategy has to leave the request valid: this is the field
	// that made the difference between being answered and being refused.
	request := models.ProductUpdateRequest{
		Schema:   1,
		Version:  "2.1.0",
		OS:       models.OSTypeLinux,
		Arch:     models.ArchTypeAMD64,
		Strategy: updateStrategy(&config.Config{UpdateStrategy: "nonsense"}),
		Info:     json.RawMessage(`{"schema":1}`),
	}
	if err := request.Valid(); err != nil {
		t.Errorf("a request with no strategy must satisfy the contract: %v", err)
	}
}

// TestTheAnswerIsReadThroughTheEnvelope.
//
// Every answer is wrapped in {"status": …, "data": …}. Decoding a body straight
// into the payload type SUCCEEDS and yields a zero value, so a client that skips
// the envelope reads every answer as "nothing to do" and never finds out — which is
// indistinguishable from a fleet that is perfectly up to date.
func TestTheAnswerIsReadThroughTheEnvelope(t *testing.T) {
	const body = `{"status":"success","data":{
		"update":{"stack":"pentagi","has_update":true,"resolution":"channel"},
		"version":"2.4.0","identified":true}}`

	answer, err := models.ParseEnvelope[models.ProductUpdateResponse]([]byte(body))
	if err != nil {
		t.Fatalf("the answer must decode: %v", err)
	}
	if answer.Update == nil || !answer.Update.HasUpdate {
		t.Fatalf("the update was lost: %+v", answer.Update)
	}
	if answer.Version == nil || *answer.Version != "2.4.0" {
		t.Errorf("the offered version was lost: %v", answer.Version)
	}

	// The failure this guards: the same bytes read WITHOUT the envelope.
	var direct models.ProductUpdateResponse
	if err := json.Unmarshal([]byte(body), &direct); err != nil {
		t.Fatalf("the wrong way must still parse, which is what makes it dangerous: %v", err)
	}
	if direct.Update != nil {
		t.Error("decoding past the envelope must not accidentally work; this test is measuring nothing")
	}
}

// TestAnAnswerWithoutAMatchedBuildIsNotReportedAsAnUpdate.
//
// `has_update` is true whenever the service has something to offer, including when
// it could not work out which image the running build is — a development build, a
// private rebuild — in which case the offer is "the newest we publish" rather than
// a step from here. Reading that as an update available would announce one on every
// check, forever, for every such installation. The version it names is still kept,
// beside a verdict that says it could not be compared to this build.
//
// The two empty shapes are here for a different reason: both have a nil Update, and
// reaching into it is a panic in a goroutine the server does not supervise.
func TestAnAnswerWithoutAMatchedBuildIsNotReportedAsAnUpdate(t *testing.T) {
	offered := "2.4.0"
	cases := map[string]struct {
		answer *models.ProductUpdateResponse
		state  State
		latest string
		expect string
	}{
		"nothing at all": {nil, StateUnknown, "", "no update information"},
		"nothing for this platform": {
			&models.ProductUpdateResponse{}, StateUnknown, "", "no update information",
		},
		"a build we do not recognise": {
			&models.ProductUpdateResponse{
				Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
				Version:    &offered,
				Identified: false,
			},
			StateUnknown, "2.4.0", "not one we published",
		},
		"up to date": {
			&models.ProductUpdateResponse{
				Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi},
				Identified: true,
			},
			StateUpToDate, "", "up to date",
		},
		"a real update": {
			&models.ProductUpdateResponse{
				Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
				Version:    &offered,
				Identified: true,
			},
			StateUpdateAvailable, "2.4.0", "an update is available",
		},
		// An offer with nothing to move to is not something the operator can act
		// on, and a badge pointing at nothing is exactly what they would try to
		// act on.
		"an update with no version to move to": {
			&models.ProductUpdateResponse{
				Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
				Identified: true,
			},
			StateUnknown, "", "no version is named",
		},
		// The version can also arrive on the stack rather than on the image.
		"a version named only on the stack": {
			&models.ProductUpdateResponse{
				Update: &models.UpdateInfo{
					Stack: models.ProductStackPentagi, HasUpdate: true, LatestVersion: &offered,
				},
				Identified: true,
			},
			StateUpdateAvailable, "2.4.0", "an update is available",
		},
	}

	for name, tc := range cases {
		logger, recorded := recordingLogger()
		service := &Service{logger: logger}
		service.record(tc.answer)

		if len(*recorded) != 1 {
			t.Fatalf("%s: expected exactly one line, got %d", name, len(*recorded))
		}
		if !strings.Contains((*recorded)[0], tc.expect) {
			t.Errorf("%s: said %q, expected it to mention %q", name, (*recorded)[0], tc.expect)
		}

		status := service.Status()
		if status.State != tc.state {
			t.Errorf("%s: kept state %q, expected %q", name, status.State, tc.state)
		}
		if status.Latest != tc.latest {
			t.Errorf("%s: kept latest %q, expected %q", name, status.Latest, tc.latest)
		}
		if status.CheckedAt.IsZero() {
			t.Errorf("%s: an answer must be dated", name)
		}
		if !status.FailedAt.IsZero() {
			t.Errorf("%s: an answer clears the last failure", name)
		}
	}
}

// TestAVerdictOutlivesTheFailuresThatFollowIt.
//
// The service is asked every few hours and the answer is shown for the whole
// interval. A failure in between does not make an update that was available an
// hour ago unavailable — the build has not changed, a replaced binary being a new
// process — so the verdict stays and the failure is dated beside it. Only an
// installation that has never heard anything reads as unreachable.
func TestAVerdictOutlivesTheFailuresThatFollowIt(t *testing.T) {
	logger, _ := recordingLogger()
	offered := "2.4.0"
	service := &Service{logger: logger, status: Status{State: StatePending}}

	service.recordFailure(errors.New("no route to host"))
	if got := service.Status(); got.State != StateUnreachable || got.FailedAt.IsZero() {
		t.Fatalf("a failure before any answer must read as unreachable and be dated: %+v", got)
	}

	service.record(&models.ProductUpdateResponse{
		Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
		Version:    &offered,
		Identified: true,
	})
	if got := service.Status(); got.State != StateUpdateAvailable || !got.FailedAt.IsZero() {
		t.Fatalf("an answer must set the verdict and clear the failure: %+v", got)
	}

	service.recordFailure(errors.New("no route to host"))
	got := service.Status()
	if got.State != StateUpdateAvailable || got.Latest != "2.4.0" {
		t.Errorf("a failure must not take back a verdict: %+v", got)
	}
	if got.FailedAt.IsZero() {
		t.Errorf("the failure must still be dated beside the verdict: %+v", got)
	}
}

// TestARunWithNoIntervalSaysSo.
//
// The badge distinguishes "switched off" from "never answered": an operator who set
// UPDATE_CHECK_INTERVAL=0 should not be told the update service is unreachable.
func TestARunWithNoIntervalSaysSo(t *testing.T) {
	logger, _ := recordingLogger()
	service := &Service{cfg: &config.Config{}, logger: logger, status: Status{State: StatePending}}

	service.Run(t.Context())

	if got := service.Status().State; got != StateDisabled {
		t.Errorf("a zero interval must read as disabled, got %q", got)
	}
}

// TestAnInstallationWithoutAServiceStillNamesItsBuild.
//
// The half of the badge that never depends on the service is the version itself,
// and it has to be right even when the service could not be built at all.
func TestAnInstallationWithoutAServiceStillNamesItsBuild(t *testing.T) {
	status := DisabledStatus(&config.Config{UpdateStrategy: "stable"})
	if status.Current != version.GetBinaryVersion() {
		t.Errorf("the build is %q, expected %q", status.Current, version.GetBinaryVersion())
	}
	if status.State != StateDisabled {
		t.Errorf("state %q, expected disabled", status.State)
	}
	if status.Strategy != "stable" {
		t.Errorf("a configured channel must be reported, got %q", status.Strategy)
	}
	// The channel the service actually answers under, not the misspelt one.
	if got := DisabledStatus(&config.Config{UpdateStrategy: "nonsense"}).Strategy; got != "preview" {
		t.Errorf("an unrecognised channel must report the default, got %q", got)
	}
	if got := DisabledStatus(nil).Strategy; got != "preview" {
		t.Errorf("no configuration must still report the default, got %q", got)
	}
}

// TestStatusIsSafeToReadWhileItChanges is for the race detector: the resolver
// reads on every page load, and the ticker writes from its own goroutine.
func TestStatusIsSafeToReadWhileItChanges(t *testing.T) {
	logger, _ := recordingLogger()
	offered := "2.4.0"
	service := &Service{logger: logger, status: Status{State: StatePending}}
	answer := &models.ProductUpdateResponse{
		Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
		Version:    &offered,
		Identified: true,
	}

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(3)
		go func() { defer wg.Done(); service.record(answer) }()
		go func() { defer wg.Done(); service.recordFailure(errors.New("flaky")) }()
		go func() { defer wg.Done(); _ = service.Status() }()
	}
	wg.Wait()

	if got := service.Status(); got.State != StateUpdateAvailable {
		t.Errorf("a verdict was lost to a concurrent failure: %+v", got)
	}
}

// recordingLogger is a logger whose messages can be read back. Debug level because
// that is the level this package speaks at, and a hook attached above it would
// record nothing.
func recordingLogger() (*logrus.Entry, *[]string) {
	recorded := &[]string{}
	base := logrus.New()
	base.SetLevel(logrus.DebugLevel)
	base.SetOutput(io.Discard)
	base.AddHook(&collector{recorded: recorded})
	return base.WithField("component", "update"), recorded
}

// collector is safe to fire from several goroutines: logrus invokes hooks outside
// its own lock, and the service logs from whichever goroutine records.
type collector struct {
	mu       sync.Mutex
	recorded *[]string
}

func (c *collector) Levels() []logrus.Level { return logrus.AllLevels }

func (c *collector) Fire(entry *logrus.Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	*c.recorded = append(*c.recorded, entry.Message)
	return nil
}
