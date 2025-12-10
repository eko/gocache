module github.com/eko/gocache/store/rueidis/v4

go 1.25

require (
	github.com/eko/gocache/lib/v4 v4.1.6
	github.com/redis/rueidis v1.0.37
	github.com/redis/rueidis/mock v1.0.37
	github.com/redis/rueidis/rueidiscompat v1.0.37
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.6.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	golang.org/x/sys v0.39.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/eko/gocache/lib/v4 => ../../lib/
