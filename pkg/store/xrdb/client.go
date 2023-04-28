package xrdb

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type Client struct {
	*gorm.DB
	config *Config
}

func (c *Client) Close() {
	db, err := c.DB.DB()
	if err == nil && db != nil {
		db.Close()
	}
}

func newClient(config *Config) (*Client, error) {
	dsn := checkConfig(config)

	var (
		inner *gorm.DB
		err   error
	)
	switch config.Type {
	case MysqlRDB:
		inner, err = gorm.Open(mysql.Open(config.DSN), &config.gormConfig)
	case SqlserverRDB:
		inner, err = gorm.Open(sqlserver.Open(config.DSN), &config.gormConfig)
	default:
		err = errors.New("unknown rdb type")
	}
	if err != nil {
		return nil, err
	}

	// 设置默认连接配置
	db, err := inner.DB()
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	if config.ConnMaxLifetime != 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.Debug {
		inner = inner.Debug()
	} else {
		inner.Logger = logger.Default.LogMode(logger.Silent)
	}

	if config.EnableTrace {
		attrs := []attribute.KeyValue{
			semconv.DBConnectionStringKey.String(dsn.Addr),
			semconv.DBUserKey.String(dsn.User),
		}
		if err := inner.Use(otelgorm.NewPlugin(
			otelgorm.WithDBName(dsn.DBName),
			otelgorm.WithAttributes(attrs...),
		)); err != nil {
			return nil, err
		}
	}

	return &Client{inner, config}, err
}

func checkConfig(config *Config) *DSN {
	if config.DSN == "" {
		panic(fmt.Sprintf("got empty %s dsn", config.Name))
	}
	if config.Type == "" {
		panic(fmt.Sprintf("got empty %s rdb type", config.Name))
	}

	dsn, err := parseDSN(config.DSN, config.Type)
	if err != nil {
		panic(fmt.Sprintf("parse %s dsn error: %v", config.Name, err))
	}
	return dsn
}
