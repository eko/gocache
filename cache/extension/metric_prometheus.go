package extension

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/yeqown/gocache/cache"
)

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

type prometheusMetric struct {
	serviceName string
	collector   *prometheus.GaugeVec
	stats       *cache.Stats
}

func newPrometheusMetric() IMetrics {
	return &prometheusMetric{
		serviceName: "default-cache",             // TODO: 支持自定义
		stats:       new(cache.Stats),            //
		collector:   initCacheCollector("cache"), // TODO：支持自定义
	}
}

func assemble(op Op, result OpResult) string {
	return string(op) + "_" + string(result)
}

// 记录某一个操作是否成功
// FIXME: 并发安全
func (p prometheusMetric) Record(op Op, result OpResult) {
	metric := assemble(op, result)
	switch op {
	case Get:
		if result == Success {
			p.stats.Hits++
		} else {
			p.stats.Miss++
		}
	case Set:
		if result == Success {
			p.stats.SetSuccess++
		} else {
			p.stats.SetError++
		}
	case Delete:
		if result == Success {
			p.stats.DeleteSuccess++
		} else {
			p.stats.DeleteError++
		}
	case Clear:
		if result == Success {
			p.stats.ClearSuccess++
		} else {
			p.stats.ClearError++
		}
	case Invalidate:
		if result == Success {
			p.stats.InvalidateSuccess++
		} else {
			p.stats.InvalidateError++
		}
	default:
		// do nothing
	}

	p.record1(metric, 1)
}

func (p prometheusMetric) record1(metric string, value int) {
	p.collector.WithLabelValues(p.serviceName, metric).Add(float64(value))

	vec, _ := p.collector.GetMetricWithLabelValues()
	vec.Desc()
}

// 获取统计状态
func (p prometheusMetric) GetStats() *cache.Stats {
	return p.stats
}
