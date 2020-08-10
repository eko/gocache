.PHONY: mocks

mocks:
	mockgen -source=interface.go -destination=test/mocks/cache/cache_interface.go -package=mocks
	mockgen -source=store_interface.go -destination=test/mocks/store/store_interface.go -package=mocks
	mockgen -source=store/redis.go -destination=test/mocks/store/clients/redis_interface.go -package=mocks
