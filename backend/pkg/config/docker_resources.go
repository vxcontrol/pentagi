package config

import (
	"fmt"
	"math"
)

// ValidateWorkerResourceLimits checks Docker's supported ranges before any worker is created.
func (c *Config) ValidateWorkerResourceLimits() error {
	if math.IsNaN(c.DockerWorkerCPUs) || math.IsInf(c.DockerWorkerCPUs, 0) ||
		c.DockerWorkerCPUs < 0 || (c.DockerWorkerCPUs > 0 && c.DockerWorkerCPUs < 0.01) ||
		c.DockerWorkerCPUs > float64(math.MaxInt64/1_000_000_000) {
		return fmt.Errorf("invalid DOCKER_WORKER_CPUS: must be 0 (disabled) or a finite value between 0.01 and %d", math.MaxInt64/1_000_000_000)
	}
	if c.DockerWorkerMemory < 0 || (c.DockerWorkerMemory > 0 && c.DockerWorkerMemory < 6<<20) {
		return fmt.Errorf("invalid DOCKER_WORKER_MEMORY: must be 0 (disabled) or at least 6291456 bytes")
	}
	return nil
}
