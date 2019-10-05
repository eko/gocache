Gache
=====


```go
// Initialize stores
ristrettoCache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 1000, MaxCost: 100, BufferItems: 64})
if err != nil {
    panic(err)
}

ristrettoStore := store.NewRistrettoStore(ristrettoCache)
redisStore := store.NewRedisStore(redis.NewClient(&redis.Options{Addr: host}))

// Initialize Prometheus metrics collector
promMetrics := metrics.NewPrometheusMetrics("my-test-app")

// Initialize a load function that loads your data from a custom source
loadFunction := func(key interface{}) (interface{}, error) {
    // ... retrieve value from available source
    return &Book{ID: 1, Title: "My test amazing book", Slug: "my-test-amazing-book"}, nil
}

// Initialize a chained cache (memory with Ristretto then Redis) with Prometheus metrics
// and a load function that will put data back into caches if none has the value
cache := cache.NewMetricCache(promMetrics, cache.NewLoadableCache(
    loadFunction,
    cache.NewChainCache(
        cache.NewCache(ristrettoStore, 5 * time.Second),
        cache.NewCache(redisStore, 15 * time.Second),
    ),
))

key := Book{Slyg: "my-test-amazing-book"}
value := Book{ID: 1, Title: "My test amazing book", Slug: "my-test-amazing-book"}

cache.Set(key, value)
returnedValue := cache.Get(key) // Returns a type interface{}, cast it back to Book struct

if cachedValue, ok := returnedValue.(*Book); ok {
    fmt.Printf("%v", cachedValue)
}
```
