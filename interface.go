package gocache

import (
	"github.com/yeqown/gocache/types"
)

// ICache represents the interface for all caches (aggregates, metric, memory, redis, ...)
type ICache interface {
	// Get 从cache中获取缓存值
	Get(key string) ([]byte, error)

	// Set 新增或者更新一个缓存键
	Set(key string, object interface{}, options *types.StoreOptions) error

	// Delete delete一个缓存key
	Delete(key string) error

	// 根据选项过期一些 key
	Invalidate(opt types.InvalidateOptions) error

	// GetType 获取当前cache的类型
	GetType() string
}
