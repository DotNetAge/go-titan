package service

import (
	"fmt"

	auth "github.com/dotnetage/go-titan/auth"
	"github.com/dotnetage/go-titan/config"
	"github.com/dotnetage/go-titan/registry"
	"github.com/dotnetage/go-titan/runtime"

	uuid "github.com/satori/go.uuid"

	"github.com/golang/glog"
	"go.uber.org/zap"
)

type Option func(*Options)

// Options 微服务运行期设置选项
type Options struct {
	ServiceDesc      *runtime.ServiceDesc
	EnableReflection bool                 // 是否启用反射特性
	HealthCheck      bool                 // 是否启用健康度检查
	Auth             auth.Tokens          // 身份验证组件
	Registry         registry.Registry    // 注册中心
	Config           *config.ServerConfig // 配置中心
	Logger           *zap.Logger          // 日志
}

func newOptions(opts ...Option) *Options {
	logger, err := zap.NewDevelopment()
	if err != nil {
		glog.Fatalf("无法创建日志: %v", err)
	}

	options := &Options{
		ServiceDesc: &runtime.ServiceDesc{
			Name:     "",
			Version:  "latest",
			EndPoint: *config.NewEndpoint("127.0.0.1:8080"),
		},
		EnableReflection: true,
		Registry:         nil,
		Logger:           logger,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
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

func HealthCheck(enabled bool) Option {
	return func(options *Options) {
		options.HealthCheck = enabled
	}
}

func Reflection(enabled bool) Option {
	return func(o *Options) {
		o.EnableReflection = enabled
	}
}

func Auth(a auth.Tokens) Option {
	return func(o *Options) {
		o.Auth = a
	}
}

func Registry(registry registry.Registry) Option {
	return func(o *Options) {
		o.Registry = registry
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
		// o.ServiceDesc.ID = fmt.Sprintf("%s-%s", name, o.ServiceDesc.Addr)
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

func WithConfig(conf *config.ServerConfig) Option {
	return func(o *Options) {
		o.Config = conf
		o.ServiceDesc.EndPoint = *conf.EndPoint
		if len(conf.Name) > 0 {
			o.ServiceDesc.Name = conf.Name
		}
	}
}
