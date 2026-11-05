package agent

import (
	"context"
	"fmt"
	"go-metrics/internal/model"
	"go-metrics/internal/repository"
	"math/rand"
	"reflect"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type collector struct {
	storage *repository.MemStorage
	mutex   *sync.Mutex
}

func newCollector() *collector {
	return &collector{
		storage: repository.NewMemStorage(),
		mutex:   &sync.Mutex{},
	}
}

func (collector *collector) CollectMemMetrics(ctx context.Context) error {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	counter, _ := collector.storage.GetOrRegister(ctx, model.Counter, "PollCount")
	counter.UpdateDelta(1)
	_ = collector.storage.Save(ctx, counter)
	random, _ := collector.storage.GetOrRegister(ctx, model.Gauge, "RandomValue")
	random.UpdateValue(rand.Float64() * 100)
	_ = collector.storage.Save(ctx, random)

	fields := [...]string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}

	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	refl := reflect.ValueOf(memStats).Elem()

	for _, fieldName := range fields {
		field := refl.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("field '%s' not found in metrics", fieldName)
		}

		value := float64(0)

		if field.CanFloat() {
			value = field.Float()
		} else if field.CanInt() {
			value = float64(field.Int())
		} else if field.CanUint() {
			value = float64(field.Uint())
		}

		metric, _ := collector.storage.GetOrRegister(ctx, model.Gauge, fieldName)
		metric.UpdateValue(value)
		_ = collector.storage.Save(ctx, random)
	}

	return nil
}

func (collector *collector) CollectGopsutilMetrics(ctx context.Context) error {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	totalMem, _ := collector.storage.GetOrRegister(ctx, model.Gauge, "TotalMemory")
	totalMem.UpdateValue(float64(vmStat.Total))
	_ = collector.storage.Save(ctx, totalMem)

	freeMem, _ := collector.storage.GetOrRegister(ctx, model.Gauge, "FreeMemory")
	freeMem.UpdateValue(float64(vmStat.Total))
	_ = collector.storage.Save(ctx, freeMem)

	cpusUtilization, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	for i, cpuPercent := range cpusUtilization {
		metric, _ := collector.storage.GetOrRegister(ctx, model.Gauge, fmt.Sprintf("CPUutilization%d", i+1))
		metric.UpdateValue(cpuPercent)
		_ = collector.storage.Save(ctx, metric)
	}

	return nil
}

func (collector *collector) GetSnapshot(ctx context.Context) []*model.Metric {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	metrics, _ := collector.storage.All(ctx)
	snapshot := make([]*model.Metric, len(metrics))

	for i, metric := range metrics {
		snapshot[i] = metric.Clone()
	}

	counter, _ := collector.storage.GetOrRegister(ctx, model.Counter, "PollCount")
	counter.UpdateDelta(-(*counter.Delta))
	_ = collector.storage.Save(ctx, counter)

	return snapshot
}
