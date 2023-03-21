package conf

import (
	"encoding/json"
	"log"
	"path/filepath"

	"github.com/5idu/pilot/pkg/flag"
	"github.com/5idu/pilot/pkg/hooks"

	"gopkg.in/yaml.v3"
)

const DefaultEnvPrefix = "PILOT_"

func init() {
	flag.Register(&flag.StringFlag{Name: "envPrefix", Usage: "--envPrefix=PILOT_", Default: DefaultEnvPrefix, Action: func(key string, fs *flag.FlagSet) {
		var envPrefix = fs.String(key)
		defaultConfiguration.LoadEnvironments(envPrefix)
	}})

	flag.Register(&flag.StringFlag{Name: "config", Usage: "--config=config.yaml", Action: func(key string, fs *flag.FlagSet) {
		hooks.Do(hooks.Stage_BeforeLoadConfig)

		var configAddr = fs.String(key)
		log.Printf("read config: %s", configAddr)
		datasource, err := NewDataSource(configAddr)
		if err != nil {
			log.Fatalf("build datasource[%s] failed: %v", configAddr, err)
		}

		unmarshaler := yaml.Unmarshal
		switch filepath.Ext(configAddr) {
		case ".yaml", ".yml":
			unmarshaler = yaml.Unmarshal
		case ".json":
			unmarshaler = json.Unmarshal
		default:
			log.Fatalf("unsupported config type: %s", filepath.Ext(configAddr))
		}

		if err := LoadFromDataSource(datasource, unmarshaler); err != nil {
			log.Fatalf("load config from datasource[%s] failed: %v", configAddr, err)
		}
		log.Printf("load config from datasource[%s] completely!", configAddr)

		hooks.Do(hooks.Stage_AfterLoadConfig)
	}})

	flag.Register(&flag.StringFlag{Name: "config-tag", Usage: "--config-tag=mapstructure", Default: "mapstructure", Action: func(key string, fs *flag.FlagSet) {
		defaultGetOptions.TagName = fs.String("config-tag")
	}})

	flag.Register(&flag.BoolFlag{Name: "watch", Usage: "--watch, watch config change event", Default: false, EnvVar: "PILOT_CONFIG_WATCH", Action: func(key string, fs *flag.FlagSet) {
		log.Printf("load config watch: %v", fs.Bool(key))
	}})
}
