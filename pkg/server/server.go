package server

import (
	"context"
	"fmt"
)

type ServiceInfo struct {
	ID       string            `json:"id"`   // like: worker-xxxxxx
	Name     string            `json:"name"` // like: worker
	Scheme   string            `json:"scheme"`
	Address  string            `json:"address"`
	Hostname string            `json:"hostname"`
	Metadata map[string]string `json:"metadata"`
}

func (s *ServiceInfo) RegistryName() string {
	return fmt.Sprintf("%s:%s/%s", s.Scheme, s.Name, s.Address)
}

// Equal allows the values to be compared by Attributes.Equal, this change is in order
// to fit the change in grpc-go:
// attributes: add Equal method; resolver: add AddressMap and State.BalancerAttributes (#4855)
func (si ServiceInfo) Equal(o interface{}) bool {
	oa, ok := o.(ServiceInfo)
	return ok &&
		oa.Name == si.Name &&
		oa.Address == si.Address
}

type Option func(c *ServiceInfo)

func ApplyOptions(options ...Option) ServiceInfo {
	info := defaultServiceInfo()
	for _, option := range options {
		option(&info)
	}
	return info
}

func WithID(id string) Option {
	return func(c *ServiceInfo) {
		c.ID = id
	}
}

func WithName(name string) Option {
	return func(c *ServiceInfo) {
		c.Name = name
	}
}

func WithMetaData(key, value string) Option {
	return func(c *ServiceInfo) {
		c.Metadata[key] = value
	}
}

func WithScheme(scheme string) Option {
	return func(c *ServiceInfo) {
		c.Scheme = scheme
	}
}

func WithAddress(address string) Option {
	return func(c *ServiceInfo) {
		c.Address = address
	}
}

func WithHostname(hostname string) Option {
	return func(c *ServiceInfo) {
		c.Hostname = hostname
	}
}

func defaultServiceInfo() ServiceInfo {
	si := ServiceInfo{
		Metadata: make(map[string]string),
	}
	return si
}

// Server ...
type Server interface {
	Serve() error
	Stop() error
	GracefulStop(ctx context.Context) error
	Info() *ServiceInfo
	Healthz() bool
}
