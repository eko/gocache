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
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
