package metrics

import (
	"github.com/eko/gocache/codec"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespaceCache = "cache"
)

var (
	cacheCollector *prometheus.GaugeVec = initCacheCollector(namespaceCache)
)

// Prometheus represents the prometheus struct for collecting metrics
type Prometheus struct {
	name      string
	collector *prometheus.GaugeVec
}

func initCacheCollector(namespace string) *prometheus.GaugeVec {
	c := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "collector",
			Namespace: namespace,
			Help:      "This represent the number of items in cache",
		},
		[]string{"service", "store", "metric"},
	)
	prometheus.MustRegister(c)
	return c
}

// NewPrometheus initializes a new prometheus metric instance
func NewPrometheus(service string) *Prometheus {
	return &Prometheus{service, cacheCollector}
}

// Record records a metric in prometheus by specyfing the store name, metric name and value
func (m *Prometheus) Record(store, metric string, value float64) {
	m.collector.WithLabelValues(m.name, store, metric).Set(value)
}

// RecordFromCodec records metrics in prometheus by retrieving values from a codec instance
func (m *Prometheus) RecordFromCodec(codec codec.CodecInterface) {
	stats := codec.GetStats()
	storeType := codec.GetStore().GetType()

	m.Record(storeType, "hit_count", float64(stats.Hits))
	m.Record(storeType, "miss_count", float64(stats.Miss))

	m.Record(storeType, "set_success", float64(stats.SetSuccess))
	m.Record(storeType, "set_error", float64(stats.SetError))

	m.Record(storeType, "delete_success", float64(stats.DeleteSuccess))
	m.Record(storeType, "delete_error", float64(stats.DeleteError))

	m.Record(storeType, "invalidate_success", float64(stats.InvalidateSuccess))
	m.Record(storeType, "invalidate_error", float64(stats.InvalidateError))
}
