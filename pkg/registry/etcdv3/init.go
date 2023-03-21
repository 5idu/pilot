package etcdv3

import (
	"github.com/5idu/pilot/pkg/registry"
)

func init() {
	registry.RegisterBuilder("etcdv3", func(confKey string) registry.Registry {
		return RawConfig(confKey).MustBuild()
	})
}
