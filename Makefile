.PHONY: mocks

mocks:
	# mocks
	mockery -case=snake -name=CacheInterface -dir=cache/ -output test/mocks/cache/
	mockery -case=snake -name=CodecInterface -dir=codec/ -output test/mocks/codec/
	mockery -case=snake -name=SetterCacheInterface -dir=cache/ -output test/mocks/cache/
	mockery -case=snake -name=MetricsInterface -dir=metrics/ -output test/mocks/metrics/
	mockery -case=snake -name=StoreInterface -dir=store/ -output test/mocks/store/

	# in package store clients mocks
	mockery -case=snake -inpkg -name=BigcacheClientInterface -dir=store/ -output store/
	mockery -case=snake -inpkg -name=MemcacheClientInterface -dir=store/ -output store/
	mockery -case=snake -inpkg -name=RedisClientInterface -dir=store/ -output store/
	mockery -case=snake -inpkg -name=RistrettoClientInterface -dir=store/ -output store/
