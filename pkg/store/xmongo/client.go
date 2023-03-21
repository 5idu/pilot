package xmongo

import (
	"context"
	"time"

	"github.com/5idu/pilot/pkg/xmetric"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Client struct {
	*mongo.Client
	config          *Config
	metricCallbacks []metric.Registration
}

func newClient(config *Config) *Client {
	// check config param
	checkConfig(config)

	mps := uint64(config.PoolLimit)

	clientOpts := options.Client()
	clientOpts.MaxPoolSize = &mps
	clientOpts.SocketTimeout = &config.SocketTimeout

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOpts.ApplyURI(config.DSN))
	if err != nil {
		panic(errors.WithMessage(err, "dial mongo"))
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		panic(errors.WithMessage(err, "ping mongo"))
	}

	c := &Client{
		Client: client,
		config: config,
	}
	_instances.Store(config.Name, c)

	if config.EnableMetric {
		cb, err := xmetric.MongoDBClientSession.Observe(func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(xmetric.MongoDBClientSession.Int64ObservableUpDownCounter, int64(client.NumberSessionsInProgress()),
				attribute.String("name", config.Name),
			)
			return nil
		})
		if err != nil {
			panic(errors.WithMessage(err, "register metric callback"))
		}
		c.metricCallbacks = append(c.metricCallbacks, cb)
	}

	return c
}

func (c *Client) NewCollection(dbname string, coll string) *Collection {
	return newCollection(dbname, c.config, c.Database(dbname).Collection(coll))
}

func (c *Client) NewDatabase(dbname string) *Database {
	return newDatabase(dbname, c.config, c.Database(dbname))
}

func (c *Client) Close() error {
	if len(c.metricCallbacks) > 0 {
		for _, cb := range c.metricCallbacks {
			cb.Unregister()
		}
	}
	return c.Client.Disconnect(context.Background())
}

func checkConfig(config *Config) {
	if config.SocketTimeout == time.Duration(0) {
		panic(errors.New("invalid config: socketTimeout"))
	}

	if config.PoolLimit == 0 {
		panic(errors.New("invalid config: poolLimit"))
	}
}
