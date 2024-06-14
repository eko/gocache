[![Test](https://github.com/eko/gocache/actions/workflows/all.yml/badge.svg?branch=master)](https://github.com/eko/gocache/actions/workflows/all.yml)
[![GoDoc](https://godoc.org/github.com/eko/gocache?status.png)](https://godoc.org/github.com/eko/gocache)
[![GoReportCard](https://goreportcard.com/badge/github.com/eko/gocache)](https://goreportcard.com/report/github.com/eko/gocache)
[![codecov](https://codecov.io/gh/eko/gocache/branch/master/graph/badge.svg)](https://codecov.io/gh/eko/gocache)

Gocache
=======

Guess what is Gocache? a Go cache library.
This is an extendable cache library that brings you a lot of features for caching data.

## Overview

Here is what it brings in detail:

* ✅ Multiple cache stores: actually in memory, redis, or your own custom store
* ✅ A chain cache: use multiple cache with a priority order (memory then fallback to a redis shared cache for instance)
* ✅ A loadable cache: allow you to call a callback function to put your data back in cache
* ✅ A metric cache to let you store metrics about your caches usage (hits, miss, set success, set error, ...)
* ✅ A marshaler to automatically marshal/unmarshal your cache values as a struct
* ✅ Define default values in stores and override them when setting data
* ✅ Cache invalidation by expiration time and/or using tags
* ✅ Use of Generics

## Built-in stores

* [Memory (bigcache)](https://github.com/allegro/bigcache) (allegro/bigcache)
* [Memory (ristretto)](https://github.com/dgraph-io/ristretto) (dgraph-io/ristretto)
* [Memory (go-cache)](https://github.com/patrickmn/go-cache) (patrickmn/go-cache)
* [Memcache](https://github.com/bradfitz/gomemcache) (bradfitz/memcache)
* [Redis](https://github.com/go-redis/redis) (go-redis/redis)
* [Redis (rueidis)](https://github.com/redis/rueidis) (redis/rueidis)
* [Freecache](https://github.com/coocood/freecache) (coocood/freecache)
* [Pegasus](https://pegasus.apache.org/) ([apache/incubator-pegasus](https://github.com/apache/incubator-pegasus)) [benchmark](https://pegasus.apache.org/overview/benchmark/)
* [Hazelcast](https://github.com/hazelcast/hazelcast-go-client) (hazelcast-go-client/hazelcast)
* More to come soon

## Built-in metrics providers

* [Prometheus](https://github.com/prometheus/client_golang)

## Installation

To begin working with the latest version of gocache, you can import the library in your project:

```go
go get github.com/eko/gocache/lib/v4
```

and then, import the store(s) you want to use between all available ones:

```go
go get github.com/eko/gocache/store/bigcache/v4
go get github.com/eko/gocache/store/freecache/v4
go get github.com/eko/gocache/store/go_cache/v4
go get github.com/eko/gocache/store/hazelcast/v4
go get github.com/eko/gocache/store/memcache/v4
go get github.com/eko/gocache/store/pegasus/v4
go get github.com/eko/gocache/store/redis/v4
go get github.com/eko/gocache/store/rediscluster/v4
go get github.com/eko/gocache/store/rueidis/v4
go get github.com/eko/gocache/store/ristretto/v4
```

Then, simply use the following import statements:

```go
import (
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/store/redis/v4"
)
```

If you run into any errors, please be sure to run `go mod tidy` to clean your go.mod file.

## Available cache features in detail

### A simple cache

Here is a simple cache instantiation with Redis but you can also look at other available stores:

#### Memcache

```go
memcacheStore := memcache_store.NewMemcache(
	memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212"),
	store.WithExpiration(10*time.Second),
)

cacheManager := cache.New[[]byte](memcacheStore)
err := cacheManager.Set(ctx, "my-key", []byte("my-value"),
	store.WithExpiration(15*time.Second), // Override default value of 10 seconds defined in the store
)
if err != nil {
    panic(err)
}

value := cacheManager.Get(ctx, "my-key")

cacheManager.Delete(ctx, "my-key")

cacheManager.Clear(ctx) // Clears the entire cache, in case you want to flush all cache
```

#### Memory (using Bigcache)

```go
bigcacheClient, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
bigcacheStore := bigcache_store.NewBigcache(bigcacheClient)

cacheManager := cache.New[[]byte](bigcacheStore)
err := cacheManager.Set(ctx, "my-key", []byte("my-value"))
if err != nil {
    panic(err)
}

value := cacheManager.Get(ctx, "my-key")
```

#### Memory (using Ristretto)

```go
import (
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
)
ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
	NumCounters: 1000,
	MaxCost: 100,
	BufferItems: 64,
})
if err != nil {
    panic(err)
}
ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)

cacheManager := cache.New[string](ristrettoStore)
err := cacheManager.Set(ctx, "my-key", "my-value", store.WithCost(2))
if err != nil {
    panic(err)
}

value := cacheManager.Get(ctx, "my-key")

cacheManager.Delete(ctx, "my-key")
```

#### Memory (using Go-cache)

```go
gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
gocacheStore := gocache_store.NewGoCache(gocacheClient)

cacheManager := cache.New[[]byte](gocacheStore)
err := cacheManager.Set(ctx, "my-key", []byte("my-value"))
if err != nil {
	panic(err)
}

value, err := cacheManager.Get(ctx, "my-key")
if err != nil {
	panic(err)
}
fmt.Printf("%s", value)
```

#### Redis

```go
redisStore := redis_store.NewRedis(redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
}))

cacheManager := cache.New[string](redisStore)
err := cacheManager.Set(ctx, "my-key", "my-value", store.WithExpiration(15*time.Second))
if err != nil {
    panic(err)
}

value, err := cacheManager.Get(ctx, "my-key")
switch err {
	case nil:
		fmt.Printf("Get the key '%s' from the redis cache. Result: %s", "my-key", value)
	case redis.Nil:
		fmt.Printf("Failed to find the key '%s' from the redis cache.", "my-key")
	default:
	    fmt.Printf("Failed to get the value from the redis cache with key '%s': %v", "my-key", err)
}
```

#### [Redis Client-Side Caching](https://redis.io/docs/manual/client-side-caching/) (using rueidis)

```go
client, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
if err != nil {
    panic(err)
}

cacheManager := cache.New[string](rueidis_store.NewRueidis(
    client,
    store.WithExpiration(15*time.Second),
    store.WithClientSideCaching(15*time.Second)),
)

if err = cacheManager.Set(ctx, "my-key", "my-value"); err != nil {
    panic(err)
}

value, err := cacheManager.Get(ctx, "my-key")
if err != nil {
    log.Fatalf("Failed to get the value from the redis cache with key '%s': %v", "my-key", err)
}
log.Printf("Get the key '%s' from the redis cache. Result: %s", "my-key", value)
```

#### Freecache

```go
freecacheStore := freecache_store.NewFreecache(freecache.NewCache(1000), store.WithExpiration(10 * time.Second))

cacheManager := cache.New[[]byte](freecacheStore)
err := cacheManager.Set(ctx, "by-key", []byte("my-value"), opts)
if err != nil {
    panic(err)
}

value := cacheManager.Get(ctx, "my-key")
```

#### Pegasus

```go
pegasusStore, err := pegasus_store.NewPegasus(&store.OptionsPegasus{
    MetaServers: []string{"127.0.0.1:34601", "127.0.0.1:34602", "127.0.0.1:34603"},
})

if err != nil {
    fmt.Println(err)
    return
}

cacheManager := cache.New[string](pegasusStore)
err = cacheManager.Set(ctx, "my-key", "my-value", store.WithExpiration(10 * time.Second))
if err != nil {
    panic(err)
}

value, _ := cacheManager.Get(ctx, "my-key")
```

#### Hazelcast

```go
hzClient, err := hazelcast.StartNewClient(ctx)
if err != nil {
    log.Fatalf("Failed to start client: %v", err)
}

hzMap, err := hzClient.GetMap(ctx, "gocache")
if err != nil {
    b.Fatalf("Failed to get map: %v", err)
}

hazelcastStore := hazelcast_store.NewHazelcast(hzMap)

cacheManager := cache.New[string](hazelcastStore)
err := cacheManager.Set(ctx, "my-key", "my-value", store.WithExpiration(15*time.Second))
if err != nil {
    panic(err)
}

value, err := cacheManager.Get(ctx, "my-key")
if err != nil {
    panic(err)
}
fmt.Printf("Get the key '%s' from the hazelcast cache. Result: %s", "my-key", value)
```

### A chained cache

Here, we will chain caches in the following order: first in memory with Ristretto store, then in Redis (as a fallback):

```go
// Initialize Ristretto cache and Redis client
ristrettoCache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 1000, MaxCost: 100, BufferItems: 64})
if err != nil {
    panic(err)
}

redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

// Initialize stores
ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)
redisStore := redis_store.NewRedis(redisClient, store.WithExpiration(5*time.Second))

// Initialize chained cache
cacheManager := cache.NewChain[any](
    cache.New[any](ristrettoStore),
    cache.New[any](redisStore),
)

// ... Then, do what you want with your cache
```

`Chain` cache also put data back in previous caches when it's found so in this case, if ristretto doesn't have the data in its cache but redis have, data will also get setted back into ristretto (memory) cache.

### A loadable cache

This cache will provide a load function that acts as a callable function and will set your data back in your cache in case they are not available:

```go
type Book struct {
	ID string
	Name string
}

// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := redis_store.NewRedis(redisClient)

// Initialize a load function that loads your data from a custom source
loadFunction := func(ctx context.Context, key any) (*Book, error) {
    // ... retrieve value from available source
    return &Book{ID: 1, Name: "My test amazing book"}, nil
}

// Initialize loadable cache
cacheManager := cache.NewLoadable[*Book](
	loadFunction,
	cache.New[*Book](redisStore),
)

// ... Then, you can get your data and your function will automatically put them in cache(s)
```

Of course, you can also pass a `Chain` cache into the `Loadable` one so if your data is not available in all caches, it will bring it back in all caches.

### A metric cache to retrieve cache statistics

This cache will record metrics depending on the metric provider you pass to it. Here we give a Prometheus provider:

```go
// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := redis_store.NewRedis(redisClient)

// Initializes Prometheus metrics service
promMetrics := metrics.NewPrometheus("my-test-app")

// Initialize metric cache
cacheManager := cache.NewMetric[any](
	promMetrics,
	cache.New[any](redisStore),
)

// ... Then, you can get your data and metrics will be observed by Prometheus
```

### A marshaler wrapper

Some caches like Redis stores and returns the value as a string so you have to marshal/unmarshal your structs if you want to cache an object. That's why we bring a marshaler service that wraps your cache and make the work for you:

```go
// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := redis_store.NewRedis(redisClient)

// Initialize chained cache
cacheManager := cache.NewMetric[any](
	promMetrics,
	cache.New[any](redisStore),
)

// Initializes marshaler
marshal := marshaler.New(cacheManager)

key := BookQuery{Slug: "my-test-amazing-book"}
value := Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}

err = marshal.Set(ctx, key, value)
if err != nil {
    panic(err)
}

returnedValue, err := marshal.Get(ctx, key, new(Book))
if err != nil {
    panic(err)
}

// Then, do what you want with the  value

marshal.Delete(ctx, "my-key")
```

The only thing you have to do is to specify the struct in which you want your value to be un-marshalled as a second argument when calling the `.Get()` method.


### Cache invalidation using tags

You can attach some tags to items you create so you can easily invalidate some of them later.

Tags are stored using the same storage you choose for your cache.

Here is an example on how to use it:

```go
// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := redis_store.NewRedis(redisClient)

// Initialize chained cache
cacheManager := cache.NewMetric[*Book](
	promMetrics,
	cache.New[*Book](redisStore),
)

// Initializes marshaler
marshal := marshaler.New(cacheManager)

key := BookQuery{Slug: "my-test-amazing-book"}
value := &Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}

// Set an item in the cache and attach it a "book" tag
err = marshal.Set(ctx, key, value, store.WithTags([]string{"book"}))
if err != nil {
    panic(err)
}

// Remove all items that have the "book" tag
err := marshal.Invalidate(ctx, store.WithInvalidateTags([]string{"book"}))
if err != nil {
    panic(err)
}

returnedValue, err := marshal.Get(ctx, key, new(Book))
if err != nil {
	// Should be triggered because item has been deleted so it cannot be found.
    panic(err)
}
```

Mix this with expiration times on your caches to have a fine-tuned control on how your data are cached.

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisStore := redis_store.NewRedis(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	}), nil)

	cacheManager := cache.New[string](redisStore)
	err := cacheManager.Set(ctx, "my-key", "my-value", store.WithExpiration(15*time.Second))
	if err != nil {
		panic(err)
	}

	key := "my-key"
	value, err := cacheManager.Get(ctx, key)
	if err != nil {
		log.Fatalf("unable to get cache key '%s' from the cache: %v", key, err)
	}

	fmt.Printf("%#+v\n", value)
}

```

### Write your own custom cache

Cache respect the following interface so you can write your own (proprietary?) cache logic if needed by implementing the following interface:

```go
type CacheInterface[T any] interface {
	Get(ctx context.Context, key any) (T, error)
	Set(ctx context.Context, key any, object T, options ...store.Option) error
	Delete(ctx context.Context, key any) error
	Invalidate(ctx context.Context, options ...store.InvalidateOption) error
	Clear(ctx context.Context) error
	GetType() string
}
```

Or, in case you use a setter cache, also implement the `GetCodec()` method:

```go
type SetterCacheInterface[T any] interface {
	CacheInterface[T]
	GetWithTTL(ctx context.Context, key any) (T, time.Duration, error)

	GetCodec() codec.CodecInterface
}
```

As all caches available in this library implement `CacheInterface`, you will be able to mix your own caches with your own.

### Write your own custom store

You also have the ability to write your own custom store by implementing the following interface:

```go
type StoreInterface interface {
	Get(ctx context.Context, key any) (any, error)
	GetWithTTL(ctx context.Context, key any) (any, time.Duration, error)
	Set(ctx context.Context, key any, value any, options ...Option) error
	Delete(ctx context.Context, key any) error
	Invalidate(ctx context.Context, options ...InvalidateOption) error
	Clear(ctx context.Context) error
	GetType() string
}
```

Of course, I suggest you to have a look at current caches or stores to implement your own.

### Custom cache key generator

You can implement the following interface in order to generate a custom cache key:

```go
type CacheKeyGenerator interface {
	GetCacheKey() string
}
```

### Benchmarks

![Benchmarks](https://raw.githubusercontent.com/eko/gocache/master/lib/misc/benchmarks.jpeg)

## Run tests

To generate mocks using mockgen library, run:

```bash
$ make mocks
```

Test suite can be run with:

```bash
$ make test # run unit test
```

## Community

Please feel free to contribute on this library and do not hesitate to open an issue if you want to discuss about a feature.
