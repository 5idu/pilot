package resolver

import (
	"context"
	"strings"

	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/registry/etcdv3"
	"github.com/5idu/pilot/pkg/util/xgo"
	"github.com/5idu/pilot/pkg/xlog"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

// NewEtcdBuilder returns a new etcdv3 resolver builder.
func NewEtcdBuilder(name string, registryConfig string) resolver.Builder {
	return &baseBuilder{
		name:           name,
		registryConfig: registryConfig,
	}
}

type baseBuilder struct {
	name string

	registryConfig string
}

// Build ...
func (b *baseBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	reg := etcdv3.RawConfig(b.registryConfig).MustSingleton()

	if !strings.HasSuffix(target.Endpoint, "/") {
		target.Endpoint += "/"
	}

	endpoints, err := reg.WatchServices(context.Background(), target.Endpoint)
	if err != nil {
		xlog.Error("watch services failed", xlog.Any("error", err))
		return nil, err
	}

	var stop = make(chan struct{})
	xgo.Go(func() {
		for {
			select {
			case endpoint := <-endpoints:
				xlog.Debug("watch services finished", xlog.Any("value", endpoint))

				var state = resolver.State{
					Addresses: make([]resolver.Address, 0),
				}
				for _, node := range endpoint.Nodes {
					var address resolver.Address
					address.Addr = node.Address
					address.ServerName = target.Endpoint
					address.Attributes = attributes.New(constant.KeyServiceInfo, node)
					state.Addresses = append(state.Addresses, address)
				}
				_ = cc.UpdateState(state)
			case <-stop:
				return
			}
		}
	})

	return &baseResolver{
		stop: stop,
	}, nil
}

// Scheme ...
func (b baseBuilder) Scheme() string {
	return b.name
}

type baseResolver struct {
	stop chan struct{}
}

// ResolveNow ...
func (b *baseResolver) ResolveNow(options resolver.ResolveNowOptions) {}

// Close ...
func (b *baseResolver) Close() { b.stop <- struct{}{} }
