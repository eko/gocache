module github.com/eko/gocache/example/metrics

go 1.25.0

require (
	github.com/dgraph-io/ristretto/v2 v2.3.0
	github.com/eko/gocache/lib/v4 v4.2.0
	github.com/eko/gocache/store/ristretto/v4 v4.2.2
)

replace (
	github.com/eko/gocache/lib/v4 => ../../lib/
	github.com/eko/gocache/store/ristretto/v4 => ../../store/ristretto/
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.52.3 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	go.uber.org/mock v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
