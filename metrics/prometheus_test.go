package metrics

import (
	"testing"

	"github.com/eko/gocache/codec"
	mocksCodec "github.com/eko/gocache/test/mocks/codec"
	mocksStore "github.com/eko/gocache/test/mocks/store"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestNewPrometheus(t *testing.T) {
	// Given
	serviceName := "my-test-service-name"

	// When
	metrics := NewPrometheus(serviceName)

	// Then
	assert.IsType(t, new(Prometheus), metrics)

	assert.Equal(t, serviceName, metrics.name)
	assert.IsType(t, new(prometheus.GaugeVec), metrics.collector)
}

func TestRecord(t *testing.T) {
	// Given
	metrics := NewPrometheus("my-test-service-name")

	// When
	metrics.Record("redis", "hit_count", 6)

	// Then
	dtoMetric := &dto.Metric{}
	metric, _ := metrics.collector.GetMetricWithLabelValues("my-test-service-name", "redis", "hit_count")
	metric.Write(dtoMetric)

	assert.Equal(t, float64(6), dtoMetric.GetGauge().GetValue())
}

func TestRecordFromCodec(t *testing.T) {
	// Given
	redisStore := &mocksStore.StoreInterface{}
	redisStore.On("GetType").Return("redis")

	stats := &codec.Stats{
		Hits:       4,
		Miss:       6,
		SetSuccess: 12,
		SetError:   3,
	}

	testCodec := &mocksCodec.CodecInterface{}
	testCodec.On("GetStats").Return(stats)
	testCodec.On("GetStore").Return(redisStore)

	metrics := NewPrometheus("my-test-service-name")

	// When
	metrics.RecordFromCodec(testCodec)

	// Then
	testCases := []struct {
		metricName string
		expected   float64
	}{
		{
			metricName: "hit_count",
			expected:   float64(stats.Hits),
		},
		{
			metricName: "miss_count",
			expected:   float64(stats.Miss),
		},
		{
			metricName: "set_success",
			expected:   float64(stats.SetSuccess),
		},
		{
			metricName: "set_error",
			expected:   float64(stats.SetError),
		},
	}

	for _, tc := range testCases {
		dtoMetric := &dto.Metric{}
		metric, _ := metrics.collector.GetMetricWithLabelValues("my-test-service-name", "redis", tc.metricName)
		metric.Write(dtoMetric)

		assert.Equal(t, tc.expected, dtoMetric.GetGauge().GetValue())
	}
}
