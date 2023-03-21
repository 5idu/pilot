package etcdv3

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/5idu/pilot/pkg/client/etcdv3"
	"github.com/5idu/pilot/pkg/registry"
	"github.com/5idu/pilot/pkg/server"
	"github.com/5idu/pilot/pkg/util/xgo"
	"github.com/5idu/pilot/pkg/util/xretry"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/bytedance/sonic"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdv3Registry struct {
	ctx    context.Context
	client *etcdv3.Client
	kvs    sync.Map
	*Config
	cancel  context.CancelFunc
	rmu     sync.RWMutex
	leaseID clientv3.LeaseID

	once sync.Once
}

const (
	// defaultRetryTimes default retry times
	defaultRetryTimes = 3
	// defaultKeepAliveTimeout is the default timeout for keepalive requests.
	defaultRegisterTimeout = 5 * time.Second
)

var _ registry.Registry = new(etcdv3Registry)

func newETCDRegistry(config *Config) (*etcdv3Registry, error) {
	if config.logger == nil {
		config.logger = xlog.With(xlog.String("mod", "registry.etcd"))
	}
	etcdv3Client, err := config.Config.Singleton()
	if err != nil {
		config.logger.Error("create etcdv3 client error", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "addr": config.Config.Endpoints}))
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	reg := &etcdv3Registry{
		ctx:    ctx,
		cancel: cancel,
		client: etcdv3Client,
		Config: config,
		kvs:    sync.Map{},
		rmu:    sync.RWMutex{},
	}

	return reg, nil
}

func (reg *etcdv3Registry) Kind() string { return "etcdv3" }

// RegisterService register service to registry
func (reg *etcdv3Registry) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	return reg.registerBiz(ctx, info)
}

// UnregisterService unregister service from registry
func (reg *etcdv3Registry) UnregisterService(ctx context.Context, info *server.ServiceInfo) error {
	return reg.unregister(ctx, reg.registerKey(info))
}

// GetService get service registered in registry with name `name`
func (reg *etcdv3Registry) GetService(ctx context.Context, key string) (*server.ServiceInfo, error) {
	getResp, getErr := reg.client.Get(ctx, key)
	if getErr != nil {
		reg.logger.Error("registey get service failed", xlog.FieldExtra(map[string]interface{}{"error": getErr.Error(), "key": key}))
		return nil, getErr
	}

	var service server.ServiceInfo
	if err := json.Unmarshal(getResp.Kvs[0].Value, &service); err != nil {
		reg.logger.Warn("invalid service", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "key": string(getResp.Kvs[0].Key), "value": string(getResp.Kvs[0].Value)}))
		return nil, err
	}

	return &service, nil
}

// ListServices list service registered in registry with name `name`
func (reg *etcdv3Registry) ListServices(ctx context.Context, prefix string) (services []*server.ServiceInfo, err error) {
	getResp, getErr := reg.client.Get(ctx, prefix, clientv3.WithPrefix())
	if getErr != nil {
		reg.logger.Error("reg.client.Get failed", xlog.FieldExtra(map[string]interface{}{"error": getErr.Error(), "addr": prefix}))
		return nil, getErr
	}

	for _, kv := range getResp.Kvs {
		var service server.ServiceInfo
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			reg.logger.Warn("invalid service", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "key": string(kv.Key), "value": string(kv.Value)}))
			continue
		}

		services = append(services, &service)
	}

	return
}

// WatchServices watch service change event, then return address list
func (reg *etcdv3Registry) WatchServices(ctx context.Context, prefix string) (chan registry.Endpoints, error) {
	watch, err := reg.client.WatchPrefix(context.Background(), prefix)
	if err != nil {
		reg.logger.Error("reg.client.WatchPrefix failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "addr": prefix}))
		return nil, err
	}

	var addresses = make(chan registry.Endpoints, 10)
	var al = &registry.Endpoints{
		Nodes: make(map[string]server.ServiceInfo),
	}

	scheme := getScheme(prefix)

	for _, kv := range watch.IncipientKeyValues() {
		updateAddrList(al, prefix, scheme, kv)
	}

	// var snapshot registry.Endpoints
	// xstruct.CopyStruct(al, &snapshot)
	addresses <- *al.DeepCopy()

	xgo.Go(func() {
		for event := range watch.C() {
			switch event.Type {
			case mvccpb.PUT:
				updateAddrList(al, prefix, scheme, event.Kv)
			case mvccpb.DELETE:
				deleteAddrList(al, prefix, scheme, event.Kv)
			}

			// var snapshot registry.Endpoints
			// xstruct.CopyStruct(al, &snapshot)
			out := al.DeepCopy()
			select {
			// case addresses <- snapshot:
			case addresses <- *out:
			default:
				reg.logger.Warn("invalid event")
			}
		}
	})

	return addresses, nil
}

func (reg *etcdv3Registry) unregister(ctx context.Context, key string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, reg.ReadTimeout)
		defer cancel()
	}

	_, err := reg.client.Delete(ctx, key)
	if err == nil {
		reg.kvs.Delete(key)
	}
	return err
}

// Close ...
func (reg *etcdv3Registry) Close() error {
	if reg.cancel != nil {
		reg.cancel()
	}
	var wg sync.WaitGroup
	reg.kvs.Range(func(k, v interface{}) bool {
		wg.Add(1)
		go func(k interface{}) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := reg.unregister(ctx, k.(string))
			if err != nil {
				reg.logger.Error("unregister service",
					xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "key": k, "value": v}))
			} else {
				reg.logger.Info("unregister service", xlog.FieldExtra(map[string]interface{}{"key": k, "value": v}))
			}
			cancel()
		}(k)
		return true
	})
	wg.Wait()
	return nil
}

func (reg *etcdv3Registry) registerBiz(ctx context.Context, info *server.ServiceInfo) error {
	key := reg.registerKey(info)
	val := reg.registerValue(info)

	return reg.registerKV(ctx, key, val)
}

func (reg *etcdv3Registry) registerKV(ctx context.Context, key, val string) error {

	opOptions := make([]clientv3.OpOption, 0)
	// opOptions = append(opOptions, clientv3.WithSerializable())
	if ttl := reg.Config.ServiceTTL.Seconds(); ttl > 0 {
		// 这里基于应用名为key做缓存，每个服务实例应该只需要创建一个lease，降低etcd的压力
		lease, err := reg.getOrGrantLeaseID(ctx)
		if err != nil {
			reg.logger.Error("getSession failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "key": key, "value": val}))
			return err
		}

		reg.once.Do(func() {
			// we use reg.ctx to manully cancel lease keepalive loop
			go reg.doKeepalive(reg.ctx)
		})

		opOptions = append(opOptions, clientv3.WithLease(lease))
	}
	_, err := reg.client.Put(ctx, key, val, opOptions...)
	if err != nil {
		reg.logger.Error("register service", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "key": key}))
		return err
	}

	reg.logger.Info("register service", xlog.FieldExtra(map[string]interface{}{"key": key, "value": val}))
	reg.kvs.Store(key, val)
	return nil
}

func (reg *etcdv3Registry) getOrGrantLeaseID(ctx context.Context) (clientv3.LeaseID, error) {
	reg.rmu.Lock()
	defer reg.rmu.Unlock()

	if reg.leaseID != 0 {
		return reg.leaseID, nil
	}

	grant, err := reg.client.Grant(ctx, int64(reg.ServiceTTL.Seconds()))
	if err != nil {
		reg.logger.Error("reg.client.Grant failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error()}))
		return 0, err
	}

	reg.leaseID = grant.ID

	return grant.ID, nil
}

func (reg *etcdv3Registry) getLeaseID() clientv3.LeaseID {
	reg.rmu.RLock()
	defer reg.rmu.RUnlock()

	return reg.leaseID
}

func (reg *etcdv3Registry) setLeaseID(leaseId clientv3.LeaseID) {
	reg.rmu.Lock()
	defer reg.rmu.Unlock()

	reg.leaseID = leaseId
}

// doKeepAlive periodically sends keep alive requests to etcd server.
// when the keep alive request fails or timeout, it will try to re-establish the lease.
func (reg *etcdv3Registry) doKeepalive(ctx context.Context) {

	reg.logger.Debug("start keepalive...")

	kac, err := reg.client.KeepAlive(ctx, reg.getLeaseID())
	if err != nil {
		reg.setLeaseID(0)
		reg.logger.Error("reg.client.KeepAlive failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error()}))
	}

	timer := time.NewTimer(defaultRegisterTimeout)
	defer timer.Stop()
	for {
		// we should register again, because the leaseID is 0
		if reg.getLeaseID() == 0 {
			cancelCtx, cancel := context.WithCancel(ctx)

			done := make(chan struct{}, 1)

			go func() {
				// do register again, and retry 3 times
				err := reg.registerAllKvs(cancelCtx)
				if err != nil {
					cancel()
					return
				}

				done <- struct{}{}
			}()

			// wait registerAllKvs success
			select {
			case <-timer.C:
				// when timeout happens
				// we should cancel the context and retry again
				cancel()
				// mark leaseID as 0 to retry register
				reg.setLeaseID(0)

				continue
			case <-done:
				// when done happens, we just receive the kac channel
				// or wait the registry context done
			}

			// try do keepalive again
			// when error or timeout happens, just continue and try again
			kac, err = reg.client.KeepAlive(ctx, reg.getLeaseID())
			if err != nil {
				reg.logger.Error("reg.client.KeepAlive failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error()}))
				time.Sleep(defaultRegisterTimeout)
				continue
			}

			reg.logger.Debug("reg.client.KeepAlive finished", xlog.FieldExtra(map[string]interface{}{"leaseid": fmt.Sprintf("%x", reg.getLeaseID())}))
		}

		select {
		case data, ok := <-kac:
			if !ok {
				// when error happens
				// mark leaseID as 0 to retry register
				reg.setLeaseID(0)

				reg.logger.Debug("need to retry registration", xlog.FieldExtra(map[string]interface{}{"leaseid": fmt.Sprintf("%x", reg.getLeaseID())}))

				continue
			}

			// just record detailed keepalive info
			reg.logger.Debug("do keepalive", xlog.FieldExtra(map[string]interface{}{"data": data, "leaseid": fmt.Sprintf("%x", reg.getLeaseID())}))
		case <-reg.ctx.Done():
			reg.logger.Debug("exit keepalive")

			return
		}
	}
}

func (reg *etcdv3Registry) registerKey(info *server.ServiceInfo) string {
	return info.RegistryName()
}

func (reg *etcdv3Registry) registerValue(info *server.ServiceInfo) string {
	val, _ := sonic.MarshalString(info)
	return val
}

func (reg *etcdv3Registry) registerAllKvs(ctx context.Context) error {
	// do register again, and retry 3 times
	return xretry.Do(defaultRetryTimes, time.Second, func() error {
		var err error

		// all kvs stored in reg.kvs, and we can range this map to register again
		reg.kvs.Range(func(key, value any) bool {
			err = reg.registerKV(ctx, key.(string), value.(string))
			if err != nil {
				reg.logger.Error("registerKV failed", xlog.FieldExtra(map[string]interface{}{"error": err.Error(), "key": key, "value": value}))
			}

			return err == nil
		})

		return err
	})
}

func deleteAddrList(al *registry.Endpoints, prefix, scheme string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix)

		if isIPPort(addr) {
			// 直接删除addr 因为Delete操作的value值为空
			delete(al.Nodes, addr)
		}
	}
}

func updateAddrList(al *registry.Endpoints, prefix, scheme string, kvs ...*mvccpb.KeyValue) {
	for _, kv := range kvs {
		var addr = strings.TrimPrefix(string(kv.Key), prefix)
		if isIPPort(addr) {
			var meta server.ServiceInfo
			if err := json.Unmarshal(kv.Value, &meta); err != nil {
				xlog.Error("unmarshal meta", xlog.Any("error", err),
					xlog.String("value", string(kv.Value)), xlog.String("key", string(kv.Key)))
				continue
			}
			al.Nodes[addr] = meta
		}
	}
}

func isIPPort(addr string) bool {
	_, _, err := net.SplitHostPort(addr)
	return err == nil
}

func getScheme(prefix string) string {
	return strings.Split(prefix, ":")[0]
}
