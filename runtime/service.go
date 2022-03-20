package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/dotnetage/go-titan/config"

	"google.golang.org/grpc/resolver"
)

// ServiceDesc 用于表示在注册中心注册的服务实例的信息
type ServiceDesc struct {
	config.EndPoint
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"` //服务版本
	Weight   int64             `json:"weight"`  //服务权重
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
}

func (s *ServiceDesc) Listen() (net.Listener, error) {
	return net.Listen(s.EndPoint.Network, s.EndPoint.Addr)
}

func BuildPrefix(info *ServiceDesc) string {
	if info.Version == "" {
		return fmt.Sprintf("/%s/", info.Name)
	}
	return fmt.Sprintf("/%s/%s/", info.Name, info.Version)
}

func BuildRegPath(info *ServiceDesc) string {
	return fmt.Sprintf("%s%s", BuildPrefix(info), info.Addr)
}

func ParseValue(value []byte) (ServiceDesc, error) {
	info := ServiceDesc{}
	if err := json.Unmarshal(value, &info); err != nil {
		return info, err
	}
	return info, nil
}

func SplitPath(path string) (ServiceDesc, error) {
	info := ServiceDesc{}
	strs := strings.Split(path, "/")
	if len(strs) == 0 {
		return info, errors.New("无效的路径")
	}
	info.Addr = strs[len(strs)-1]
	return info, nil
}

// Exist helper function
func Exist(l []resolver.Address, addr resolver.Address) bool {
	for i := range l {
		if l[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}

// Remove helper function
func Remove(s []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr.Addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

//
//func BuildResolverUrl(app string) string {
//	return etcd.schema + ":///" + app
//}
