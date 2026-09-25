package providers

import (
	"testing"

	"pentagi/pkg/config"

	"github.com/stretchr/testify/assert"
)

func TestImageSelection_Resolve_FallsBackUnlessTheAnswerIsAnAllowedImage(t *testing.T) {
	t.Parallel()

	kali := &config.Config{
		DockerDefaultImageForPentest: "vxcontrol/kali-linux",
		DockerAllowedImages:          "vxcontrol/kali-linux, debian:latest",
	}
	private := &config.Config{
		DockerDefaultImageForPentest: "vxcontrol/kali-linux",
		DockerAllowedImages:          "registry.internal:5000/kali, docker.io/library/debian",
	}
	open := &config.Config{DockerDefaultImageForPentest: "vxcontrol/kali-linux"}

	tests := []struct {
		name         string
		cfg          *config.Config
		defaultImage string
		answer       string
		image        string
		rejected     bool
	}{
		{"a listed image passes", kali, "debian:latest", "debian:latest", "debian:latest", false},
		{"a listed repository admits any tag", kali, "debian:latest", "vxcontrol/kali-linux:2026.1", "vxcontrol/kali-linux:2026.1", false},
		{"an unlisted image falls back", kali, "debian:latest", "node:latest", "vxcontrol/kali-linux", true},
		{"prose instead of a name falls back", kali, "debian:latest", "I would use kali linux here", "vxcontrol/kali-linux", true},
		{"an empty answer falls back", kali, "debian:latest", "   ", "vxcontrol/kali-linux", true},
		{"case and spacing are normalized", kali, "debian:latest", "  Debian:Latest\n", "debian:latest", false},
		{"a repository behind a port admits its tags", private, "debian:latest", "registry.internal:5000/kali:2026.1", "registry.internal:5000/kali:2026.1", false},
		{"the same repository untagged passes", private, "debian:latest", "registry.internal:5000/kali", "registry.internal:5000/kali", false},
		{"a fully qualified entry admits the familiar name", private, "debian:latest", "debian:latest", "debian:latest", false},
		{"another repository on the same registry falls back", private, "debian:latest", "registry.internal:5000/other:1", "vxcontrol/kali-linux", true},
		{
			"an allow-list entry that is not an image reference admits nothing",
			&config.Config{DockerDefaultImageForPentest: "vxcontrol/kali-linux", DockerAllowedImages: "Not An Image!"},
			"debian:latest", "node:latest", "vxcontrol/kali-linux", true,
		},
		{"no allow-list admits any image", open, "debian:latest", "python:latest", "python:latest", false},
		{"no allow-list still refuses what is not an image", open, "debian:latest", "not an image reference at all", "vxcontrol/kali-linux", true},
		{
			"without a pentest image the default image is the fallback",
			&config.Config{DockerAllowedImages: "debian:latest"},
			"  Debian:Latest ", "node:latest", "debian:latest", true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			image, rejected := newImagePolicy(tt.cfg, tt.defaultImage).resolve(tt.answer)
			assert.Equal(t, tt.image, image)
			assert.Equal(t, tt.rejected, rejected)
		})
	}
}

func TestImageSelection_PentestImage_PrefersTheConfiguredImageToTheDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cfg          *config.Config
		defaultImage string
		want         string
	}{
		{
			name:         "a configured pentest image",
			cfg:          &config.Config{DockerImageSelectionMode: ImageSelectionModeFixed, DockerDefaultImageForPentest: "vxcontrol/kali-linux"},
			defaultImage: "debian:latest",
			want:         "vxcontrol/kali-linux",
		},
		{
			name:         "no pentest image falls back to the normalized default",
			cfg:          &config.Config{DockerImageSelectionMode: ImageSelectionModeFixed, DockerAllowedImages: "debian:latest"},
			defaultImage: "  Debian:Latest ",
			want:         "debian:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, newImagePolicy(tt.cfg, tt.defaultImage).pentestImage())
		})
	}
}

func TestImageSelection_SkipsModel_OnlyInFixedMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode string
		want bool
	}{
		{name: "fixed mode", mode: ImageSelectionModeFixed, want: true},
		{name: "fixed mode written loosely", mode: " Fixed ", want: true},
		{name: "no mode keeps the model choosing", mode: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			policy := newImagePolicy(&config.Config{
				DockerImageSelectionMode:     tt.mode,
				DockerDefaultImageForPentest: "vxcontrol/kali-linux",
			}, "debian:latest")

			assert.Equal(t, tt.want, policy.skipsModel())
		})
	}
}

func TestImageSelection_IsPentestImage_FollowsTheConfiguredImage(t *testing.T) {
	t.Parallel()

	hardened := &config.Config{DockerDefaultImageForPentest: "registry.example.com/security/hardened-kali"}

	tests := []struct {
		name  string
		image string
		cfg   *config.Config
		want  bool
	}{
		{"a configured pentest image is recognised", "registry.example.com/security/hardened-kali:2026", hardened, true},
		{"the built-in name does not win over configuration", "vxcontrol/kali-linux", hardened, false},
		{"without configuration the built-in name applies", "vxcontrol/kali-linux:2026", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, (&flowProvider{image: tt.image, cfg: tt.cfg}).isPentestImage())
		})
	}
}
