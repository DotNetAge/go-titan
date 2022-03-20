# `Service` 服务组件

服务对象是titan微服务的入口，她可以:

- 注册GRPC服务，作为单个或多个服务的运行容器
- 自动完成服务注册，支持 etcd、naco和 consol
- 支持本地日志与共享日志
- 支持JWT,可自动读取用户令牌中的信息并写入用户上下文内通过`IsAuth`与`CurrentUser`获取
- 可支持TLS安全连接
- 可支持Unary与Stream模式GRPC服务
- 可支持服务方法反射（能让evan等工具进行直接调用）
