package prometheusmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	namespace     = "ns"
	subsystem     = "subsys"
	flushInterval = 1 * time.Second
)

var (
	sourceRegistry metrics.Registry
	targetRegistry *prometheus.Registry
	exporter       *Exporter
)

func TestMain(m *testing.M) {
	sourceRegistry = metrics.NewRegistry()
	targetRegistry = prometheus.NewRegistry()
	exporter = NewExporter(namespace, subsystem, sourceRegistry, targetRegistry, flushInterval)
	os.Exit(m.Run())
}

func TestExporter_ExportOnce_Counter(t *testing.T) {
	// Given an exporter with a registered counter of value 1,
	metrics.GetOrRegisterCounter("mycounter", sourceRegistry).Inc(1)
	// Then exporting to Prometheus,
	exporter.ExportOnce()
	// Should deliver 1 gauge of value 1.
	out, err := targetRegistry.Gather()
	assert.NoError(t, err)
	assert.Len(t, out, 1)
	assert.Equal(t, "ns_subsys_mycounter_count", *(out[0].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	assert.Equal(t, float64(1), *(out[0].Metric[0].Gauge.Value))

}

func TestExporter_ExportOnce_Gauge(t *testing.T) {
	// Given an exporter with a registered gauge of value 2,
	metrics.GetOrRegisterGauge("mygauge", sourceRegistry).Update(2)
	// Then exporting to Prometheus,
	exporter.ExportOnce()
	// Should deliver 1 gauge of value 2.
	out, err := targetRegistry.Gather()
	assert.NoError(t, err)
	assert.Len(t, out, 1)
	assert.Equal(t, "ns_subsys_mygauge_value", *(out[0].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	assert.Equal(t, float64(2), *(out[0].Metric[0].Gauge.Value))
}

func TestExporter_ExportOnce_Meter(t *testing.T) {
	// Given an exporter with a registered meter,
	metrics.GetOrRegisterMeter("mymeter", sourceRegistry).Mark(1)
	// Then exporting to Prometheus,
	exporter.ExportOnce()
	// Should deliver 5 gauges,
	out, err := targetRegistry.Gather()
	assert.NoError(t, err)
	assert.Len(t, out, 5)
	//
	assert.Equal(t, "ns_subsys_mymeter_15m_rate", *(out[0].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	assert.Equal(t, float64(0), *(out[0].Metric[0].Gauge.Value))
	//
	assert.Equal(t, "ns_subsys_mymeter_1m_rate", *(out[1].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[1].Type))
	assert.Equal(t, float64(0), *(out[1].Metric[0].Gauge.Value))
	//
	assert.Equal(t, "ns_subsys_mymeter_5m_rate", *(out[2].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[2].Type))
	assert.Equal(t, float64(0), *(out[2].Metric[0].Gauge.Value))
	//
	assert.Equal(t, "ns_subsys_mymeter_count", *(out[3].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[3].Type))
	assert.Equal(t, float64(1), *(out[3].Metric[0].Gauge.Value))
	//
	assert.Equal(t, "ns_subsys_mymeter_mean_rate", *(out[4].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[4].Type))
	assert.NotZero(t, *(out[4].Metric[0].Gauge.Value))
}

func TestExporter_ExportOnce_Timer(t *testing.T) {
	// Given an exporter with a registered meter,
	metrics.GetOrRegisterTimer("mytimer", sourceRegistry).Update(1 * time.Second)
	// Then exporting to Prometheus,
	exporter.ExportOnce()
	// Should deliver 14 gauges!
	out, err := targetRegistry.Gather()
	assert.NoError(t, err)
	assert.Len(t, out, 14)

	assert.Equal(t, "ns_subsys_mytimer_15m_rate", *(out[0].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	assert.Equal(t, float64(0), *(out[0].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_1m_rate", *(out[1].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[1].Type))
	assert.Equal(t, float64(0), *(out[1].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_5m_rate", *(out[2].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[2].Type))
	assert.Equal(t, float64(0), *(out[2].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_75p", *(out[3].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[3].Type))
	assert.Equal(t, float64(1*time.Second), *(out[3].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_95p", *(out[4].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[4].Type))
	assert.Equal(t, float64(1*time.Second), *(out[4].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_99_9p", *(out[5].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[5].Type))
	assert.Equal(t, float64(1*time.Second), *(out[5].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_99p", *(out[6].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[6].Type))
	assert.Equal(t, float64(1*time.Second), *(out[6].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_count", *(out[7].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[7].Type))
	assert.Equal(t, float64(1), *(out[7].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_max", *(out[8].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[8].Type))
	assert.Equal(t, float64(1*time.Second), *(out[8].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_mean", *(out[9].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[9].Type))
	assert.Equal(t, float64(1*time.Second), *(out[9].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_mean_rate", *(out[10].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[10].Type))
	assert.NotZero(t, *(out[10].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_median", *(out[11].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[11].Type))
	assert.Equal(t, float64(1*time.Second), *(out[11].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_min", *(out[12].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[12].Type))
	assert.Equal(t, float64(1*time.Second), *(out[12].Metric[0].Gauge.Value))

	assert.Equal(t, "ns_subsys_mytimer_stddev", *(out[13].Name))
	assert.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[13].Type))
	assert.Zero(t, *(out[13].Metric[0].Gauge.Value))
}

func TestPrometheusMetricName(t *testing.T) {
	metric := "metric"
	// Given a set of scenarios of values of types as returned by Go Metrics,
	scenarios := map[string]string{
		"95%":     "metric_95p",
		"99.9%":   "metric_99_9p",
		"1m.rate": "metric_1m_rate",
	}
	// Then expect PrometheusMetricKey to return the final clean metric name.
	for in, out := range scenarios {
		assert.Equal(t, out, PrometheusMetricName(metric, in))
	}
}

func TestReplaceUnsafeKeyCharacters(t *testing.T) {
	// Given a set of scenarios of unsafe to safe key translations,
	scenarios := map[string]string{
		"":        "unnamed",
		"_":       "_",
		"foo":     "foo",
		"foo_bar": "foo_bar", // Retain '_'
		"foo.bar": "foo_bar",
		"foo/bar": "foo_bar",
		"99%":     "99p", // Percentiles
	}
	// Then expect ReplaceUnsafeKeyCharacters to return the safe key.
	for in, out := range scenarios {
		assert.Equal(t, out, ReplaceUnsafeKeyCharacters(in))
	}
}

func TestAsFloat64(t *testing.T) {
	// Given a set of scenarios of values of types as returned by Go Metrics,
	scenarios := map[interface{}]float64{
		"":          0,
		"foo":       1,
		int64(12):   float64(12),
		float64(13): float64(13),
	}
	// Then expect AsFloat64 to return the converted value we can hand out as a Prometheus Gauge.
	for in, out := range scenarios {
		assert.Equal(t, out, AsFloat64(in))
	}
}
