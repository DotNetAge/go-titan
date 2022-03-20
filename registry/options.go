package registry

import "go.uber.org/zap"

type Options struct {
	Endpoints   []string
	DialTimeout int
	Logger      *zap.Logger
}

type Option func(*Options)

func WithEndPoints(endpoints ...string) Option {
	return func(o *Options) {
		o.Endpoints = endpoints
	}
}

func DialTimeout(timeout int) Option {
	return func(o *Options) {
		o.DialTimeout = timeout
	}
}

func DefaultOptions() *Options {
	return &Options{
		Endpoints:   make([]string, 0),
		DialTimeout: 3,
		Logger:      zap.NewNop(),
	}
}
