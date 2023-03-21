package registry

import (
	"context"
	"io"

	"github.com/5idu/pilot/pkg/server"
)

// Registry register/unregister service
// registry impl should control rpc timeout
type Registry interface {
	RegisterService(context.Context, *server.ServiceInfo) error
	UnregisterService(context.Context, *server.ServiceInfo) error
	GetService(context.Context, string) (*server.ServiceInfo, error)
	ListServices(context.Context, string) ([]*server.ServiceInfo, error)
	WatchServices(context.Context, string) (chan Endpoints, error)
	Kind() string
	io.Closer
}
