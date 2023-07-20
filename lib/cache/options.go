package cache

import "time"

type Option func(*options)

type options struct {
	loadDefaultExpireTime time.Duration
}

func newOptions() *options {
	return &options{
		loadDefaultExpireTime: 0,
	}
}

func (o *options) WithLoadDefaultExpireTime(expireTime time.Duration) Option {
	return func(o *options) {
		o.loadDefaultExpireTime = expireTime
	}
}
