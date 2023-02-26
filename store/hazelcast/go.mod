module github.com/eko/gocache/store/hazelcast/v4

go 1.19

require (
	github.com/eko/gocache/lib/v4 v4.1.2
	github.com/golang/mock v1.6.0
	github.com/hazelcast/hazelcast-go-client v1.3.2
	github.com/stretchr/testify v1.8.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shirou/gopsutil/v3 v3.21.5 // indirect
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/tklauser/numcpus v0.2.1 // indirect
	golang.org/x/exp v0.0.0-20221126150942-6ab00d035af9 // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/eko/gocache/lib/v4 => ../../lib/
