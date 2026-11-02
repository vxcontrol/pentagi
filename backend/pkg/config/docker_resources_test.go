package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerResources_NewConfig_RejectsInvalidWorkerBudgets(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"DOCKER_WORKER_CPUS", "-1"}, {"DOCKER_WORKER_CPUS", "NaN"}, {"DOCKER_WORKER_CPUS", "+Inf"},
		{"DOCKER_WORKER_CPUS", "0.001"}, {"DOCKER_WORKER_CPUS", "9223372037"},
		{"DOCKER_WORKER_MEMORY", "-1"}, {"DOCKER_WORKER_MEMORY", "1024"},
		{"DOCKER_WORKER_MEMORY", "9223372036854775808"},
	} {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			clearConfigEnv(t)
			t.Chdir(t.TempDir())
			t.Setenv("DOCKER_WORKER_CPUS", "0")
			t.Setenv("DOCKER_WORKER_MEMORY", "0")
			t.Setenv(tc.key, tc.value)
			_, err := NewConfig()
			if tc.value == "9223372036854775808" {
				require.ErrorContains(t, err, "out of range")
			} else {
				require.ErrorContains(t, err, tc.key)
			}
		})
	}
}

func TestDockerResources_NewConfig_AcceptsOptInAndDisabledBudgets(t *testing.T) {
	for _, tc := range []struct{ cpu, memory string }{{"0", "0"}, {"0.5", "268435456"}, {"0.01", "6291456"}} {
		t.Run(tc.cpu+" "+tc.memory, func(t *testing.T) {
			clearConfigEnv(t)
			t.Chdir(t.TempDir())
			t.Setenv("DOCKER_WORKER_CPUS", tc.cpu)
			t.Setenv("DOCKER_WORKER_MEMORY", tc.memory)
			_, err := NewConfig()
			require.NoError(t, err)
		})
	}
}
