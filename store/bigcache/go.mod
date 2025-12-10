module github.com/eko/gocache/store/bigcache/v4

go 1.25

require (
	github.com/allegro/bigcache/v3 v3.1.0
	github.com/eko/gocache/lib/v4 v4.2.2
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.6.0
)

replace github.com/eko/gocache/lib/v4 => ../../lib/

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
