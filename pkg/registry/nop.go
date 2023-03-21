package registry

import (
	"context"

	"github.com/5idu/pilot/pkg/server"
)

// Nop registry, used for local development/debugging
type Local struct{}

func (n Local) GetService(ctx context.Context, key string) (*server.ServiceInfo, error) {
	panic("implement me")
}

// ListServices ...
func (n Local) ListServices(ctx context.Context, s string) ([]*server.ServiceInfo, error) {
	panic("implement me")
}

func (n Local) WatchServices(ctx context.Context, prefix string) (chan Endpoints, error) {
	panic("implement me")
}

// RegisterService ...
func (n Local) RegisterService(ctx context.Context, si *server.ServiceInfo) error {
	return nil
}

// UnregisterService ...
func (n Local) UnregisterService(ctx context.Context, si *server.ServiceInfo) error {
	return nil
}

// Close ...
func (n Local) Close() error { return nil }

// Close ...
func (n Local) Kind() string { return "local" }
