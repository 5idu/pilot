package xrdb

import (
	"time"

	cfg "github.com/5idu/pilot/pkg/conf"
	"github.com/5idu/pilot/pkg/constant"
	"github.com/5idu/pilot/pkg/singleton"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// StdConfig .
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("rdb." + name))
}

// RawConfig .
func RawConfig(key string) *Config {
	config := DefaultConfig()
	config.Name = key

	if err := cfg.UnmarshalKey(key, &config); err != nil {
		xlog.Panic("unmarshal config", xlog.FieldErr(err), xlog.String("name", key))
	}

	return config
}

// Config options
type Config struct {
	Name string
	// DSN地址: mysql://root:secret@tcp(127.0.0.1:3307)/mysql?timeout=20s&readTimeout=20s
	DSN string `json:"dsn"`
	// rdb type
	Type RdbType `json:"type"`
	// Debug开关
	Debug bool `json:"debug"`
	// 最大空闲连接数
	MaxIdleConns int `json:"maxIdleConns"`
	// 最大活动连接数
	MaxOpenConns int `json:"maxOpenConns"`
	// 连接的最大存活时间
	ConnMaxLifetime time.Duration `json:"connMaxLifetime"`

	gormConfig gorm.Config `json:"-"`

	EnableTrace  bool `json:"enableTrace"`
	EnableMetric bool `json:"enableMetric"`
}

type RdbType string

const (
	MysqlRDB     RdbType = "mysql"
	SqlserverRDB RdbType = "sqlserver"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DSN:             "",
		Debug:           false,
		MaxIdleConns:    10,
		MaxOpenConns:    10,
		ConnMaxLifetime: cast.ToDuration("300s"),
		EnableTrace:     true,
		EnableMetric:    true,
	}
}

func (config *Config) WithGormConfig(gormConfig gorm.Config) *Config {
	config.gormConfig = gormConfig
	return config
}

func (config *Config) MustBuild() *Client {
	client, err := newClient(config)
	if err != nil {
		panic(err)
	}
	return client
}

func (config *Config) MustSingleton() *Client {
	if val, ok := singleton.Load(constant.ModuleStoreRDB, config.Name); ok && val != nil {
		return val.(*Client)
	}

	db := config.MustBuild()
	singleton.Store(constant.ModuleStoreRDB, config.Name, db)
	return db
}
