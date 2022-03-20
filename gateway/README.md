# 设计的初衷

由于gRPC入门的学习路径很陡峭，不容易入手。而且Go世界中并没有找到什么很趁手的纯grpc开发框架，所以就萌生了简化开发纯gRPC微服务的工具集的想法。

## 入门

### gRPC服务端

1. 定义Protobuf
2. 构建Protobuf文件
3. 实现 grpc service
4. 启用 MicroService

在Titan中只要构建 `MicroService` 就可以完成大量复杂的gRPC起动操作：

#### 注册gRPC服务

```go
package main

import (
	"context"
	"titan/auth/api/v1/authpb"
	"go-titan/boot"
	"go-titan/config"
	"titan/auth/auth"
)

func main() {
	factory := auth.NewAuthServiceFileFactory()
	authSvc := auth.NewAuthService(factory, "./cert.key")

	boot.NewMicroService(). 
		Register(&authpb.Auth_ServiceDesc, authSvc). 
		Start("127.0.0.1:8081") // 起动时设置服务地址
}
```
需要使用`Register`方法对每个gRPC生成类中的`XXXDesc`描述类及接口实现进行注册。

最后调用`Start`方法进行起动即可。

#### 启用服务注册

Titan 内置支持etcd作为服务注册中心，直接设置etcd的服务地址就可以完成接入。

调用`SetRegistry`方法可以同时设置多个etcd的服务地址，代码如下所示：

```go
boot.NewMicroService(). 
    Register(&authpb.Auth_ServiceDesc, authSvc).
    SetRegistry("127.0.0.1:5672","10.0.0.2:5672").
    Start("127.0.0.1:8081") 
```

#### 链接跟踪

Titan 默认支持Zipkin服务进行链路跟踪，首先需要起动zipkin服务，然后通过`OpenTracing`方法指定Zipkin的服务地址与端口就可以起用链接跟踪。

```go
boot.NewMicroService(). 
    Register(&authpb.Auth_ServiceDesc, authSvc).
    SetRegistry("127.0.0.1:5672","10.0.0.2:5672").
	OpenTracing("127.0.0.1:9411").
    Start("127.0.0.1:8081") 
```

#### 兼容WebAPI的gRPC服务

为了可以方便地让您的微服务可以支持Http Restful API, `MicroService`内置了API网关，如果

```go
boot.NewMicroService(). 
    Register(&authpb.Auth_ServiceDesc, authSvc).
    SetRegistry("127.0.0.1:5672","10.0.0.2:5672").
	OpenTracing("127.0.0.1:9411").
    WebAPI("127.0.0.1:8082").
    Transport(authpb.RegisterAuthHandlerFromEndpoint).
	Start("127.0.0.1:8081") 
```





```go
package main

import (
	"context"
	"titan/auth/api/v1/authpb"
	"titan/catalog/api/v1/catalogpb"
	"go-titan/boot"
)

func main() {
	boot.NewMicroService("service-name").
		Register(pb.RegisterXXXServer, &YourServiceImpl{}).
		// Config(cfg). // 直接从对象中获取设置，则不需要执行以下的方法
		SetRegistry("10.0.0.1:5672"). // 设置服务发现注册中心
		OpenTracing("127.0.0.1:9411"). // 启用OpenTracing		SetOption(serverOption). // 设置gRPC服务器的起动选项
		ServeTLS(certFile, keyFile). // 启用安全访问
		WebAPI("127.0.0.1:8081"). // 启用WebAPI兼容接口
		CORS(corsCfn...). // 开启跨域设置
		Route("/custom_path", customHandler). // 附加路由
		//Config(&cfg).                         // 直接设置配置
		OnAuth(func(ctx context.Context) { // 服务验证Token
		}).
		Start("127.0.0.1:8080")
}
```

#### 安全认证事件

`OnAuth`

### 网关

```go
package main

import (
	"context"
	"titan/auth/api/v1/authpb"
	"titan/catalog/api/v1/catalogpb"
	"go-titan/boot"
)

func main() {
	boot.NewGateway(). // 代理gRPC服务(可选)
		Transport("auth-server", authpb.RegisterAuthHandlerFromEndpoint).
		Transport("internal-svr", catalogpb.RegisterCatalogHandlerFromEndpoint).
		Config(gatewayConfig).
		SetRegistry("10.0.0.1:5672"). // 设置注册中心
		SetOption(dialogOption). // 设置拨号选项
		OpenTracing(true). // 启用链路跟踪
		LoadBalancing(true). // 启用负载均衡
		ServeTLS("cert.rsa", "key.pub"). // 启用安全通信 SSH
		CORS(corsConfig). // 启用跨域访问
		Route("/auth", oAuthHandler). // 增加自定义路由
		OnAuth(func(ctx context.Context) { // 前置验证Token
		}).
		Start()
}
```

### 客户端连接器

```go
package main

import "go-titan/boot"

func main() {

	conn := boot.NewServiceConnector("auth-server", []string{"10.0.0.1:5672"}).
		SetOption(dialogOption). // 设置拨号选项 
		OpenTracing(true). // 启用链路跟踪
		LoadBalancing(true). // 启用负载均衡
		ServeTLS("cert.key"). // 启用安全通信 SSH		
		Connect()

	client := pb.CreateAuthClient(conn)

}
```