package registry

import (
	"log"

	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/xlog"
)

// var _registerers = sync.Map{}
var registryBuilder = make(map[string]Builder)

type Config map[string]struct {
	Kind          string `json:"kind" description:"底层注册器类型, eg: etcdv3, consul"`
	ConfigKey     string `json:"configKey" description:"底册注册器的配置键"`
	DeplaySeconds int    `json:"deplaySeconds" description:"延迟注册"`
}

// default register
var DefaultRegisterer Registry = &Local{}

func init() {
	// 初始化注册中心
	conf.OnLoaded(func(c *conf.Configuration) {
		xlog.Info("hook config, init registry")
		var config Config
		if err := c.UnmarshalKey(constant.ConfigKey("registry"), &config); err != nil {
			xlog.Infof("hook config, read registry config failed: %v", err)
			return
		}

		for name, item := range config {
			var itemKind = item.Kind
			if itemKind == "" {
				itemKind = "etcdv3"
			}

			if item.ConfigKey == "" {
				item.ConfigKey = constant.ConfigKey("registry.default")
			}

			build, ok := registryBuilder[itemKind]
			if !ok {
				xlog.Infof("invalid registry kind: %s", itemKind)
				continue
			}

			xlog.Infof("build registrerer %s with config: %s", name, item.ConfigKey)
			DefaultRegisterer = build(item.ConfigKey)
		}
	})
}

type Builder func(string) Registry

type BuildFunc func(string) (Registry, error)

func RegisterBuilder(kind string, build Builder) {
	if _, ok := registryBuilder[kind]; ok {
		log.Panicf("duplicate register registry builder: %s", kind)
	}
	registryBuilder[kind] = build
}
