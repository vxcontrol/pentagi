package cloud

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		// A release build and a branch build both have to round-trip unchanged: the
		// revision suffix is what identifies a build cut from a branch, and dropping it
		// would make every such build look like the release it branched from.
		{name: "release version", raw: "2.0.0", want: "2.0.0"},
		{name: "branch build with revision", raw: "2.0.0-87ac00f", want: "2.0.0-87ac00f"},
		{name: "leading v is a display convention", raw: "v2.0.0", want: "2.0.0"},
		{name: "surrounding whitespace", raw: "  2.0.0  ", want: "2.0.0"},
		{name: "two-part version", raw: "2.0", want: "2.0"},
		{name: "build metadata", raw: "2.0.0+build.5", want: "2.0.0+build.5"},

		// Everything a development build can be named. None of these is a version, and
		// sending them gets the whole request refused.
		{name: "branch name", raw: "develop", want: UnknownVersion},
		{name: "branch name with revision", raw: "develop-87ac00f", want: UnknownVersion},
		{name: "empty", raw: "", want: UnknownVersion},
		{name: "whitespace only", raw: "   ", want: UnknownVersion},
		{name: "bare commit hash", raw: "87ac00f", want: UnknownVersion},

		// "latest" is a valid thing to ask a package endpoint for, and a meaningless
		// description of the installer you are running.
		{name: "latest is never an installer version", raw: "latest", want: UnknownVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeVersion(tt.raw); got != tt.want {
				t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// TestNormalizedVersionAlwaysPassesTheContract is the invariant the table above only
// samples: whatever a build calls itself, what leaves the installer is acceptable. Without
// it, a new build-naming scheme would be caught in production rather than here.
func TestNormalizedVersionAlwaysPassesTheContract(t *testing.T) {
	raws := []string{
		"develop", "develop-87ac00f", "", "   ", "latest", "v2.0.0", "2.0.0-87ac00f",
		"feature/some-branch", "2.0.0.1", "release_candidate", "v", "..", "1.2.3.4.5",
	}

	for _, raw := range raws {
		normalized := NormalizeVersion(raw)
		if err := validateSemver(normalized); err != nil {
			t.Errorf("NormalizeVersion(%q) = %q, which the contract rejects: %v", raw, normalized, err)
		}
	}
}
