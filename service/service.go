package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dotnetage/go-titan/auth"

	health "google.golang.org/grpc/health/grpc_health_v1"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type (
	MicroService interface {
		// Register 将GRPC服务实例关联至到微服务容器
		Register(registerDesc *grpc.ServiceDesc, srv interface{}) MicroService

		// Options 获取微服务容器内的全部选项
		Options() *Options

		// Instances 获取已注册的GRPC服务实例
		Instances() []interface{}

		// Server 获取内置的gRPC服务器实例
		Server() *grpc.Server

		// Start 启动微服务
		Start()
	}

	microService struct {
		logger              *zap.Logger
		options             *Options
		server              *grpc.Server
		rpcServiceDescs     []*grpc.ServiceDesc
		rpcServiceInstances []interface{}
	}
)

func New(options ...Option) MicroService {
	b := &microService{
		options:             newOptions(options...),
		rpcServiceDescs:     make([]*grpc.ServiceDesc, 0),
		rpcServiceInstances: make([]interface{}, 0),
	}
	b.logger = b.options.Logger
	return b
}

func (b *microService) Instances() []interface{} {
	return b.rpcServiceInstances
}

func (b *microService) Server() *grpc.Server {
	return b.server
}

func (b *microService) WaitForClose() {

	// 优雅地关机
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for { // 用一个死循环阻断主进程
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			b.logger.Sugar().Infof("正在尝试关闭%s服务...", b.options.ServiceDesc.Name)
			b.server.GracefulStop()

			if b.options.Registry != nil {
				if err := b.options.Registry.Unregister(); err != nil {
					b.logger.Fatal("注销服务失败", zap.Error(err))
				}
			}
			time.Sleep(time.Second)
			b.logger.Info("验证服务已下线")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func (b *microService) Register(registerDesc *grpc.ServiceDesc, srv interface{}) MicroService {
	b.rpcServiceDescs = append(b.rpcServiceDescs, registerDesc)
	b.rpcServiceInstances = append(b.rpcServiceInstances, srv)

	if b.options.ServiceDesc.Name == "" {
		Name(registerDesc.ServiceName)(b.options)
	}

	return b
}

func (b *microService) Start() {
	b.initGRPCServer()

	lis, err := b.options.ServiceDesc.Listen()

	if err != nil {
		b.logger.Fatal(fmt.Sprintf("起动网络侦听失败:%v", err))
	}

	if lis == nil {
		b.logger.Fatal(fmt.Sprintf("端口%v初始化失败，请检查是否被点用或地址是否有效", b.options.ServiceDesc.Addr))
	}

	go func() {
		services := b.server.GetServiceInfo()
		for k, _ := range services {
			b.logger.Info(fmt.Sprintf("启用%v服务", k))
		}
		b.logger.Info(fmt.Sprintf("验证服务已成功上线: %s", b.options.ServiceDesc.EndPoint.Addr))
		err = b.server.Serve(lis)
		if err != nil {
			panic(fmt.Sprintf("服务启动失败 : %v", err))
		}
	}()

	if b.options.Registry != nil {
		b.logger.Info(fmt.Sprintf("正在向注册中心注册验证服务: %s", b.options.ServiceDesc.Name))
		if err = b.options.Registry.Register(b.options.ServiceDesc); err != nil {
			b.logger.Fatal("服务注册组件起动失败", zap.Error(err))
		}
	}

	b.WaitForClose()
}

func (b *microService) Options() *Options {
	return b.options
}

func (b *microService) initGRPCServer() {

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			//grpc_opentracing.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(b.options.Logger),
			grpc_recovery.UnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(b.onAuth),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(b.options.Logger),
			grpc_recovery.StreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(b.onAuth),
		)),
	}

	grpc_zap.ReplaceGrpcLoggerV2(b.options.Logger)

	if b.options.ServiceDesc.EndPoint.TLS {
		creds, err := credentials.NewServerTLSFromFile(b.options.ServiceDesc.EndPoint.CertFile,
			b.options.ServiceDesc.EndPoint.KeyFile)

		if err != nil {
			b.logger.Fatal("加载安全证书时出错")
			panic(err)
		}
		serverOptions = append(serverOptions, grpc.Creds(creds))
	}

	b.server = grpc.NewServer(serverOptions...)
	if b.options.EnableReflection {
		reflection.Register(b.server)
	}

	if len(b.rpcServiceDescs) == 0 {
		b.options.Logger.Fatal("没有任何可运行的服务，请使用Use方法先进行服务注册")
	}

	if b.options.HealthCheck {
		health.RegisterHealthServer(b.server, NewHealthServer())
	}

	// 注册服务器
	for i, desc := range b.rpcServiceDescs {
		b.server.RegisterService(desc, b.rpcServiceInstances[i])
	}
}

func (b *microService) onAuth(ctx context.Context) (context.Context, error) {
	if b.options.Auth != nil {
		accessToken, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if len(accessToken) > 0 && err == nil {
			// 验证Token是否正确
			user, err := b.options.Auth.Inspect(accessToken)
			if err != nil {
				b.logger.Error("用户访问令牌无效", zap.Error(err))
				return ctx, nil
			}
			return auth.ContextWithUser(ctx, user), nil
		}
	}
	return ctx, nil
}

//
//func loadTLSCredentials(certFile, keyFile string) (credentials.TransportCredentials, error) {
//	// Load server's certificate and private key
//	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
//	if err != nil {
//		return nil, err
//	}
//
//	// Create the credentials and return it
//	config := &tls.Config{
//		Certificates: []tls.Certificate{serverCert},
//		ClientAuth:   tls.NoClientCert,
//	}
//
//	return credentials.NewTLS(config), nil
//}
