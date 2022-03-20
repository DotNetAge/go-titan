package registry

import "go-titan/runtime"

// Registry 服务注册器
type Registry interface {

	// Register 向注册中心注册服务器
	Register(*runtime.ServiceDesc) error

	// Unregister 注销已注册的服务
	Unregister() error

	// GetServices 获取已注册的服务实例的列表
	GetServices() ([]*runtime.ServiceDesc, error)
}
