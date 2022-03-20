package consul

import (
	"fmt"
	"go-titan/config"
	"go-titan/registry"
	"go-titan/runtime"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type ConsulRegistry struct {
	options *registry.Options
	srvInfo *runtime.ServiceDesc
	logger  *zap.Logger
	client  *api.Client
}

func NewConsulRegistry(opts ...registry.Option) registry.Registry {
	options := registry.DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	addr := "127.0.0.1:8500"
	if len(options.Endpoints) > 0 {
		addr = options.Endpoints[0]
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr

	client, err := api.NewClient(cfg)
	if err != nil {
		options.Logger.Fatal("创建Consul客户端错误", zap.Error(err))
		panic(err)
	}

	return &ConsulRegistry{
		options: options,
		client:  client,
		logger:  options.Logger,
	}

}

func (r *ConsulRegistry) Register(node *runtime.ServiceDesc) error {
	r.srvInfo = node
	agent := r.client.Agent()

	r.logger.Sugar().Infof("Check 服务器地址: %v", fmt.Sprintf("%v:%v/%v", node.GetHost(), node.GetPort(), node.Name))

	// TODO: 注册后总是会检查失败
	reg := &api.AgentServiceRegistration{
		ID:      node.ID,        // 服务节点的名称
		Name:    node.Name,      // 服务名称
		Port:    node.GetPort(), // 服务端口
		Address: node.LocalIP(), // 服务 IP
		Tags:    node.Tags,      // // tag，可以为空
		Check: &api.AgentServiceCheck{
			Interval: "5s",
			TCP:      fmt.Sprintf("%s:%d", node.LocalIP(), node.GetPort()),
			// GRPC:                           fmt.Sprintf("%s:%d/%s", node.GetHost(), node.GetPort(), node.Name),
			DeregisterCriticalServiceAfter: "5s", // 注销时间，相当于过期时间
		},
	}

	if err := agent.ServiceRegister(reg); err != nil {
		r.logger.Fatal("Consul 注册失败", zap.Error(err))
		return err
	}
	return nil
}

// 注销已注册的服务
func (r *ConsulRegistry) Unregister() error {
	agent := r.client.Agent()
	if err := agent.ServiceDeregister(r.srvInfo.ID); err != nil {
		r.logger.Fatal("Consul注销服务失败", zap.Error(err))
		return err
	}
	return nil
}

// 获取注册服务器的信息
func (r *ConsulRegistry) GetServices() ([]*runtime.ServiceDesc, error) {
	catalogService, _, _ := r.client.Catalog().Service(r.srvInfo.Name, "", nil)
	if len(catalogService) > 0 {
		result := make([]*runtime.ServiceDesc, len(catalogService))
		for index, server := range catalogService {
			s := &runtime.ServiceDesc{
				ID:       server.ServiceID,
				Name:     server.ServiceName,
				EndPoint: *config.NewEndpoint(fmt.Sprintf("%s:%v", server.Address, server.ServicePort)),
				// Metadata: sever.ServiceMeta,
			}
			result[index] = s
		}
		return result, nil
	}
	return nil, nil
}
