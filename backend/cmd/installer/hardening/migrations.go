package hardening

import (
	"os"
	"slices"

	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
)

type checkPathType string

const (
	directory checkPathType = "directory"
	file      checkPathType = "file"
)

func DoMigrateSettings(s state.State) error {
	// migration from DOCKER_CERT_PATH to PENTAGI_DOCKER_CERT_PATH
	dockerCertPathVar, exists := s.GetVar("DOCKER_CERT_PATH")
	dockerCertPath := dockerCertPathVar.Value
	if exists && dockerCertPath != "" {
		exists = checkPathInHostFS(dockerCertPath, directory)
	}
	if exists && dockerCertPath != "" && dockerCertPath != controller.DefaultDockerCertPath {
		if err := s.SetVar("PENTAGI_DOCKER_CERT_PATH", dockerCertPath); err != nil {
			return err
		}
		if err := s.SetVar("DOCKER_CERT_PATH", controller.DefaultDockerCertPath); err != nil {
			return err
		}
	}

	configsPath := controller.GetEmbeddedLLMConfigsPath(files.NewFiles())

	// A provider config variable that names a file on the host becomes the mount source,
	// and the variable itself the path that file is mounted at in the container.
	for _, config := range []struct{ variable, mountedPath string }{
		{"LLM_SERVER_CONFIG_PATH", controller.DefaultCustomConfigsPath},
		{"OLLAMA_SERVER_CONFIG_PATH", controller.DefaultOllamaConfigsPath},
		{"BEDROCK_CONFIG_PATH", controller.DefaultBedrockConfigsPath},
	} {
		configPathVar, exists := s.GetVar(config.variable)
		configPath := configPathVar.Value
		isInContainer := slices.Contains(configsPath, configPath) || configPath == config.mountedPath
		if !exists || isInContainer || configPath == "" || !checkPathInHostFS(configPath, file) {
			continue
		}
		if err := s.SetVar("PENTAGI_"+config.variable, configPath); err != nil {
			return err
		}
		if err := s.SetVar(config.variable, config.mountedPath); err != nil {
			return err
		}
	}

	return nil
}

func checkPathInHostFS(path string, pathType checkPathType) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	switch pathType {
	case directory:
		return info.IsDir()
	case file:
		return !info.IsDir()
	default:
		return false
	}
}
