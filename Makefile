.PHONY: mocks

mocks:
	mockgen -source=cache/interface.go -destination=test/mocks/cache/cache_interface.go -package=mocks
	mockgen -source=codec/interface.go -destination=test/mocks/codec/codec_interface.go -package=mocks
	mockgen -source=metrics/interface.go -destination=test/mocks/metrics/metrics_interface.go -package=mocks
	mockgen -source=store/interface.go -destination=test/mocks/store/store_interface.go -package=mocks
	mockgen -source=store/bigcache.go -destination=test/mocks/store/clients/bigcache_interface.go -package=mocks
	mockgen -source=store/memcache.go -destination=test/mocks/store/clients/memcache_interface.go -package=mocks
	mockgen -source=store/redis.go -destination=test/mocks/store/clients/redis_interface.go -package=mocks
	mockgen -source=store/ristretto.go -destination=test/mocks/store/clients/ristretto_interface.go -package=mocks
	mockgen -source=store/freecache.go -destination=test/mocks/store/clients/freecache_interface.go -package=mocks
