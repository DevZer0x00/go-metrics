package agent

import (
	"reflect"
	"runtime"

	"github.com/rs/zerolog/log"
)

type MemMetrics struct {
	Label string
	Value float64
}

func CollectMemMetrics() []MemMetrics {
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

	metrics := make([]MemMetrics, len(fields))

	refl := reflect.ValueOf(memStats).Elem()

	for index, fieldName := range fields {
		field := refl.FieldByName(fieldName)
		if !field.IsValid() {
			log.
				Fatal().
				Str("fieldName", fieldName).
				Msg("field not found in metrics")
		}

		value := float64(0)

		if field.CanFloat() {
			value = field.Float()
		} else if field.CanInt() {
			value = float64(field.Int())
		} else if field.CanUint() {
			value = float64(field.Uint())
		}

		metrics[index] = MemMetrics{
			Label: fieldName,
			Value: value,
		}
	}

	return metrics
}
