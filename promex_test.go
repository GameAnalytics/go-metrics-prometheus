package promex

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/require"
)

const (
	namespace     = "ns"
	subsystem     = "subsys"
	flushInterval = 1 * time.Second
)

func newRegistriesAndExporter() (metrics.Registry, *prometheus.Registry, *Exporter) {
	src := metrics.NewRegistry()
	target := prometheus.NewRegistry()
	ex := NewExporter(namespace, subsystem, src, target, flushInterval)
	return src, target, ex
}

func TestExporter_ExportOnce_Counter(t *testing.T) {
	src, target, ex := newRegistriesAndExporter()
	// Given an exporter with a registered counter of value 1,
	metrics.GetOrRegisterCounter("mycounter", src).Inc(1)
	// Then exporting to Prometheus,
	ex.ExportOnce()
	// Should deliver 1 gauge of value 1.
	out, err := target.Gather()
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "ns_subsys_mycounter_count", *(out[0].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	require.Equal(t, float64(1), *(out[0].Metric[0].Gauge.Value))

}

func TestExporter_ExportOnce_Gauge(t *testing.T) {
	src, target, ex := newRegistriesAndExporter()
	// Given an exporter with a registered gauge of value 2,
	metrics.GetOrRegisterGauge("mygauge", src).Update(2)
	// Then exporting to Prometheus,
	ex.ExportOnce()
	// Should deliver 1 gauge of value 2.
	out, err := target.Gather()
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "ns_subsys_mygauge_value", *(out[0].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	require.Equal(t, float64(2), *(out[0].Metric[0].Gauge.Value))
}

func TestExporter_ExportOnce_Meter(t *testing.T) {
	src, target, ex := newRegistriesAndExporter()
	// Given an exporter with a registered meter,
	metrics.GetOrRegisterMeter("mymeter", src).Mark(1)
	// Then exporting to Prometheus,
	ex.ExportOnce()
	// Should deliver 5 gauges,
	out, err := target.Gather()
	require.NoError(t, err)
	require.Len(t, out, 5)
	//
	require.Equal(t, "ns_subsys_mymeter_15m_rate", *(out[0].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	require.Equal(t, float64(0), *(out[0].Metric[0].Gauge.Value))
	//
	require.Equal(t, "ns_subsys_mymeter_1m_rate", *(out[1].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[1].Type))
	require.Equal(t, float64(0), *(out[1].Metric[0].Gauge.Value))
	//
	require.Equal(t, "ns_subsys_mymeter_5m_rate", *(out[2].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[2].Type))
	require.Equal(t, float64(0), *(out[2].Metric[0].Gauge.Value))
	//
	require.Equal(t, "ns_subsys_mymeter_count", *(out[3].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[3].Type))
	require.Equal(t, float64(1), *(out[3].Metric[0].Gauge.Value))
	//
	require.Equal(t, "ns_subsys_mymeter_mean_rate", *(out[4].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[4].Type))
	require.NotZero(t, *(out[4].Metric[0].Gauge.Value))
}

func TestExporter_ExportOnce_Timer(t *testing.T) {
	src, target, ex := newRegistriesAndExporter()
	// Given an exporter with a registered meter,
	metrics.GetOrRegisterTimer("mytimer", src).Update(1 * time.Second)
	// Then exporting to Prometheus,
	ex.ExportOnce()
	// Should deliver 14 gauges!
	out, err := target.Gather()
	require.NoError(t, err)
	require.Len(t, out, 14)

	require.Equal(t, "ns_subsys_mytimer_15m_rate", *(out[0].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[0].Type))
	require.Equal(t, float64(0), *(out[0].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_1m_rate", *(out[1].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[1].Type))
	require.Equal(t, float64(0), *(out[1].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_5m_rate", *(out[2].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[2].Type))
	require.Equal(t, float64(0), *(out[2].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_75p", *(out[3].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[3].Type))
	require.Equal(t, float64(1*time.Second), *(out[3].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_95p", *(out[4].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[4].Type))
	require.Equal(t, float64(1*time.Second), *(out[4].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_99_9p", *(out[5].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[5].Type))
	require.Equal(t, float64(1*time.Second), *(out[5].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_99p", *(out[6].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[6].Type))
	require.Equal(t, float64(1*time.Second), *(out[6].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_count", *(out[7].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[7].Type))
	require.Equal(t, float64(1), *(out[7].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_max", *(out[8].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[8].Type))
	require.Equal(t, float64(1*time.Second), *(out[8].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_mean", *(out[9].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[9].Type))
	require.Equal(t, float64(1*time.Second), *(out[9].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_mean_rate", *(out[10].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[10].Type))
	require.NotZero(t, *(out[10].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_median", *(out[11].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[11].Type))
	require.Equal(t, float64(1*time.Second), *(out[11].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_min", *(out[12].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[12].Type))
	require.Equal(t, float64(1*time.Second), *(out[12].Metric[0].Gauge.Value))

	require.Equal(t, "ns_subsys_mytimer_stddev", *(out[13].Name))
	require.Equal(t, io_prometheus_client.MetricType_GAUGE, *(out[13].Type))
	require.Zero(t, *(out[13].Metric[0].Gauge.Value))
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
		require.Equal(t, out, PrometheusMetricName(metric, in))
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
		require.Equal(t, out, ReplaceUnsafeKeyCharacters(in))
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
		require.Equal(t, out, AsFloat64(in))
	}
}
