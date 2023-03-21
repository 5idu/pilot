package xmetric

import (
	"log"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/xmetric/otelgrpc"
)

func init() {
	// 加载完配置，初始化 metric
	conf.OnLoaded(func(c *conf.Configuration) {
		log.Println("hook config, init metric config")

		key := constant.ConfigKey("metric.otelgrpc")
		if conf.Get(key) != nil {
			var config = otelgrpc.RawConfig(key)
			config.Build()
			return
		}
	})
}
