package hardening

import (
	"os"

	"pentagi/cmd/installer/state"
)

func DoSyncNetworkSettings(s state.State) error {
	// sync HTTP_PROXY or HTTPS_PROXY to PROXY_URL if they are set in the OS
	httpProxy, httpProxyExists := os.LookupEnv("HTTP_PROXY")
	if httpProxyExists && httpProxy != "" {
		if err := s.SetVar("PROXY_URL", httpProxy); err != nil {
			return err
		}
	}

	httpsProxy, httpsProxyExists := os.LookupEnv("HTTPS_PROXY")
	if httpsProxyExists && httpsProxy != "" {
		if err := s.SetVar("PROXY_URL", httpsProxy); err != nil {
			return err
		}
	}

	dockerEnvVarsNames := []string{
		"DOCKER_HOST",
		"DOCKER_TLS_VERIFY",
		"DOCKER_CERT_PATH",
	}

	vars, exists := s.GetVars(dockerEnvVarsNames)
	for _, envVar := range dockerEnvVarsNames {
		if exists[envVar] && vars[envVar].Value != "" {
			return nil // redefine is allowed only for unset docker connection settings
		}
	}

	// get the environment variables from the OS
	isOSDockerEnvVarsSet := false
	osDockerEnvVars := make(map[string]string, len(dockerEnvVarsNames))
	for _, envVar := range dockerEnvVarsNames {
		value, exists := os.LookupEnv(envVar)
		osDockerEnvVars[envVar] = value // set even empty value to avoid inconsistency while setting vars
		if exists && value != "" {
			isOSDockerEnvVarsSet = true
		}
	}

	// do nothing if the OS docker environment variables are not set (use defaults)
	if !isOSDockerEnvVarsSet {
		return nil
	}

	// sync all variables in the state at the same time to avoid inconsistencies
	return s.SetVars(osDockerEnvVars)
}
