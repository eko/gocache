module github.com/eko/gocache/v4/store/rueidis

go 1.19

require (
	github.com/eko/gocache/v4/lib v0.0.0
	github.com/golang/mock v1.6.0
	github.com/rueian/rueidis v0.0.86
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20221126150942-6ab00d035af9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/eko/gocache/v4/lib => ../../lib/
