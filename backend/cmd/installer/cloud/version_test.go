package cloud

import (
	"testing"

	"github.com/vxcontrol/cloud/models"
)

func TestVersion_NormalizeVersion_SendsOnlyWhatTheContractAccepts(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		// A revision suffix identifies a branch build; dropping it would pass it off as the release.
		{name: "release version", raw: "2.0.0", want: "2.0.0"},
		{name: "branch build with revision", raw: "2.0.0-87ac00f", want: "2.0.0-87ac00f"},
		{name: "leading v is a display convention", raw: "v2.0.0", want: "2.0.0"},
		{name: "surrounding whitespace", raw: "  2.0.0  ", want: "2.0.0"},
		{name: "two-part version", raw: "2.0", want: "2.0"},
		{name: "build metadata", raw: "2.0.0+build.5", want: "2.0.0+build.5"},

		// Everything a build with no release behind it can be named; sending any of
		// them gets the request refused.
		{name: "edition build naming its base version", raw: "2.0.0-ce.h87ac00f", want: "2.0.0-ce.h87ac00f"},
		{name: "enterprise edition build", raw: "2.0.0-ee.h87ac00f", want: "2.0.0-ee.h87ac00f"},
		{name: "edition release", raw: "2.0.0-ce", want: "2.0.0-ce"},
		{name: "edition alone is not a version", raw: "ce", want: "0.0.0"},
		{name: "edition and revision without a release", raw: "ce.h87ac00f", want: "0.0.0"},
		{name: "empty", raw: "", want: "0.0.0"},
		{name: "whitespace only", raw: "   ", want: "0.0.0"},
		{name: "bare commit hash", raw: "87ac00f", want: "0.0.0"},
		{name: "branch path", raw: "feature/some-branch", want: "0.0.0"},
		{name: "four-part version", raw: "2.0.0.1", want: "0.0.0"},
		{name: "five-part version", raw: "1.2.3.4.5", want: "0.0.0"},
		{name: "word with an underscore", raw: "release_candidate", want: "0.0.0"},
		{name: "bare v", raw: "v", want: "0.0.0"},
		{name: "dots only", raw: "..", want: "0.0.0"},
		// "latest" is valid to ask a package endpoint for, and meaningless as the running installer.
		{name: "latest is never an installer version", raw: "latest", want: "0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeVersion(tt.raw)
			if got != tt.want {
				t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.raw, got, tt.want)
			}
			if err := models.GetValidator().Var(got, "semver"); err != nil {
				t.Errorf("NormalizeVersion(%q) = %q, which the contract rejects: %v", tt.raw, got, err)
			}
		})
	}
}
