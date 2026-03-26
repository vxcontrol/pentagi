package observability

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

const defCollectPeriod = 10 * time.Second

func startProcessMetricCollect(meter otelmetric.Meter, attrs []attribute.KeyValue) error {
	proc := process.Process{
		Pid: int32(os.Getpid()),
	}

	collectRssMem := func(ctx context.Context, m otelmetric.Int64Observer) error {
		procMemInfo, err := proc.MemoryInfoWithContext(ctx)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("failed to get process resident memory")
			return fmt.Errorf("failed to get process resident memory: %w", err)
		}
		m.Observe(int64(procMemInfo.RSS), otelmetric.WithAttributes(attrs...))
		return nil
	}
	collectVirtMem := func(ctx context.Context, m otelmetric.Int64Observer) error {
		procMemInfo, err := proc.MemoryInfoWithContext(ctx)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("failed to get process virtual memory")
			return fmt.Errorf("failed to get process virtual memory: %w", err)
		}
		m.Observe(int64(procMemInfo.VMS), otelmetric.WithAttributes(attrs...))
		return nil
	}
	collectCpuPercent := func(ctx context.Context, m otelmetric.Float64Observer) error {
		procCpuPercent, err := proc.PercentWithContext(ctx, time.Duration(0))
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("failed to get CPU usage percent")
			return fmt.Errorf("failed to get CPU usage percent: %w", err)
		}
		m.Observe(procCpuPercent, otelmetric.WithAttributes(attrs...))
		return nil
	}

	if _, err := proc.MemoryInfo(); err == nil {
		_, _ = meter.Int64ObservableGauge(
			"process_resident_memory_bytes",
			otelmetric.WithInt64Callback(collectRssMem),
		)
		_, _ = meter.Int64ObservableGauge(
			"process_virtual_memory_bytes",
			otelmetric.WithInt64Callback(collectVirtMem),
		)
	}
	if _, err := proc.Percent(time.Duration(0)); err == nil {
		_, _ = meter.Float64ObservableGauge(
			"process_cpu_usage_percent",
			otelmetric.WithFloat64Callback(collectCpuPercent),
		)
	}

	return nil
}

func startGoRuntimeMetricCollect(meter otelmetric.Meter, attrs []attribute.KeyValue) error {
	var (
		lastUpdate         time.Time = time.Now()
		mx                 sync.Mutex
		procRuntimeMemStat runtime.MemStats
	)
	runtime.ReadMemStats(&procRuntimeMemStat)

	getMemStats := func() *runtime.MemStats {
		mx.Lock()
		defer mx.Unlock()

		now := time.Now()
		if now.Sub(lastUpdate) > defCollectPeriod {
			runtime.ReadMemStats(&procRuntimeMemStat)
		}
		lastUpdate = now
		return &procRuntimeMemStat
	}

	meter.Int64ObservableGauge("go_cgo_calls",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(runtime.NumCgoCall(), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_goroutines",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(runtime.NumGoroutine()), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_heap_objects_bytes",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().HeapInuse), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_heap_objects_counter",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().HeapObjects), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_stack_inuse_bytes",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().StackInuse), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_stack_sys_bytes",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().StackSys), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_total_allocs_bytes",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().TotalAlloc), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_heap_allocs_bytes",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().HeapAlloc), otelmetric.WithAttributes(attrs...))
			return nil
		}))
	meter.Int64ObservableGauge("go_pause_gc_total_nanosec",
		otelmetric.WithInt64Callback(func(ctx context.Context, m otelmetric.Int64Observer) error {
			m.Observe(int64(getMemStats().PauseTotalNs), otelmetric.WithAttributes(attrs...))
			return nil
		}))

	return nil
}

func startDumperMetricCollect(stats Dumper, meter otelmetric.Meter, attrs []attribute.KeyValue) error {
	var (
		err        error
		lastStats  map[string]float64
		lastUpdate time.Time = time.Now()
		mx         sync.Mutex
	)

	if lastStats, err = stats.DumpStats(); err != nil {
		logrus.WithError(err).Errorf("failed to get stats dump")
		return err
	}

	getStats := func() map[string]float64 {
		mx.Lock()
		defer mx.Unlock()

		now := time.Now()
		if now.Sub(lastUpdate) <= defCollectPeriod {
			return lastStats
		}
		if lastStats, err = stats.DumpStats(); err != nil {
			return lastStats
		}
		lastUpdate = now
		return lastStats
	}

	for key := range lastStats {
		metricName := key
		_, _ = meter.Float64ObservableCounter(metricName,
			otelmetric.WithFloat64Callback(func(ctx context.Context, m otelmetric.Float64Observer) error {
				if value, ok := getStats()[metricName]; ok {
					m.Observe(value, otelmetric.WithAttributes(attrs...))
					return nil
				}
				return fmt.Errorf("metric '%s' not found", metricName)
			}))
	}

	return nil
}
