package file

import (
	"github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/flag"
	"github.com/5idu/pilot/pkg/xlog"
)

// DataSourceFile defines file scheme
const DataSourceFile = "file"

func init() {
	conf.Register(DataSourceFile, func() conf.DataSource {
		var (
			watchConfig = flag.Bool("watch")
			configAddr  = flag.String("config")
		)
		if configAddr == "" {
			xlog.Panic("new file dataSource, configAddr is empty")
			return nil
		}
		return NewDataSource(configAddr, watchConfig)
	})
}
