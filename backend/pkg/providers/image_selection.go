package providers

import (
	"strings"

	"pentagi/pkg/config"

	"github.com/distribution/reference"
)

// ImageSelectionModeFixed skips the model call and always uses the configured
// pentest image. Any other value keeps the model choosing the image.
const ImageSelectionModeFixed = "fixed"

type imagePolicy struct {
	mode     string
	pentest  string
	fallback string
	allowed  []string
}

func newImagePolicy(cfg *config.Config, defaultImage string) imagePolicy {
	policy := imagePolicy{fallback: normalizeImage(defaultImage)}
	if cfg == nil {
		return policy
	}

	policy.mode = strings.ToLower(strings.TrimSpace(cfg.DockerImageSelectionMode))
	policy.pentest = normalizeImage(cfg.DockerDefaultImageForPentest)
	for _, name := range strings.Split(cfg.DockerAllowedImages, ",") {
		if name = normalizeImage(name); name != "" {
			policy.allowed = append(policy.allowed, name)
		}
	}

	return policy
}

func normalizeImage(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (p imagePolicy) pentestImage() string {
	if p.pentest != "" {
		return p.pentest
	}
	return p.fallback
}

func (p imagePolicy) skipsModel() bool {
	return p.mode == ImageSelectionModeFixed
}

func (p imagePolicy) resolve(answer string) (selected string, rejected bool) {
	image := normalizeImage(answer)
	if image == "" {
		return p.pentestImage(), true
	}

	if _, err := reference.ParseNormalizedNamed(image); err != nil {
		return p.pentestImage(), true
	}

	if len(p.allowed) == 0 {
		return image, false
	}

	for _, name := range p.allowed {
		if image == name || sameRepository(image, name) {
			return image, false
		}
	}

	return p.pentestImage(), true
}

func sameRepository(image, allowed string) bool {
	allowedRef, err := reference.ParseNormalizedNamed(allowed)
	if err != nil {
		return false
	}

	if _, tagged := allowedRef.(reference.Tagged); tagged {
		return false
	}

	parsed, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return false
	}

	return reference.FamiliarName(parsed) == reference.FamiliarName(allowedRef)
}

func (fp *flowProvider) isPentestImage() bool {
	pentest := pentestDockerImage
	if fp.cfg != nil {
		if configured := normalizeImage(fp.cfg.DockerDefaultImageForPentest); configured != "" {
			pentest = configured
		}
	}

	return strings.HasPrefix(normalizeImage(fp.image), pentest)
}
