package etcdv3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/5idu/pilot/pkg/xlog"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client ...
type Client struct {
	*clientv3.Client
	config *Config
}

// New ...
func newClient(config *Config) (*Client, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(grpcprom.StreamClientInterceptor),
	}

	if config.EnableTrace {
		dialOptions = append(dialOptions,
			grpc.WithChainUnaryInterceptor(traceUnaryClientInterceptor()),
			grpc.WithChainStreamInterceptor(traceStreamClientInterceptor()),
		)
	}

	conf := clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          config.ConnectTimeout,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		DialOptions:          dialOptions,
		AutoSyncInterval:     config.AutoSyncInterval,
	}

	config.logger = xlog.With(xlog.String("mod", "etcdv3"))

	if config.Endpoints == nil {
		return nil, fmt.Errorf("client etcd endpoints empty, empty endpoints")
	}

	if !config.Secure {
		conf.DialOptions = append(conf.DialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if config.BasicAuth {
		conf.Username = config.UserName
		conf.Password = config.Password
	}

	tlsEnabled := false
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if config.CaCert != "" {
		certBytes, err := os.ReadFile(config.CaCert)
		if err != nil {
			panic(errors.WithMessage(err, "parse CaCert failed"))
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
		tlsEnabled = true
	}

	if config.CertFile != "" && config.KeyFile != "" {
		tlsCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			panic(errors.WithMessage(err, "load CertFile or KeyFile failed"))
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		conf.TLS = tlsConfig
	}

	client, err := clientv3.New(conf)

	if err != nil {
		panic(errors.WithMessage(err, "client etcd start failed"))
	}

	cc := &Client{
		Client: client,
		config: config,
	}

	return cc, nil
}

// GetKeyValue queries etcd key, returns mvccpb.KeyValue
func (client *Client) GetKeyValue(ctx context.Context, key string) (kv *mvccpb.KeyValue, err error) {
	rp, err := client.Client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(rp.Kvs) > 0 {
		return rp.Kvs[0], nil
	}

	return
}

// GetPrefix get prefix
func (client *Client) GetPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	var (
		vars = make(map[string]string)
	)

	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return vars, err
	}

	for _, kv := range resp.Kvs {
		vars[string(kv.Key)] = string(kv.Value)
	}

	return vars, nil
}

// DelPrefix 按前缀删除
func (client *Client) DelPrefix(ctx context.Context, prefix string) (deleted int64, err error) {
	resp, err := client.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}
	return resp.Deleted, err
}

// GetValues queries etcd for keys prefixed by prefix.
func (client *Client) GetValues(ctx context.Context, keys ...string) (map[string]string, error) {
	var (
		firstRevision = int64(0)
		vars          = make(map[string]string)
		maxTxnOps     = 128
		// getOps        = make([]string, 0, maxTxnOps)
	)

	doTxn := func(ops []string) error {
		txnOps := make([]clientv3.Op, 0, maxTxnOps)

		for _, k := range ops {
			txnOps = append(txnOps, clientv3.OpGet(k,
				clientv3.WithPrefix(),
				clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
				clientv3.WithRev(firstRevision)))
		}

		result, err := client.Txn(ctx).Then(txnOps...).Commit()
		if err != nil {
			return err
		}
		for i, r := range result.Responses {
			originKey := ops[i]
			originKeyFixed := originKey
			if !strings.HasSuffix(originKeyFixed, "/") {
				originKeyFixed = originKey + "/"
			}
			for _, ev := range r.GetResponseRange().Kvs {
				k := string(ev.Key)
				if k == originKey || strings.HasPrefix(k, originKeyFixed) {
					vars[string(ev.Key)] = string(ev.Value)
				}
			}
		}
		if firstRevision == 0 {
			firstRevision = result.Header.GetRevision()
		}
		return nil
	}
	cnt := len(keys) / maxTxnOps
	for i := 0; i <= cnt; i++ {
		switch temp := (i == cnt); temp {
		case false:
			if err := doTxn(keys[i*maxTxnOps : (i+1)*maxTxnOps]); err != nil {
				return vars, err
			}
		case true:
			if err := doTxn(keys[i*maxTxnOps:]); err != nil {
				return vars, err
			}
		}
	}
	return vars, nil
}

// GetLeaseSession 创建租约会话
func (client *Client) GetLeaseSession(ctx context.Context, opts ...concurrency.SessionOption) (leaseSession *concurrency.Session, err error) {
	return concurrency.NewSession(client.Client, opts...)
}

func (client *Client) DelKeys(ctx context.Context, keys ...string) error {
	var (
		maxTxnOps = 128
	)

	doTxn := func(ops []string) error {
		txnOps := make([]clientv3.Op, 0, maxTxnOps)

		for _, k := range ops {
			txnOps = append(txnOps, clientv3.OpDelete(k))
		}

		_, err := client.Txn(ctx).Then(txnOps...).Commit()
		return err
	}
	cnt := len(keys) / maxTxnOps
	for i := 0; i <= cnt; i++ {
		switch temp := (i == cnt); temp {
		case false:
			if err := doTxn(keys[i*maxTxnOps : (i+1)*maxTxnOps]); err != nil {
				return err
			}
		case true:
			if err := doTxn(keys[i*maxTxnOps:]); err != nil {
				return err
			}
		}
	}
	return nil
}
