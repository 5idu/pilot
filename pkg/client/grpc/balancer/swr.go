package balancer

import (
	"errors"
	"sync"

	"github.com/smallnest/weighted"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const (
	// NameSmoothWeightRoundRobin ...
	NameSmoothWeightRoundRobin = "swr"
)

// PickerBuildInfo ...
type PickerBuildInfo struct {
	// ReadySCs is a map from all ready SubConns to the Addresses used to
	// create them.
	ReadySCs map[balancer.SubConn]base.SubConnInfo
	*attributes.Attributes
}

// PickerBuilder ...
type PickerBuilder interface {
	Build(info PickerBuildInfo) balancer.Picker
}

func init() {
	balancer.Register(
		NewBalancerBuilderV2(NameSmoothWeightRoundRobin, &swrPickerBuilder{}, base.Config{HealthCheck: true}),
	)
}

type swrPickerBuilder struct{}

// Build ...
func (s swrPickerBuilder) Build(info PickerBuildInfo) balancer.Picker {
	return newSWRPicker(info)
}

type swrPicker struct {
	readySCs map[balancer.SubConn]base.SubConnInfo
	mu       sync.Mutex
	// next         int
	buckets      *weighted.SW
	routeBuckets map[string]*weighted.SW
	*attributes.Attributes
}

func newSWRPicker(info PickerBuildInfo) *swrPicker {
	picker := &swrPicker{
		buckets:      &weighted.SW{},
		readySCs:     info.ReadySCs,
		routeBuckets: map[string]*weighted.SW{},
	}
	picker.parseBuildInfo(info)
	return picker
}

// Pick ...
func (p *swrPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var buckets = p.buckets
	if bs, ok := p.routeBuckets[info.FullMethodName]; ok {
		// 根据URI进行流量分组路由
		buckets = bs
	}

	sub, ok := buckets.Next().(balancer.SubConn)
	if ok {
		return balancer.PickResult{SubConn: sub}, nil
	}

	return balancer.PickResult{}, errors.New("pick failed")
}

func (p *swrPicker) parseBuildInfo(info PickerBuildInfo) {
	var hostedSubConns = map[string]balancer.SubConn{}

	for subConn, info := range info.ReadySCs {
		p.buckets.Add(subConn, 1)
		host := info.Address.Addr
		hostedSubConns[host] = subConn
		p.buckets.Add(subConn, 1)
	}

	if info.Attributes == nil {
		return
	}
}
