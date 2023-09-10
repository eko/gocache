module github.com/eko/gocache/store/pegasus/v4

go 1.19

require (
	github.com/XiaoMi/pegasus-go-client v0.0.0-20220519103347-ba0e68465cd5
	github.com/eko/gocache/lib/v4 v4.1.5
	github.com/smartystreets/assertions v1.13.0
	github.com/smartystreets/goconvey v1.7.2
	github.com/spf13/cast v1.5.0
)

require (
	github.com/cenkalti/backoff/v4 v4.1.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/pegasus-kv/thrift v0.13.0 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	golang.org/x/exp v0.0.0-20221126150942-6ab00d035af9 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637 // indirect
	k8s.io/apimachinery v0.0.0-20191123233150-4c4803ed55e3 // indirect
)

replace github.com/eko/gocache/lib/v4 => ../../lib/
