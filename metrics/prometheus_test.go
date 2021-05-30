package metrics

import (
	"testing"
	"time"

	"github.com/eko/gocache/v2/codec"
	mocksCodec "github.com/eko/gocache/v2/test/mocks/codec"
	mocksStore "github.com/eko/gocache/v2/test/mocks/store"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewPrometheus(t *testing.T) {
	// Given
	serviceName := "my-test-service-name"

	// When
	metrics := NewPrometheus(serviceName)

	// Then
	assert.IsType(t, new(Prometheus), metrics)

	assert.Equal(t, serviceName, metrics.service)
	assert.IsType(t, new(prometheus.GaugeVec), metrics.collector)
}

func TestRecord(t *testing.T) {
	// Given
	metrics := NewPrometheus("my-test-service-name")

	// When
	metrics.record("redis", "hit_count", 6)

	// Then
	metric, err := metrics.collector.GetMetricWithLabelValues("my-test-service-name", "redis", "hit_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v := testutil.ToFloat64(metric)

	assert.Equal(t, float64(6), v)
}

func TestRecordFromCodec(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	redisStore := mocksStore.NewMockStoreInterface(ctrl)
	redisStore.EXPECT().GetType().Return("redis")

	stats := &codec.Stats{
		Hits:              4,
		Miss:              6,
		SetSuccess:        12,
		SetError:          3,
		DeleteSuccess:     8,
		DeleteError:       5,
		InvalidateSuccess: 2,
		InvalidateError:   1,
	}

	testCodec := mocksCodec.NewMockCodecInterface(ctrl)
	testCodec.EXPECT().GetStats().Return(stats)
	testCodec.EXPECT().GetStore().Return(redisStore)

	metrics := NewPrometheus("my-test-service-name")

	// When
	metrics.RecordFromCodec(testCodec)

	// Wait for data to be processed
	for len(metrics.codecChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

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
		{
			metricName: "delete_success",
			expected:   float64(stats.DeleteSuccess),
		},
		{
			metricName: "delete_error",
			expected:   float64(stats.DeleteError),
		},
		{
			metricName: "invalidate_success",
			expected:   float64(stats.InvalidateSuccess),
		},
		{
			metricName: "invalidate_error",
			expected:   float64(stats.InvalidateError),
		},
	}

	for _, tc := range testCases {
		metric, err := metrics.collector.GetMetricWithLabelValues("my-test-service-name", "redis", tc.metricName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v := testutil.ToFloat64(metric)

		assert.Equal(t, tc.expected, v)
	}
}
