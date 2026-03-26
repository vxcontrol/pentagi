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

	// migration from LLM_SERVER_CONFIG_PATH to PENTAGI_LLM_SERVER_CONFIG_PATH
	llmServerConfigPathVar, exists := s.GetVar("LLM_SERVER_CONFIG_PATH")
	llmServerConfigPath := llmServerConfigPathVar.Value
	isEmbeddedCustomConfig := slices.Contains(configsPath, llmServerConfigPath) ||
		llmServerConfigPath == controller.DefaultCustomConfigsPath
	if exists && !isEmbeddedCustomConfig && llmServerConfigPath != "" {
		exists = checkPathInHostFS(llmServerConfigPath, file)
	}
	if exists && !isEmbeddedCustomConfig && llmServerConfigPath != "" {
		if err := s.SetVar("PENTAGI_LLM_SERVER_CONFIG_PATH", llmServerConfigPath); err != nil {
			return err
		}
		if err := s.SetVar("LLM_SERVER_CONFIG_PATH", controller.DefaultCustomConfigsPath); err != nil {
			return err
		}
	}

	// migration from OLLAMA_SERVER_CONFIG_PATH to PENTAGI_OLLAMA_SERVER_CONFIG_PATH
	ollamaServerConfigPathVar, exists := s.GetVar("OLLAMA_SERVER_CONFIG_PATH")
	ollamaServerConfigPath := ollamaServerConfigPathVar.Value
	isEmbeddedOllamaConfig := slices.Contains(configsPath, ollamaServerConfigPath) ||
		ollamaServerConfigPath == controller.DefaultOllamaConfigsPath
	if exists && !isEmbeddedOllamaConfig && ollamaServerConfigPath != "" {
		exists = checkPathInHostFS(ollamaServerConfigPath, file)
	}
	if exists && !isEmbeddedOllamaConfig && ollamaServerConfigPath != "" {
		if err := s.SetVar("PENTAGI_OLLAMA_SERVER_CONFIG_PATH", ollamaServerConfigPath); err != nil {
			return err
		}
		if err := s.SetVar("OLLAMA_SERVER_CONFIG_PATH", controller.DefaultOllamaConfigsPath); err != nil {
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
