package gateway

import (
	"fmt"
	"go-titan/config"
	"go-titan/registry"
	"go-titan/runtime"

	"github.com/golang/glog"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type Option func(*Options)

// Options 微服务运行期设置选项
type Options struct {
	ServiceDesc *runtime.ServiceDesc
	Transports  []*config.EndPoint
	Registry    registry.Registry // 注册中心
	Logger      *zap.Logger       // 日志
	CORS        *config.CORSConfig
}

func newOptions(opts ...Option) *Options {
	logger, err := zap.NewDevelopment()
	if err != nil {
		glog.Fatalf("无法创建日志: %v", err)
	}

	options := &Options{
		ServiceDesc: &runtime.ServiceDesc{
			Name:     "api-gateway",
			Version:  "latest",
			EndPoint: *config.NewEndpoint("127.0.0.1:8081"),
		},
		Registry:   nil,
		Logger:     logger,
		Transports: make([]*config.EndPoint, 0),
		CORS:       &config.CORSConfig{},
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

func WithConfig(conf *config.GatewayConfig) Option {
	return func(options *Options) {
		if len(conf.Name) > 0 {
			options.ServiceDesc.Name = conf.Name
		}
		options.ServiceDesc.EndPoint = *conf.EndPoint
		options.CORS = conf.CORS
		options.Transports = conf.Transports
	}
}

func Trans(endpoints ...*config.EndPoint) Option {
	return func(o *Options) {
		o.Transports = append(o.Transports, endpoints...)
	}
}

func Registry(registry registry.Registry) Option {
	return func(o *Options) {
		o.Registry = registry
	}
}

func AllowOrigins(origins ...string) Option {
	return func(options *Options) {
		options.CORS.Origins = origins
	}
}

func AllowMethods(methods ...string) Option {
	return func(options *Options) {
		options.CORS.Methods = methods
	}
}

func AllowHeaders(headers ...string) Option {
	return func(options *Options) {
		options.CORS.Headers = headers
	}
}

func Logger(logger *zap.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// Name 设置微服务的唯一服务名称
func Name(name string) Option {
	return func(o *Options) {
		o.ServiceDesc.Name = name
		o.ServiceDesc.ID = fmt.Sprintf("%s-%s", name, uuid.NewV4().String())
	}
}

// Version 设置微服务的运行版本 默认为 "latest"
func Version(version string) Option {
	return func(o *Options) {
		o.ServiceDesc.Version = version
	}
}

func TLS(certFile, keyFile string) Option {
	return func(o *Options) {
		o.ServiceDesc.TLS = true
		o.ServiceDesc.CertFile = certFile
		o.ServiceDesc.KeyFile = keyFile
	}
}

func TTL(ttl int) Option {
	return func(o *Options) {
		o.ServiceDesc.TTL = ttl
	}
}

func Listen(addr string) Option {
	return func(o *Options) {
		o.ServiceDesc.Addr = addr
	}
}
