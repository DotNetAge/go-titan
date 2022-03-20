package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type EndPoint struct {
	Name     string `mapstructure:"name"`
	Addr     string `mapstructure:"addr"`
	TTL      int    `mapstructure:"ttl"`
	Network  string `mapstructure:"network"`
	TLS      bool   `mapstructure:"tls"`  // 是否使用证TLS证书
	KeyFile  string `mapstructure:"key"`  // 私钥文件
	CertFile string `mapstructure:"cert"` // 公钥文件
	CAFile   string `mapstructure:"ca"`   // Ca文件
}

func NewEndpoint(addr string) *EndPoint {
	ep := DefaultEndPoint()
	ep.Addr = addr
	return ep
}

func DefaultEndPoint() *EndPoint {
	return &EndPoint{
		Addr:    ":8080",
		TTL:     10,
		Network: "tcp",
	}
}

func (cfg *EndPoint) SetDefault() {
	//if cfg.Addr == "" {
	//	cfg.Addr = "127.0.0.1:8080"
	//}
	if cfg.TTL == 0 {
		cfg.TTL = 10
	}

	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
}

func (cfg *EndPoint) GetPort() int {
	if cfg.Addr == "" {
		return 0
	}

	seg := strings.Split(cfg.Addr, ":")
	if len(seg) == 1 {
		return 80
	}
	port, _ := strconv.Atoi(seg[1])
	return port
}

func (cfg *EndPoint) GetHost() string {
	if cfg.Addr == "" {
		return cfg.LocalIP() // "0.0.0.0"
	}
	seg := strings.Split(cfg.Addr, ":")
	return seg[0]
}

// 获取本机ip地址
func (cfg *EndPoint) LocalIP() string {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "0.0.0.0"
}

// 自动获取本机的ip以及端口号，ip:port格式
func (cfg *EndPoint) LocalListener(listener net.Listener, host string, err error) {
	host = "0.0.0.0:0"
	listener, err = net.Listen("tcp", host)
	if err == nil {
		addr := listener.Addr().String()
		_, portString, _ := net.SplitHostPort(addr)
		host = fmt.Sprintf("%s:%s", cfg.LocalIP(), portString)
	}
	return
}
