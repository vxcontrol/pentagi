package docker

import (
	"fmt"
	"math/bits"

	"github.com/moby/moby/api/types/container"
)

type workerResourceLimits struct {
	nanoCPUs int64
	memory   int64
}

func (l workerResourceLimits) apply(h *container.HostConfig) error {
	if l.nanoCPUs > 0 {
		nanoCPUs := lowerPositiveLimit(h.NanoCPUs, l.nanoCPUs)
		if h.CPUQuota != 0 || h.CPUPeriod != 0 {
			period, err := workerCPUPeriod(h.CPUPeriod)
			if err != nil {
				return err
			}
			quota := workerCPUQuota(nanoCPUs, period)
			if quota >= 1000 {
				h.CPUQuota = lowerPositiveLimit(h.CPUQuota, quota)
				h.CPUPeriod, h.NanoCPUs = period, 0
			} else {
				// Docker's fixed NanoCPU period can express this ceiling while a
				// short caller period would require an invalid sub-1000us quota.
				h.NanoCPUs, h.CPUQuota, h.CPUPeriod = nanoCPUs, 0, 0
			}
		} else {
			h.NanoCPUs = nanoCPUs
		}
	}
	if l.memory > 0 {
		h.Memory = lowerPositiveLimit(h.Memory, l.memory)
		h.MemorySwap = h.Memory
		if h.MemoryReservation > h.Memory {
			h.MemoryReservation = h.Memory
		}
	}
	return nil
}

func (l workerResourceLimits) check(h *container.HostConfig) error {
	if l.nanoCPUs == 0 && l.memory == 0 {
		return nil
	}
	if h == nil {
		return fmt.Errorf("worker host configuration is unavailable")
	}
	if l.nanoCPUs > 0 {
		bounded := h.NanoCPUs > 0 && h.NanoCPUs <= l.nanoCPUs
		if h.CPUQuota > 0 {
			period, err := workerCPUPeriod(h.CPUPeriod)
			if err != nil {
				return err
			}
			bounded = bounded || h.CPUQuota <= workerCPUQuota(l.nanoCPUs, period)
		}
		if !bounded {
			return fmt.Errorf("worker exceeds configured CPU ceiling")
		}
	}
	if l.memory > 0 && (h.Memory <= 0 || h.Memory > l.memory || h.MemorySwap != h.Memory) {
		return fmt.Errorf("worker exceeds configured memory ceiling or permits swap")
	}
	return nil
}

func workerCPUPeriod(period int64) (int64, error) {
	if period == 0 {
		return 100000, nil
	}
	if period < 1000 || period > 1000000 {
		return 0, fmt.Errorf("worker CPU period must be between 1000 and 1000000 microseconds")
	}
	return period, nil
}

// Use an exact, overflow-safe product so a rounded ratio cannot reject a
// worker created at the configured ceiling or round a quota above that ceiling.
func workerCPUQuota(nanoCPUs, period int64) int64 {
	hi, lo := bits.Mul64(uint64(nanoCPUs), uint64(period))
	quota, _ := bits.Div64(hi, lo, 1_000_000_000)
	return int64(quota)
}

func lowerPositiveLimit(current, ceiling int64) int64 {
	if current > 0 && current < ceiling {
		return current
	}
	return ceiling
}
