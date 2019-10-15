[![TravisBuildStatus](https://api.travis-ci.org/eko/gocache.svg?branch=master)](https://travis-ci.org/eko/gocache)
[![GoDoc](https://godoc.org/github.com/eko/gocache?status.png)](https://godoc.org/github.com/eko/gocache)
[![GoReportCard](https://goreportcard.com/badge/github.com/eko/gocache)](https://goreportcard.com/report/github.com/eko/gocache)

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

## Built-in stores

* [Memory (bigcache)](https://github.com/allegro/bigcache) (allegro/bigcache)
* [Memory (ristretto)](https://github.com/dgraph-io/ristretto) (dgraph-io/ristretto)
* [Memcache](https://github.com/bradfitz/gomemcache) (bradfitz/memcache)
* [Redis](https://github.com/go-redis/redis/v7) (go-redis/redis)
* More to come soon

## Built-in metrics providers

* [Prometheus](https://github.com/prometheus/client_golang)

## Available cache features in detail

### A simple cache

Here is a simple cache instanciation with Redis but you can also look at other available stores:

#### Memcache

```go
memcacheStore := store.NewMemcache(
	memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212"),
	&store.Options{
		Expiration: 10*time.Second,
	},
)

cacheManager := cache.New(memcacheStore)
err := cacheManager.Set("my-key", []byte("my-value), &cache.Options{
	Expiration: 15*time.Second, // Override default value of 10 seconds defined in the store
})
if err != nil {
    panic(err)
}

value := cacheManager.Get("my-key")
```

#### Memory (using Bigcache)

```go
bigcacheClient, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
bigcacheStore := store.NewBigcache(bigcacheClient, nil) // No otions provided (as second argument)

cacheManager := cache.New(bigcacheStore)
err := cacheManager.Set("my-key", "my-value", nil)
if err != nil {
    panic(err)
}

value := cacheManager.Get("my-key")
```

#### Memory (using Ristretto)

```go
ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
	NumCounters: 1000,
	MaxCost: 100,
	BufferItems: 64,
})
if err != nil {
    panic(err)
}
ristrettoStore := store.NewRistretto(ristrettoCache, nil)

cacheManager := cache.New(ristrettoStore)
err := cacheManager.Set("my-key", "my-value", &cache.Options{Cost: 2})
if err != nil {
    panic(err)
}

value := cacheManager.Get("my-key")
```

#### Redis

```go
redisStore := store.NewRedis(redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
}), nil)

cacheManager := cache.New(redisStore)
err := cacheManager.Set("my-key", "my-value", &cache.Options{Expiration: 15*time.Second})
if err != nil {
    panic(err)
}

value := cacheManager.Get("my-key")
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
ristrettoStore := store.NewRistretto(ristrettoCache, nil)
redisStore := store.NewRedis(redisClient, &cache.Options{Expiration: 5*time.Second})

// Initialize chained cache
cacheManager := cache.NewChain(
    cache.New(ristrettoStore),
    cache.New(redisStore),
)

// ... Then, do what you want with your cache
```

`Chain` cache also put data back in previous caches when it's found so in this case, if ristretto doesn't have the data in its cache but redis have, data will also get setted back into ristretto (memory) cache.

### A loadable cache

This cache will provide a load function that acts as a callable function and will set your data back in your cache in case they are not available:

```go
// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := store.NewRedis(redisClient, nil)

// Initialize a load function that loads your data from a custom source
loadFunction := func(key interface{}) (interface{}, error) {
    // ... retrieve value from available source
    return &Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}, nil
}

// Initialize loadable cache
cacheManager := cache.NewLoadable(
	loadFunction,
	cache.New(redisStore),
)

// ... Then, you can get your data and your function will automatically put them in cache(s)
```

Of course, you can also pass a `Chain` cache into the `Loadable` one so if your data is not available in all caches, it will bring it back in all caches.

### A metric cache to retrieve cache statistics

This cache will record metrics depending on the metric provider you pass to it. Here we give a Prometheus provider:

```go
// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := store.NewRedis(redisClient, nil)

// Initializes Prometheus metrics service
promMetrics := metrics.NewPrometheus("my-test-app")

// Initialize metric cache
cacheManager := cache.NewMetric(
	promMetrics,
	cache.New(redisStore),
)

// ... Then, you can get your data and metrics will be observed by Prometheus
```

### A marshaler wrapper

Some caches like Redis stores and returns the value as a string so you have to marshal/unmarshal your structs if you want to cache an object. That's why we bring a marshaler service that wraps your cache and make the work for you:

```go
// Initialize Redis client and store
redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
redisStore := store.NewRedis(redisClient, nil)

// Initialize chained cache
cacheManager := cache.NewMetric(
	promMetrics,
	cache.New(redisStore),
)

// Initializes marshaler
marshaller := marshaler.New(cacheManager)

key := BookQuery{Slug: "my-test-amazing-book"}
value := Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}

err = marshaller.Set(key, value)
if err != nil {
    panic(err)
}

returnedValue, err := marshaller.Get(key, new(Book))
if err != nil {
    panic(err)
}

// Then, do what you want with the  value
```

The only thing you have to do is to specify the struct in which you want your value to be unmarshalled as a second argument when calling the `.Get()` method.

### All together!

Finally, you can mix all of these available caches or bring them together to build the cache you want to.
Here is a full example of how it can looks like:

```go
package main

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/marshaler"
	"github.com/eko/gocache/metrics"
	"github.com/eko/gocache/store"
	"github.com/go-redis/redis/v7"
)

// Book is a test struct that represents a single book
type Book struct {
	ID   int
	Name string
	Slug string
}

func main() {
	// Initialize Prometheus metrics collector
	promMetrics := metrics.NewPrometheus("my-test-app")

	// Initialize Ristretto store
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost: 100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	ristrettoStore := store.NewRistretto(ristrettoCache, &cache.Options{Cost: 4})

	// Initialize Redis store
	redisStore := store.NewRedis(redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"}), &cache.Options{Expiration: 5*time.Second})

	// Initialize a load function that loads your data from a custom source
	loadFunction := func(key interface{}) (interface{}, error) {
		// ... retrieve value from available source
		return &Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}, nil
	}

	// Initialize a chained cache (memory with Ristretto then Redis) with Prometheus metrics
	// and a load function that will put data back into caches if none has the value
	cacheManager := cache.NewMetric(promMetrics, cache.NewLoadable(loadFunction,
		cache.NewChain(
			cache.New(ristrettoStore),
			cache.New(redisStore),
		),
	))

	marshaller := marshaler.New(cacheManager)

	key := Book{Slug: "my-test-amazing-book"}
	value := Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}

	err = marshaller.Set(key, value, nil)
	if err != nil {
		panic(err)
	}

	returnedValue, err := marshaller.Get(key, new(Book))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", returnedValue)
}
```

## Community

Please feel free to contribute on this library and do not hesitate to open an issue if you want to discuss about a feature.

## Run tests

Test suite can be run with:

```bash
$ go test -v ./...
```
