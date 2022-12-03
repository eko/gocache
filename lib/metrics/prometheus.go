package metrics

import (
	"github.com/eko/gocache/lib/v4/codec"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespaceCache = "cache"
)

var cacheCollector *prometheus.GaugeVec = initCacheCollector(namespaceCache)

// Prometheus represents the prometheus struct for collecting metrics
type Prometheus struct {
	service      string
	collector    *prometheus.GaugeVec
	codecChannel chan codec.CodecInterface
}

func initCacheCollector(namespace string) *prometheus.GaugeVec {
	c := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "collector",
			Namespace: namespace,
			Help:      "This represent the number of items in cache",
		},
		[]string{"service", "store", "metric"},
	)
	return c
}

// NewPrometheus initializes a new prometheus metric instance
func NewPrometheus(service string) *Prometheus {
	prometheus := &Prometheus{
		service:      service,
		collector:    cacheCollector,
		codecChannel: make(chan codec.CodecInterface, 10000),
	}

	go prometheus.recorder()

	return prometheus
}

// Record records a metric in prometheus by specifying the store name, metric name and value
func (m *Prometheus) record(store, metric string, value float64) {
	m.collector.WithLabelValues(m.service, store, metric).Set(value)
}

// Recorder records metrics in prometheus by retrieving values from the codec channel
func (m *Prometheus) recorder() {
	for codec := range m.codecChannel {
		stats := codec.GetStats()
		storeType := codec.GetStore().GetType()

		m.record(storeType, "hit_count", float64(stats.Hits))
		m.record(storeType, "miss_count", float64(stats.Miss))

		m.record(storeType, "set_success", float64(stats.SetSuccess))
		m.record(storeType, "set_error", float64(stats.SetError))

		m.record(storeType, "delete_success", float64(stats.DeleteSuccess))
		m.record(storeType, "delete_error", float64(stats.DeleteError))

		m.record(storeType, "invalidate_success", float64(stats.InvalidateSuccess))
		m.record(storeType, "invalidate_error", float64(stats.InvalidateError))
	}
}

// RecordFromCodec sends the given codec into the codec channel to be read from recorder
func (m *Prometheus) RecordFromCodec(codec codec.CodecInterface) {
	m.codecChannel <- codec
}
