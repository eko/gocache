.PHONY: update-stores-version mocks test benchmark-store

# Usage: VERSION=v4.1.3 make update-stores-version
update-stores-version:
	ls store/ | xargs -I % bash -c "sed -i '' -E 's,github.com/eko/gocache/lib/v4 v[0-9]\.[0-9]\.[0-9],github.com/eko/gocache/lib/v4 ${VERSION},g' store/%/go.mod"

mocks:
	mockgen -source=lib/cache/interface.go -destination=lib/cache/cache_mock.go -package=cache
	mockgen -source=lib/codec/interface.go -destination=lib/codec/codec_mock.go -package=codec
	mockgen -source=lib/metrics/interface.go -destination=lib/metrics/metrics_mock.go -package=metrics
	mockgen -source=lib/store/interface.go -destination=lib/store/store_mock.go -package=store
	mockgen -source=store/bigcache/bigcache.go -destination=store/bigcache/bigcache_mock.go -package=bigcache
	mockgen -source=store/memcache/memcache.go -destination=store/memcache/memcache_mock.go -package=memcache
	mockgen -source=store/redis/redis.go -destination=store/redis/redis_mock.go -package=redis
	mockgen -source=store/rediscluster/rediscluster.go -destination=store/rediscluster/rediscluster_mock.go -package=rediscluster
	mockgen -source=store/ristretto/ristretto.go -destination=store/ristretto/ristretto_mock.go -package=ristretto
	mockgen -source=store/freecache/freecache.go -destination=store/freecache/freecache_mock.go -package=freecache
	mockgen -source=store/go_cache/go_cache.go -destination=store/go_cache/go_cache_mock.go -package=go_cache
	mockgen -source=store/hazelcast/hazelcast.go -destination=store/hazelcast/hazelcast_mock.go -package=hazelcast

test:
	cd lib; GOGC=10 go test -v -p=4 ./...
	cd store/bigcache; GOGC=10 go test -v -p=4 ./...
	cd store/freecache; GOGC=10 go test -v -p=4 ./...
	cd store/go_cache; GOGC=10 go test -v -p=4 ./...
	cd store/memcache; GOGC=10 go test -v -p=4 ./...
	cd store/pegasus; GOGC=10 go test -v -p=4 ./...
	cd store/redis; GOGC=10 go test -v -p=4 ./...
	cd store/rediscluster; GOGC=10 go test -v -p=4 ./...
	cd store/ristretto; GOGC=10 go test -v -p=4 ./...
