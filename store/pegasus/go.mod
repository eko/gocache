module github.com/eko/gocache/store/pegasus/v4

go 1.25

require (
	github.com/XiaoMi/pegasus-go-client v0.0.0-20220519103347-ba0e68465cd5
	github.com/eko/gocache/lib/v4 v4.1.6
	github.com/smartystreets/assertions v1.13.0
	github.com/smartystreets/goconvey v1.7.2
	github.com/spf13/cast v1.5.0
)

require (
	github.com/cenkalti/backoff/v4 v4.1.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/pegasus-kv/thrift v0.13.0 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	k8s.io/apimachinery v0.16.13 // indirect
)

replace github.com/eko/gocache/lib/v4 => ../../lib/
