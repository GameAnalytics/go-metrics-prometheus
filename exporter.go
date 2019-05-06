package prometheusmetrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
	"strings"
	"time"
)

type Exporter struct {
	namespace      string
	subsystem      string
	srcRegistry    metrics.Registry
	targetRegistry prometheus.Registerer
	flushInterval  time.Duration
	gauges         map[string]prometheus.Gauge
	_              struct{}
}

func NewExporter(
	namespace string,
	subsystem string,
	srcRegistry metrics.Registry,
	targetRegistry prometheus.Registerer,
	flushInterval time.Duration,
) *Exporter {
	return &Exporter{
		namespace:      ReplaceUnsafeKeyCharacters(namespace),
		subsystem:      ReplaceUnsafeKeyCharacters(subsystem),
		srcRegistry:    srcRegistry,
		targetRegistry: targetRegistry,
		flushInterval:  flushInterval,
		gauges:         make(map[string]prometheus.Gauge),
	}
}

func (e *Exporter) Run() {
	for _ = range time.Tick(e.flushInterval) {
		e.ExportOnce()
	}
}

func (e *Exporter) ExportOnce() {
	for metricName, values := range e.srcRegistry.GetAll() {
		for valueName, value := range values {
			e.getOrRegisterPrometheusGauge(metricName, valueName).Set(AsFloat64(value))
		}
	}
}

func (e *Exporter) getOrRegisterPrometheusGauge(metricName, valueName string) prometheus.Gauge {
	name := PrometheusMetricName(metricName, valueName)
	key := PrometheusMetricKey(e.namespace, e.subsystem, name, valueName)
	gauge, ok := e.gauges[key]
	if !ok {
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: e.namespace,
			Subsystem: e.subsystem,
			Name:      name,
		})
		e.targetRegistry.MustRegister(gauge)
		e.gauges[key] = gauge
	}
	return gauge
}

func AsFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case string: // Go Metrics health checks are sent as error strings
		if len(v) > 0 {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func PrometheusMetricName(metric, value string) string {
	base := fmt.Sprintf("%s_%s", metric, value)
	safe := ReplaceUnsafeKeyCharacters(base)
	clean := strings.TrimRight(safe, "_")
	return clean
}

func PrometheusMetricKey(namespace, subsystem, metric, valueName string) string {
	clean := fmt.Sprintf("%s_%s_%s", namespace, subsystem, PrometheusMetricName(metric, valueName))
	return clean
}

func ReplaceUnsafeKeyCharacters(key string) string {
	if key == "" {
		return "unnamed"
	}
	bs := []byte(key)
	for i := 0; i < len(bs); i++ {
		char := bs[i]
		// Turn '%' into `p` for when dealing with percentiles.
		if char == '%' {
			bs[i] = 'p'
			continue
		}
		// All non alphanumerics become underscores.
		valid := char >= 'A' && char <= 'Z' || // A-Z is ok
			char >= 'a' && char <= 'z' || // a-z is ok
			char >= '0' && char <= '9' // 0-9 is ok
		if !valid {
			bs[i] = '_'
		}
	}
	return string(bs)
}
