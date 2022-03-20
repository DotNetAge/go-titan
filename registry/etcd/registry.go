package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dotnetage/go-titan/registry"
	"github.com/dotnetage/go-titan/runtime"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// ETCDRegistry for grpc server
type ETCDRegistry struct {
	// EtcdAddrs   []string
	// DialTimeout int
	options     *registry.Options
	closeCh     chan struct{}
	leasesID    clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse

	srvInfo *runtime.ServiceDesc
	srvTTL  int
	cli     *clientv3.Client
	logger  *zap.Logger
}

// NewRegister create a register base on etcd
func NewETCDRegistry(opts ...registry.Option) registry.Registry {
	options := registry.DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	reg := &ETCDRegistry{options: options}
	reg.logger = options.Logger
	return reg
}

// ETCDRegistry a service
func (r *ETCDRegistry) Register(srvInfo *runtime.ServiceDesc) error {
	var err error

	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return errors.New("无效的IP")
	}

	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.options.Endpoints,
		DialTimeout: time.Duration(r.options.DialTimeout) * time.Second,
	}); err != nil {
		return err
	}

	r.srvInfo = srvInfo
	r.srvTTL = srvInfo.EndPoint.TTL

	if err = r.register(); err != nil {
		return err
	}
	r.logger.Info(fmt.Sprintf("%v 服务已成功注册", srvInfo.Name))
	r.closeCh = make(chan struct{})

	go r.keepAlive()

	return nil
	//return r.closeCh, nil
}

// Unregister stop register
func (r *ETCDRegistry) Unregister() error {
	r.closeCh <- struct{}{}
	return nil
}

// // UpdateHandler return http handler
// func (r *ETCDRegistry) UpdateHandler() http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
// 		wi := req.URL.Query().Get("weight")
// 		weight, err := strconv.Atoi(wi)
// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}
// 		var update = func() error {
// 			r.srvInfo.Weight = int64(weight)
// 			data, err := json.Marshal(r.srvInfo)
// 			if err != nil {
// 				return err
// 			}
// 			_, err = r.cli.Put(context.Background(), runtime.BuildRegPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
// 			return err
// 		}
// 		if err := update(); err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(err.Error()))
// 			return
// 		}
// 		w.Write([]byte("成功更新服务器权重"))
// 	})
// }

func (r *ETCDRegistry) GetServices() ([]*runtime.ServiceDesc, error) {
	resp, err := r.cli.Get(context.Background(), runtime.BuildRegPath(r.srvInfo))
	if err != nil {
		return []*runtime.ServiceDesc{r.srvInfo}, err
	}

	svcs := make([]*runtime.ServiceDesc, 0)
	count := int(resp.Count)

	for i := 0; i < count; i++ {
		svc := runtime.ServiceDesc{}
		if err := json.Unmarshal(resp.Kvs[i].Value, &svc); err != nil {
			return svcs, err
		}
		svcs = append(svcs, &svc)
	}
	return svcs, nil
}

// register 注册节点
func (r *ETCDRegistry) register() error {
	leaseCtx, cancel := context.WithTimeout(context.Background(), time.Duration(r.options.DialTimeout)*time.Second)
	defer cancel()

	leaseResp, err := r.cli.Grant(leaseCtx, int64(r.srvTTL))
	if err != nil {
		return err
	}
	r.leasesID = leaseResp.ID
	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), leaseResp.ID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}
	_, err = r.cli.Put(context.Background(), runtime.BuildRegPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
	return err
}

// unregister 删除节点
func (r *ETCDRegistry) unregister() error {
	_, err := r.cli.Delete(context.Background(), runtime.BuildRegPath(r.srvInfo))
	return err
}

// keepAlive 轮询服务状态并进行重新注册
func (r *ETCDRegistry) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)
	for {
		select {
		case <-r.closeCh:
			fmt.Println("正在注销服务")
			if err := r.unregister(); err != nil {
				r.logger.Error("注销失败", zap.Error(err))
			}
			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
				r.logger.Error("revoke failed", zap.Error(err))
			}
			return
		case res := <-r.keepAliveCh:
			if res == nil {
				if err := r.register(); err != nil {
					r.logger.Error("注册失败", zap.Error(err))
				}
			}
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					r.logger.Error("注册失败", zap.Error(err))
				}
			}
		}
	}
}
