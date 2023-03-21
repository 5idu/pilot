package xtrace

import (
	"log"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/xtrace/jaeger"
	"github.com/5idu/pilot/pkg/xtrace/otelgrpc"
)

func init() {
	// 加载完配置，初始化 trace
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Println("hook config, init trace config")

		key := constant.ConfigKey("trace.jaeger")
		if conf.Get(key) != nil {
			var config = jaeger.RawConfig(key)
			config.Build()
			return
		}

		key = constant.ConfigKey("trace.otelgrpc")
		if conf.Get(key) != nil {
			var config = otelgrpc.RawConfig(key)
			config.Build()
			return
		}
	})
}
