package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/5idu/pilot/pkg/util/xstring"
	"github.com/5idu/pilot/pkg/xlog"
	"github.com/5idu/pilot/pkg/xmetric"
	"github.com/5idu/pilot/pkg/xtrace"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	errSlowCommand = errors.New("grpc unary slow command")
)

// timeoutUnaryClientInterceptor gRPC客户端超时拦截器
func timeoutUnaryClientInterceptor(_logger *xlog.Logger, timeout time.Duration, slowThreshold time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		now := time.Now()
		// 若无自定义超时设置，默认设置超时
		_, ok := ctx.Deadline()
		if !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		err := invoker(ctx, method, req, reply, cc, opts...)
		du := time.Since(now)
		remoteIP := "unknown"
		if remote, ok := peer.FromContext(ctx); ok && remote.Addr != nil {
			remoteIP = remote.Addr.String()
		}

		if slowThreshold > time.Duration(0) && du > slowThreshold {
			_logger.Error("slow", xlog.FieldExtra(map[string]interface{}{
				"error":  errSlowCommand.Error(),
				"method": method,
				"name":   cc.Target(),
				"cost":   du,
				"addr":   remoteIP,
			}))
		}
		return err
	}
}

// loggerUnaryClientInterceptor gRPC客户端日志中间件
func loggerUnaryClientInterceptor(_logger *xlog.Logger, name string, accessInterceptorLevel string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)

		if err != nil {
			// 业务报错只做 warning
			_logger.Warn(
				"access",
				xlog.FieldExtra(map[string]interface{}{
					"type":   "unary",
					"name":   name,
					"method": method,
					"cost":   time.Since(beg),
					"req":    json.RawMessage(xstring.Json(req)),
					"reply":  json.RawMessage(xstring.Json(reply)),
				}),
			)
			return err
		} else {
			if accessInterceptorLevel == "info" {
				_logger.Info(
					"access",
					xlog.FieldExtra(map[string]interface{}{
						"type":   "unary",
						"name":   name,
						"method": method,
						"cost":   time.Since(beg),
						"req":    json.RawMessage(xstring.Json(req)),
						"reply":  json.RawMessage(xstring.Json(reply)),
					}),
				)
			}
		}

		return nil
	}
}

func traceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.RPCSystemGRPC,
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		ctx, span := tracer.Start(ctx, method, xtrace.MetadataReaderWriter(md), trace.WithAttributes(attrs...))
		ctx = metadata.NewOutgoingContext(ctx, md)
		span.SetAttributes(
			semconv.RPCMethodKey.String(method),
		)

		err = invoker(ctx, method, req, reply, cc, opts...)

		span.SetStatus(codes.Ok, "ok")

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()

		return err
	}
}

func metricUnaryClientInterceptor(name string) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		attrs := []attribute.KeyValue{
			attribute.String("name", name),
			attribute.String("method", method),
			attribute.String("target", cc.Target()),
		}
		if err != nil {
			xmetric.GRPCCallFault.Inc(ctx, attrs...)
		}
		xmetric.GRPCCallDuration.Record(ctx, time.Since(beg), attrs...)
		return err
	}
}
