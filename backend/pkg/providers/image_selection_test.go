package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePentestImage(t *testing.T) {
	tests := []struct {
		name          string
		pentestImage  string
		defaultImage  string
		expectedImage string
	}{
		{
			name:          "uses configured pentest image",
			pentestImage:  "vxcontrol/kali-linux",
			defaultImage:  "debian:latest",
			expectedImage: "vxcontrol/kali-linux",
		},
		{
			name:          "normalizes case and whitespace",
			pentestImage:  "  VxControl/Kali-Linux\n",
			defaultImage:  "debian:latest",
			expectedImage: "vxcontrol/kali-linux",
		},
		{
			name:          "falls back to the default image when pentest image is empty",
			pentestImage:  "",
			defaultImage:  "debian:latest",
			expectedImage: "debian:latest",
		},
		{
			name:          "falls back to the normalized default image",
			pentestImage:  "  ",
			defaultImage:  "  Debian:Latest ",
			expectedImage: "debian:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedImage, resolvePentestImage(tt.pentestImage, tt.defaultImage))
		})
	}
}
