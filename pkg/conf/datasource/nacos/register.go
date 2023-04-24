package nacos

import (
	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/flag"
	"github.com/5idu/pilot/pkg/util/xnet"
	"github.com/5idu/pilot/pkg/xlog"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	xcast "github.com/spf13/cast"
)

// DataSourceNacos defines nacos scheme
const DataSourceNacos = "nacos"

func init() {
	conf.Register(DataSourceNacos, func() conf.DataSource {
		var (
			configAddr = flag.String("config")
		)
		if configAddr == "" {
			xlog.Panic("new nacos dataSource, configAddr is empty")
			return nil
		}
		// configAddr is a string in this format:
		// nacos://ip:port?data_id=xx&group=xx&namespace_id=xx&timeout=10000&access_key=xx&secret_key=xx&not_load_cache_at_start=true&update_cache_when_empty=true
		urlObj, err := xnet.ParseURL(configAddr)
		if err != nil {
			xlog.Panic("parse configAddr error", xlog.Any("error", err))
			return nil
		}
		// create clientConfig
		clientConfig := constant.ClientConfig{
			TimeoutMs:            urlObj.QueryUint64("timeout", 10000),
			NotLoadCacheAtStart:  urlObj.QueryBool("not_load_cache_at_start", true),
			UpdateCacheWhenEmpty: urlObj.QueryBool("update_cache_when_empty", true),
			NamespaceId:          urlObj.Query().Get("namespace_id"),
			AccessKey:            urlObj.Query().Get("access_key"),
			SecretKey:            urlObj.Query().Get("secret_key"),
		}
		// create serverConfigs
		serverConfigs := []constant.ServerConfig{
			{
				IpAddr: urlObj.HostName,
				Port:   getPort(urlObj.Port, 8848),
			},
		}
		// create config client
		client, err := clients.NewConfigClient(
			vo.NacosClientParam{
				ClientConfig:  &clientConfig,
				ServerConfigs: serverConfigs,
			},
		)
		if err != nil {
			xlog.Panic("create config client error", xlog.Any("error", err))
			return nil
		}
		return NewDataSource(client, urlObj.Query().Get("group"), urlObj.Query().Get("data_id"))
	})
}

func getPort(port string, expect uint64) uint64 {
	ret, err := xcast.ToUint64E(port)
	if err != nil {
		return expect
	}
	return ret
}
