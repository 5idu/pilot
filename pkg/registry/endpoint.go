package registry

import "github.com/5idu/pilot/pkg/server"

type Endpoints struct {
	// 服务节点列表
	Nodes map[string]server.ServiceInfo
}

func newEndpoints() *Endpoints {
	return &Endpoints{
		Nodes: make(map[string]server.ServiceInfo),
	}
}

func (in *Endpoints) DeepCopy() *Endpoints {
	if in == nil {
		return nil
	}

	out := newEndpoints()
	in.DeepCopyInfo(out)
	return out
}

func (in *Endpoints) DeepCopyInfo(out *Endpoints) {
	for key, info := range in.Nodes {
		out.Nodes[key] = info
	}
}
