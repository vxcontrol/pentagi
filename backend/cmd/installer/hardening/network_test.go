package hardening

import (
	"errors"
	"maps"
	"os"
	"testing"

	"pentagi/cmd/installer/loader"
)

// networkEnv unsets every variable DoSyncNetworkSettings reads, then sets the given ones.
func networkEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for _, name := range []string{
		"HTTP_PROXY", "HTTPS_PROXY", "DOCKER_HOST", "DOCKER_TLS_VERIFY", "DOCKER_CERT_PATH", "PENTAGI_DOCKER_CERT_PATH",
	} {
		t.Setenv(name, "")
		os.Unsetenv(name)
	}
	for name, value := range env {
		t.Setenv(name, value)
	}
}

func TestNetwork_DoSyncNetworkSettings_CopiesTheHostProxyAndDockerSettings(t *testing.T) {
	const (
		httpProxy  = "http://proxy.example.com:8080"
		httpsProxy = "https://proxy.example.com:8443"
		host       = "tcp://docker.example.com:2376"
		newHost    = "tcp://new.example.com:2376"
		certs      = "/path/to/certs"
	)
	dir, file := hardeningHostPaths(t)
	tests := []struct {
		name  string
		state map[string]string // already in the state before the sync
		env   map[string]string // the host environment; an absent name is unset
		want  map[string]string // the whole state afterwards
	}{
		{"an http proxy", nil, map[string]string{"HTTP_PROXY": httpProxy}, map[string]string{"PROXY_URL": httpProxy}},
		{"an https proxy", nil, map[string]string{"HTTPS_PROXY": httpsProxy}, map[string]string{"PROXY_URL": httpsProxy}},
		{"the https proxy wins over the http proxy", nil,
			map[string]string{"HTTP_PROXY": httpProxy, "HTTPS_PROXY": httpsProxy}, map[string]string{"PROXY_URL": httpsProxy}},
		{"an empty http proxy is ignored", nil, map[string]string{"HTTP_PROXY": ""}, map[string]string{}},
		{"an empty https proxy is ignored", nil, map[string]string{"HTTPS_PROXY": ""}, map[string]string{}},
		{"nothing set", nil, map[string]string{}, map[string]string{}},
		{"every docker variable", nil,
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "1", "DOCKER_CERT_PATH": certs},
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "1", "DOCKER_CERT_PATH": certs, "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"a docker host alone syncs the others empty", nil, map[string]string{"DOCKER_HOST": host},
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": "", "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"tls verify alone", nil, map[string]string{"DOCKER_TLS_VERIFY": "1"},
			map[string]string{"DOCKER_HOST": "", "DOCKER_TLS_VERIFY": "1", "DOCKER_CERT_PATH": "", "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"a certificates path alone", nil, map[string]string{"DOCKER_CERT_PATH": certs},
			map[string]string{"DOCKER_HOST": "", "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": certs, "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"empty docker variables are ignored", nil,
			map[string]string{"DOCKER_HOST": "", "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": ""}, map[string]string{}},
		{"set and empty docker variables together", nil,
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": certs},
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": certs, "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"a proxy and docker together", nil,
			map[string]string{"HTTP_PROXY": httpProxy, "HTTPS_PROXY": httpsProxy, "DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "1", "DOCKER_CERT_PATH": certs},
			map[string]string{"PROXY_URL": httpsProxy, "DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "1", "DOCKER_CERT_PATH": certs, "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"whitespace is copied verbatim", nil,
			map[string]string{"HTTP_PROXY": "   ", "HTTPS_PROXY": "\t\n", "DOCKER_HOST": "   ", "DOCKER_TLS_VERIFY": "\t", "DOCKER_CERT_PATH": "\n"},
			map[string]string{"PROXY_URL": "\t\n", "DOCKER_HOST": "   ", "DOCKER_TLS_VERIFY": "\t", "DOCKER_CERT_PATH": "\n", "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"a host certificates directory moves to the pentagi variable", nil,
			map[string]string{"DOCKER_HOST": host, "DOCKER_CERT_PATH": dir},
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": "/opt/pentagi/docker/ssl", "PENTAGI_DOCKER_CERT_PATH": dir}},
		{"a certificates file is not moved", nil,
			map[string]string{"DOCKER_HOST": host, "DOCKER_CERT_PATH": file},
			map[string]string{"DOCKER_HOST": host, "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": file, "PENTAGI_DOCKER_CERT_PATH": ""}},
		{"a docker host in the state is kept", map[string]string{"DOCKER_HOST": host},
			map[string]string{"DOCKER_HOST": newHost}, map[string]string{"DOCKER_HOST": host}},
		{"tls verify in the state blocks the sync", map[string]string{"DOCKER_TLS_VERIFY": "1"},
			map[string]string{"DOCKER_HOST": newHost}, map[string]string{"DOCKER_TLS_VERIFY": "1"}},
		{"a certificates path in the state blocks the sync", map[string]string{"DOCKER_CERT_PATH": "/existing/certs"},
			map[string]string{"DOCKER_HOST": newHost}, map[string]string{"DOCKER_CERT_PATH": "/existing/certs"}},
		{"a pentagi certificates path in the state blocks the sync", map[string]string{"PENTAGI_DOCKER_CERT_PATH": "/existing/certs"},
			map[string]string{"DOCKER_HOST": newHost}, map[string]string{"PENTAGI_DOCKER_CERT_PATH": "/existing/certs"}},
		{"empty docker values in the state do not block the sync",
			map[string]string{"DOCKER_HOST": "", "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": ""},
			map[string]string{"DOCKER_HOST": newHost},
			map[string]string{"DOCKER_HOST": newHost, "DOCKER_TLS_VERIFY": "", "DOCKER_CERT_PATH": "", "PENTAGI_DOCKER_CERT_PATH": ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkEnv(t, tt.env)
			st := &mockState{vars: map[string]loader.EnvVar{}}
			for name, value := range tt.state {
				st.vars[name] = loader.EnvVar{Name: name, Value: value, Line: 1}
			}

			if err := DoSyncNetworkSettings(st); err != nil {
				t.Fatalf("DoSyncNetworkSettings: %v", err)
			}

			got := map[string]string{}
			for name, envVar := range st.vars {
				got[name] = envVar.Value
				if _, before := tt.state[name]; !before && !envVar.IsChanged {
					t.Errorf("%s was added without being marked changed", name)
				}
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("state = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNetwork_DoSyncNetworkSettings_ReturnsTheStateError(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		st   *mockState
	}{
		{"the http proxy cannot be written", map[string]string{"HTTP_PROXY": "http://proxy.example.com:8080"},
			&mockState{setVarError: map[string]error{"PROXY_URL": errMocked}}},
		{"the https proxy cannot be written", map[string]string{"HTTPS_PROXY": "https://proxy.example.com:8443"},
			&mockState{setVarError: map[string]error{"PROXY_URL": errMocked}}},
		{"the docker settings cannot be written", map[string]string{"DOCKER_HOST": "tcp://docker.example.com:2376"},
			&mockState{setVarsError: errMocked}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networkEnv(t, tt.env)

			if err := DoSyncNetworkSettings(tt.st); !errors.Is(err, errMocked) {
				t.Errorf("err = %v, want %v", err, errMocked)
			}
		})
	}
}
