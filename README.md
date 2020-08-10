## gocache 

forked from https://github.com/eko/gocache and modified

### 特性

[1] store和cache 面向接口实现，可扩展（支持多种store）

[2] cache组件有 `PureCache` `ChainCache`, 且内置`命中率数据统计` `序列化和反序列化`扩展。 

### 快速上手

#### 1. 直接使用

```go
```

#### 使用扩展

```go

```

### 介绍

1. ICache 缓存组件接口

```go
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
```

2. IStore 存储组件接口
```go
// IStore is the interface for all available stores
type IStore interface {
	// Get .
	Get(key string) ([]byte, error)

	// Set .
	Set(key string, value interface{}, options *types.StoreOptions) error

	// Delete .
	Delete(key string) error

	// Invalidate .
	Invalidate(options types.InvalidateOptions) error

	// Clear .
	Clear() error

	// GetType .
	GetType() string
}
```

### TODO

[1] redisStore 优化 tags 实现; memcache 优化 tags 是否可以优化。

[2] 命中率等指标记录并输出，完善metrics