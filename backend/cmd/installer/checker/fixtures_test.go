package checker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"pentagi/cmd/installer/loader"

	"github.com/moby/moby/client"
)

// composeFiles are the deployment descriptors the installer ships and extracts.
var composeFiles = []string{
	"../files/fs/docker-compose.yml",
	"../files/fs/docker-compose-graphiti.yml",
	"../files/fs/docker-compose-langfuse.yml",
	"../files/fs/docker-compose-observability.yml",
}

type mockState struct {
	vars    map[string]loader.EnvVar
	envPath string
}

func (m *mockState) GetVar(key string) (loader.EnvVar, bool) {
	if val, exists := m.vars[key]; exists {
		return val, true
	}
	return loader.EnvVar{}, false
}

func (m *mockState) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
	return m.vars, make(map[string]bool, len(names))
}

func (m *mockState) GetEnvPath() string                     { return m.envPath }
func (m *mockState) Exists() bool                           { return true }
func (m *mockState) Reset() error                           { return nil }
func (m *mockState) Commit() error                          { return nil }
func (m *mockState) IsDirty() bool                          { return false }
func (m *mockState) GetEulaConsent() bool                   { return true }
func (m *mockState) SetEulaConsent() error                  { return nil }
func (m *mockState) SetStack(stack []string) error          { return nil }
func (m *mockState) GetStack() []string                     { return []string{} }
func (m *mockState) SetVar(name, value string) error        { return nil }
func (m *mockState) ResetVar(name string) error             { return nil }
func (m *mockState) SetVars(vars map[string]string) error   { return nil }
func (m *mockState) ResetVars(names []string) error         { return nil }
func (m *mockState) GetAllVars() map[string]loader.EnvVar   { return m.vars }
func (m *mockState) WriteVars(vars map[string]string) error { return nil }

var checkerAPIVersion = regexp.MustCompile(`^/v[0-9.]+`)

// checkerFakeDocker is a daemon answering only its "METHOD /path" routes; anything else is a missing object.
func checkerFakeDocker(t *testing.T, routes map[string]http.HandlerFunc) *client.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := routes[r.Method+" "+checkerAPIVersion.ReplaceAllString(r.URL.Path, "")]; ok {
			handler(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"message":"no such object: %s"}`, r.URL.Path)
	}))
	t.Cleanup(server.Close)

	cli, err := client.New(client.WithHost("tcp://"+server.Listener.Addr().String()), client.WithAPIVersion("1.47"))
	if err != nil {
		t.Fatalf("docker client: %v", err)
	}
	t.Cleanup(func() { cli.Close() })
	return cli
}

// checkerJSON answers a route with a fixed status and JSON body.
func checkerJSON(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}
