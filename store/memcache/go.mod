module github.com/eko/gocache/store/memcache/v4

go 1.22

require (
	github.com/bradfitz/gomemcache v0.0.0-20230124162541-5f7a7d875746
	github.com/eko/gocache/lib/v4 v4.1.6
	github.com/stretchr/testify v1.8.1
	go.uber.org/mock v0.4.0
	golang.org/x/sync v0.7.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/eko/gocache/lib/v4 => ../../lib/
