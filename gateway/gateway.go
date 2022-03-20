package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go-titan/config"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc/credentials"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"go.uber.org/zap"
)

type (
	Middleware func(http.Handler) http.Handler

	// Gateway 网关定义
	Gateway interface {
		// Options 返回网关配置
		Options() *Options
		// Use 添加中间件实例
		Use(middlewares ...Middleware) Gateway
		// Handle 添加Http处理器
		Handle(method, pattern string, handler runtime.HandlerFunc) Gateway
		// Transport 向网关注册处理器方法
		Transport(serverName string, registerFunc ...ClientRegisterFunc) Gateway
		// Start 起动网关
		Start()
	}

	routFunc           func(r *runtime.ServeMux) error
	ClientRegisterFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) (err error)

	defaultGateway struct {
		options         *Options
		httpServer      *http.Server
		logger          *zap.Logger
		clientRegisters map[string][]ClientRegisterFunc
		handlers        []routFunc
		middlewares     []Middleware
	}
)

func New(opts ...Option) Gateway {

	b := &defaultGateway{
		options:         newOptions(opts...),
		handlers:        make([]routFunc, 0),
		clientRegisters: make(map[string][]ClientRegisterFunc),
		middlewares:     make([]Middleware, 0),
	}

	b.logger = b.options.Logger
	return b
}

func (g *defaultGateway) Options() *Options {
	return g.options
}

func (b *defaultGateway) Handle(method, pattern string, handler runtime.HandlerFunc) Gateway {
	b.handlers = append(b.handlers, func(r *runtime.ServeMux) error {
		return r.HandlePath(method, pattern, handler)
	})
	return b
}

func (b *defaultGateway) Use(middlewares ...Middleware) Gateway {
	for _, middleware := range middlewares {
		b.middlewares = append(b.middlewares, middleware)
	}
	return b
}

func (a *defaultGateway) Transport(serverName string, registerFunc ...ClientRegisterFunc) Gateway {
	a.clientRegisters[serverName] = registerFunc
	return a
}

func (b *defaultGateway) Start() {

	// 创建一个可取消的上下文(如：请求发到一半可随时取消)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 批量注册客户拨号连接
	gwmux := runtime.NewServeMux(defaultMarshalerOption())

	if len(b.clientRegisters) > 0 && len(b.options.Transports) == 0 {
		b.logger.Fatal("没有指定gRPC服务地址")
		return
	}

	for serverName, registerFunc := range b.clientRegisters {
		endpoint := b.options.Transports[0]
		if endpoint.Name != serverName {
			for _, ep := range b.options.Transports {
				if ep.Name == serverName {
					endpoint = ep
					break
				}
			}
		}

		b.logger.Sugar().Infof("正在连接服务 %v", serverName)
		for _, regFnc := range registerFunc {
			// 这里不能传入 Cancel 的上下文，否则就会持续连接失败
			err := regFnc(context.Background(),
				gwmux,
				endpoint.Addr,
				buildDialOptions(endpoint, b.logger))

			if err != nil {
				b.logger.Sugar().Fatalf("连接服务失败 %v", err)
			}
		}
	}

	// 附加的路由
	if len(b.handlers) > 0 {
		for _, h := range b.handlers {
			if e := h(gwmux); e != nil {
				b.logger.Fatal(e.Error())
			}
		}
	}

	defaultHandler := func() http.Handler { return gwmux }()

	// 附加中间件
	if len(b.middlewares) > 0 {
		for i := range b.middlewares {
			m := b.middlewares[len(b.middlewares)-i-1]
			defaultHandler = m(defaultHandler)
		}
	}

	b.httpServer = &http.Server{
		Addr:    b.options.ServiceDesc.EndPoint.Addr,
		Handler: b.options.CORS.Allows(defaultHandler),
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		defer signal.Stop(stop)
		<-stop
		cancel()
		b.logger.Info("尝试关闭网关服务...")

		if err := b.httpServer.Shutdown(context.Background()); err != nil {
			b.logger.Sugar().Fatalf("关闭网关服务器失败:%v", err)
		}
		//
		//if b.resolver != nil {
		//	b.resolver.Close()
		//}
	}()

	b.logger.Sugar().Infof("正在起动网关，运行于%s", b.httpServer.Addr)

	if b.options.ServiceDesc.EndPoint.TLS {
		// 启用https网关
		b.logger.Info("启用TLS的安全连接")
		if err := b.httpServer.ListenAndServeTLS(b.options.ServiceDesc.EndPoint.CertFile,
			b.options.ServiceDesc.EndPoint.KeyFile); err != http.ErrServerClosed {
			b.logger.Sugar().Fatalf("无法起动网关 %v", err)
			// panic(err)
		}
	} else {
		b.logger.Info("启用不安全连接")
		if err := b.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			b.logger.Sugar().Fatalf("无法起动网关 %v", err)
		}
	}

	b.logger.Sugar().Info("网关服务已下线")

}

func defaultMarshalerOption() runtime.ServeMuxOption {
	return runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseEnumNumbers: true, // 枚举字段的值使用数字
				UseProtoNames:  true,
				// 传给 clients 的 json key 使用下划线 `_`
				// AccessToken string `protobuf:"bytes,1,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`
				// 这里说明应使用 access_token
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true, // 忽略 client 发送的不存在的 poroto 字段
			},
		},
	)
}

func buildDialOptions(endpoint *config.EndPoint, logger *zap.Logger) []grpc.DialOption {
	grpc_zap.ReplaceGrpcLoggerV2(logger)
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_zap.UnaryClientInterceptor(logger),
			//grpc_retry.UnaryClientInterceptor(), // 服务重试
			//grpc_opentracing.UnaryClientInterceptor(), // 链路跟踪
			//grpc_prometheus.UnaryClientInterceptor,
		)),
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(
				grpc_zap.StreamClientInterceptor(logger),
				//grpc_retry.StreamClientInterceptor(),
			)),
		// grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`), // 已经支持负载均衡
	}

	if endpoint.TLS {
		// creds, _ := credentials.NewClientTLSFromFile(endpoint.CertFile, "")
		creds, err := loadTLSCredentials(endpoint.CAFile,
			endpoint.CertFile,
			endpoint.KeyFile)
		if err != nil {
			logger.Fatal(err.Error())
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return opts
}

func loadTLSCredentials(caFile, clientCertFile, clientKeyFile string) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	c := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(c), nil
}

//// OpenTracing 启用链路跟踪
//func (a *ApiGateway) OpenTracing(addr string) *ApiGateway {
//	if addr != "" {
//		a.config.OpenTracing.Addr = addr
//	}
//
//	if a.config.EndPoint.Addr != "" {
//
//		opt, err := NewZipkinClientOption(a.name,
//			a.config.EndPoint.Addr,
//			a.logger)
//
//		if err == nil {
//			a.dialOptions = append(a.dialOptions, opt)
//		}
//	}
//	return a
//}
//
