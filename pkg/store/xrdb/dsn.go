package xrdb

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"
	"github.com/microsoft/go-mssqldb/msdsn"
)

// DSN ...
type DSN struct {
	User     string            // Username
	Password string            // Password (requires User)
	Net      string            // Network type
	Addr     string            // Network address (requires Net)
	DBName   string            // Database name
	Params   map[string]string // Connection parameters
}

// parseDSN parses the DSN string to a NodeConfig.
func parseDSN(dsn string, rdbtype RdbType) (cfg *DSN, err error) {
	cfg = new(DSN)
	switch rdbtype {
	case MysqlRDB:
		parsedCfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}
		if err := copier.Copy(cfg, parsedCfg); err != nil {
			return nil, err
		}
	case SqlserverRDB:
		parsedCfg, err := msdsn.Parse(dsn)
		if err != nil {
			return nil, err
		}
		cfg = &DSN{
			User:     parsedCfg.User,
			Password: parsedCfg.Password,
			Net:      parsedCfg.Protocols[0],
			Addr:     parsedCfg.Host,
			DBName:   parsedCfg.Database,
			Params:   parsedCfg.Parameters,
		}
	}
	return
}
