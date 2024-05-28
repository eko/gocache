module github.com/eko/gocache/store/rediscluster/v4

go 1.22

require (
	github.com/eko/gocache/lib/v4 v4.1.6
	github.com/redis/go-redis/v9 v9.0.2
	github.com/stretchr/testify v1.8.1
	go.uber.org/mock v0.4.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/eko/gocache/lib/v4 => ../../lib/
