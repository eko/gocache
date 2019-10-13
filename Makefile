.PHONY: mocks

mocks:
	mockery -case=snake -name=CacheInterface -dir=cache/ -output test/mocks/cache/
	mockery -case=snake -name=CodecInterface -dir=codec/ -output test/mocks/codec/
	mockery -case=snake -name=SetterCacheInterface -dir=cache/ -output test/mocks/cache/
	mockery -case=snake -name=MetricsInterface -dir=metrics/ -output test/mocks/metrics/
	mockery -case=snake -name=StoreInterface -dir=store/ -output test/mocks/store/
	mockery -case=snake -name=MemcacheClientInterface -dir=store/ -output test/mocks/store/
	mockery -case=snake -name=RedisClientInterface -dir=store/ -output test/mocks/store/
	mockery -case=snake -name=RistrettoClientInterface -dir=store/ -output test/mocks/store/
