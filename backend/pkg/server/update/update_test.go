package update

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/version"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/cloud/models"
)

func TestUpdate_NormalizeVersion_SendsOnlyAVersionTheContractAccepts(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"a commit hash", "a1b2c3d", "0.0.0"},
		{"nothing", "", "0.0.0"},
		{"a tag with v", "v2.1.0", "0.0.0"},
		{"a release", "2.1.0", "2.1.0"},
		{"a branch build with a version core", "2.1.0-93e99748", "2.1.0-93e99748"},
		{"an edition build naming its base version", "2.1.0-ce.h93e99748", "2.1.0-ce.h93e99748"},
		{"an enterprise edition build", "2.1.0-ee.h93e99748", "2.1.0-ee.h93e99748"},
		{"an edition release", "2.1.0-ce", "2.1.0-ce"},
		{"an edition alone", "ce", "0.0.0"},
		{"an edition and a revision without a release", "ce.h93e99748", "0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVersion(tt.raw)
			if got != tt.want {
				t.Errorf("%q normalised to %q, want %q", tt.raw, got, tt.want)
			}
			if err := models.GetValidator().Var(got, "semver"); err != nil {
				t.Errorf("%q normalised to %q, which the contract refuses: %v", tt.raw, got, err)
			}
		})
	}
}

func TestUpdate_UpdateStrategy_DropsAChannelTheContractWouldRefuse(t *testing.T) {
	tests := []struct {
		name       string
		configured string
		want       models.UpdateStrategy
	}{
		{"the preview channel", "preview", "preview"},
		{"the stable channel", "stable", "stable"},
		{"the nightly channel", "nightly", "nightly"},
		{"a value padded the way a .env file pads it", "  stable  ", "stable"},
		{"nothing configured", "", ""},
		{"a capitalised channel", "Stable", ""},
		{"a channel nobody publishes", "latest", ""},
		{"another channel nobody publishes", "beta", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateStrategy(&config.Config{UpdateStrategy: tt.configured}); got != tt.want {
				t.Errorf("%q became %q, want %q", tt.configured, got, tt.want)
			}
		})
	}
}

func TestUpdate_Send_ReadsTheAnswerThroughTheEnvelope(t *testing.T) {
	const body = `{"status":"success","data":{
		"update":{"stack":"pentagi","has_update":true,"resolution":"channel"},
		"version":"2.4.0","identified":true}}`

	var direct models.ProductUpdateResponse
	if err := json.Unmarshal([]byte(body), &direct); err != nil || direct.Update != nil {
		t.Fatalf("the answer must parse past the envelope into nothing, or this test measures nothing: %v", err)
	}

	sent := &client{call: func(ctx context.Context, _ []byte) ([]byte, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		return []byte(body), nil
	}}
	answer, err := sent.send(t.Context(), models.ProductUpdateRequest{
		Schema:   1,
		Version:  "2.1.0",
		OS:       models.OSTypeLinux,
		Arch:     models.ArchTypeAMD64,
		Strategy: updateStrategy(&config.Config{UpdateStrategy: "nonsense"}),
		Info:     json.RawMessage(`{"schema":1}`),
	})
	if err != nil {
		t.Fatalf("a request with its strategy dropped must satisfy the contract and its answer decode: %v", err)
	}
	if answer.Update == nil || !answer.Update.HasUpdate {
		t.Fatalf("the update was lost: %+v", answer.Update)
	}
	if answer.Version == nil || *answer.Version != "2.4.0" {
		t.Errorf("the offered version was lost: %v", answer.Version)
	}
}

func TestUpdate_Record_ReportsAnUpdateOnlyForAMatchedBuildNamingAVersion(t *testing.T) {
	offered := "2.4.0"
	cases := []struct {
		name   string
		answer *models.ProductUpdateResponse
		state  State
		latest string
		expect string
	}{
		{"nothing at all", nil, StateUnknown, "", "no update information"},
		{"nothing for this platform", &models.ProductUpdateResponse{}, StateUnknown, "", "no update information"},
		{"a build we do not recognise", &models.ProductUpdateResponse{
			Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
			Version:    &offered,
			Identified: false,
		}, StateUnknown, "2.4.0", "not one we published"},
		{"up to date", &models.ProductUpdateResponse{
			Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi},
			Identified: true,
		}, StateUpToDate, "", "up to date"},
		{"a real update", &models.ProductUpdateResponse{
			Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
			Version:    &offered,
			Identified: true,
		}, StateUpdateAvailable, "2.4.0", "an update is available"},
		{"an update with no version to move to", &models.ProductUpdateResponse{
			Update:     &models.UpdateInfo{Stack: models.ProductStackPentagi, HasUpdate: true},
			Identified: true,
		}, StateUnknown, "", "no version is named"},
		{"a version named only on the stack", &models.ProductUpdateResponse{
			Update: &models.UpdateInfo{
				Stack: models.ProductStackPentagi, HasUpdate: true, LatestVersion: &offered,
			},
			Identified: true,
		}, StateUpdateAvailable, "2.4.0", "an update is available"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger, recorded := updateRecordingLogger()
			service := &Service{logger: logger}
			service.record(tc.answer)

			if len(*recorded) != 1 {
				t.Fatalf("expected exactly one line, got %d", len(*recorded))
			}
			if !strings.Contains((*recorded)[0], tc.expect) {
				t.Errorf("said %q, expected it to mention %q", (*recorded)[0], tc.expect)
			}

			status := service.Status()
			if status.State != tc.state {
				t.Errorf("kept state %q, expected %q", status.State, tc.state)
			}
			if status.Latest != tc.latest {
				t.Errorf("kept latest %q, expected %q", status.Latest, tc.latest)
			}
			if status.CheckedAt.IsZero() {
				t.Error("an answer must be dated")
			}
			if !status.FailedAt.IsZero() {
				t.Error("an answer clears the last failure")
			}
		})
	}
}

func TestUpdate_RecordFailure_KeepsAVerdictAlreadyReached(t *testing.T) {
	logger, _ := updateRecordingLogger()
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

func TestUpdate_Run_ReadsAZeroIntervalAsDisabled(t *testing.T) {
	logger, _ := updateRecordingLogger()
	service := &Service{cfg: &config.Config{}, logger: logger, status: Status{State: StatePending}}

	service.Run(t.Context())

	if got := service.Status().State; got != StateDisabled {
		t.Errorf("a zero interval must read as disabled, got %q", got)
	}
}

func TestUpdate_DisabledStatus_StillNamesTheBuildAndItsChannel(t *testing.T) {
	previousVer, previousRev := version.PackageVer, version.PackageRev
	t.Cleanup(func() { version.PackageVer, version.PackageRev = previousVer, previousRev })
	version.PackageVer, version.PackageRev = "2.1.0", ""

	// The reported build carries its edition: a release is a prerelease of the
	// number it names, and what identifies the running binary is the whole string.
	status := DisabledStatus(&config.Config{UpdateStrategy: "stable"})
	if status.Current != "2.1.0-ce" {
		t.Errorf("the build is %q, expected 2.1.0-ce", status.Current)
	}
	if status.State != StateDisabled {
		t.Errorf("state %q, expected disabled", status.State)
	}
	if status.Strategy != "stable" {
		t.Errorf("a configured channel must be reported, got %q", status.Strategy)
	}
	if got := DisabledStatus(&config.Config{UpdateStrategy: "nonsense"}).Strategy; got != "preview" {
		t.Errorf("an unrecognised channel must report the default, got %q", got)
	}
	if got := DisabledStatus(nil).Strategy; got != "preview" {
		t.Errorf("no configuration must still report the default, got %q", got)
	}
}

// Run under -race: the resolver reads on every page load while the ticker writes.
func TestUpdate_Status_IsSafeToReadWhileItChanges(t *testing.T) {
	logger, _ := updateRecordingLogger()
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

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the concurrent writers never finished")
	}

	if got := service.Status(); got.State != StateUpdateAvailable {
		t.Errorf("a verdict was lost to a concurrent failure: %+v", got)
	}
}

// updateRecordingLogger logs at Debug, the level this package speaks at, into a readable list.
func updateRecordingLogger() (*logrus.Entry, *[]string) {
	recorded := &[]string{}
	base := logrus.New()
	base.SetLevel(logrus.DebugLevel)
	base.SetOutput(io.Discard)
	base.AddHook(&updateCollector{recorded: recorded})
	return base.WithField("component", "update"), recorded
}

// updateCollector is locked because logrus fires hooks outside its own lock.
type updateCollector struct {
	mu       sync.Mutex
	recorded *[]string
}

func (c *updateCollector) Levels() []logrus.Level { return logrus.AllLevels }

func (c *updateCollector) Fire(entry *logrus.Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	*c.recorded = append(*c.recorded, entry.Message)
	return nil
}
