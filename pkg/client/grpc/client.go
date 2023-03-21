package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/5idu/pilot/pkg/client/grpc/resolver"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConn = grpc.ClientConn

func newGRPCClient(config *Config) *grpc.ClientConn {
	var ctx = context.Background()

	dialOptions := getDialOptions(config)

	// 默认使用block连接，失败后fallback到异步连接
	if config.DialTimeout > time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.DialTimeout)

		defer cancel()
	}

	conn, err := grpc.DialContext(ctx, config.Addr, append(dialOptions, grpc.WithBlock())...)
	if err != nil {
		config.logger.Error("dial grpc server failed, connect without block", xlog.FieldExtra(map[string]interface{}{"error": err.Error()}))

		conn, err = grpc.DialContext(context.Background(), config.Addr, dialOptions...)
		if err != nil {
			panic(errors.WithMessage(err, "connect without block failed"))
		}
	}

	return conn
}

func getDialOptions(config *Config) []grpc.DialOption {
	dialOptions := config.dialOptions

	if config.KeepAlive != nil {
		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(*config.KeepAlive))
	}

	dialOptions = append(dialOptions,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithResolvers(resolver.NewEtcdBuilder("etcd", config.RegistryConfig)),
		grpc.WithDisableServiceConfig(),
	)

	svcCfg := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, config.BalancerName)
	dialOptions = append(dialOptions, grpc.WithDefaultServiceConfig(svcCfg))

	return dialOptions
}
