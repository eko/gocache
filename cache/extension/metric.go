package extension

import (
	"github.com/yeqown/gocache/cache"
	"github.com/yeqown/gocache/store"
)

type Op string

type OpResult string

const (
	Set        Op = "Set"
	Get        Op = "Get"
	Delete     Op = "Delete"
	Invalidate Op = "Invalidate"
	Clear      Op = "Clear"

	Success OpResult = "success"
	Failed  OpResult = "failed"
)

type IMetrics interface {
	GetStats() *cache.Stats
	Record(op Op, result OpResult)
}

// TODO: 记录到缓存或者其他组件中去
type metricWrapper struct {
	metric IMetrics
	c      cache.ICache
}

// 统计扩展的装饰器
func WrapWithMetrics(c cache.ICache) cache.ICache {
	return &metricWrapper{
		metric: newPrometheusMetric(),
		c:      c,
	}
}

func (m metricWrapper) Get(key interface{}, returnObj interface{}) (interface{}, error) {
	result, err := m.c.Get(key, returnObj)
	if err != nil {
		m.metric.Record(Get, Failed)
	}
	m.metric.Record(Get, Success)

	return result, err
}

func (m metricWrapper) Set(key, object interface{}, options *store.Options) error {
	err := m.c.Set(key, object, options)
	if err != nil {
		m.metric.Record(Set, Failed)
	}
	m.metric.Record(Set, Success)

	return err
}

func (m metricWrapper) Delete(key interface{}) error {
	err := m.c.Delete(key)
	if err != nil {
		m.metric.Record(Delete, Failed)
	}
	m.metric.Record(Delete, Success)

	return err
}

func (m metricWrapper) Invalidate(options store.InvalidateOptions) error {
	err := m.c.Invalidate(options)
	if err != nil {
		m.metric.Record(Invalidate, Failed)
	}
	m.metric.Record(Invalidate, Success)

	return err
}

func (m metricWrapper) Clear() error {
	err := m.c.Clear()
	if err != nil {
		m.metric.Record(Clear, Failed)
	}
	m.metric.Record(Clear, Success)

	return err
}

func (m metricWrapper) GetType() string {
	return m.c.GetType() + ".metric"
}

func (m metricWrapper) GetStats() *cache.Stats {
	return m.metric.GetStats()
}
