package gateway

import (
	"go-titan/registry/etcd"

	"go.uber.org/zap"
)

type ServiceConnector struct {
	serviceName   string
	registryAddrs []string
	logger        *zap.Logger
}

func NewServiceConnector(serviceName string, registryAddrs ...string) *ServiceConnector {
	l, _ := zap.NewProduction()
	return &ServiceConnector{
		serviceName:   serviceName,
		registryAddrs: registryAddrs,
		logger:        l,
	}
}

func (s *ServiceConnector) resolveAddr() string {
	r := etcd.NewResolver(s.registryAddrs, s.logger)
	return r.Scheme()
}

//func (s *ServiceConnector) Connect() (*grpc.ClientConn, error) {
//	openTracingOpt, err := NewZipkinClientOption("zipkin-httpServer", "127.0.0.1:9411", s.logger)
//	if err != nil {
//		s.logger.Fatal("unable to create tracer: %+v\n", zap.Error(err))
//	}
//	opts := defaultDialOptions(s.logger)
//	opts = append(opts, openTracingOpt)
//	return grpc.Dial(s.resolveAddr(), opts...)
//}
