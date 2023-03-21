package xgrpc

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

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
	"google.golang.org/grpc/status"
)

func defaultStreamServerInterceptor(logger *xlog.Logger, c *Config) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var beg = time.Now()
		var event = "normal"
		defer func() {
			if c.SlowQueryThresholdInMilli > 0 {
				if int64(time.Since(beg))/1e6 > c.SlowQueryThresholdInMilli {
					event = "slow"
				}
			}

			extra := map[string]interface{}{
				"type":   "unary",
				"method": info.FullMethod,
				"cost":   time.Since(beg),
				"event":  event,
			}
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				extra["stack"] = stack
				event = "recover"
			}

			for key, val := range getPeer(stream.Context()) {
				extra[key] = val
			}

			if err != nil {
				extra["error"] = err.Error()
				logger.Error("access", xlog.FieldExtra(extra))
				return
			}

			if c.EnableAccessLog {
				logger.Info("access", xlog.FieldExtra(extra))
			}
		}()
		return handler(srv, stream)
	}
}

func defaultUnaryServerInterceptor(logger *xlog.Logger, c *Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var beg = time.Now()
		var event = "normal"
		defer func() {
			if c.SlowQueryThresholdInMilli > 0 {
				if int64(time.Since(beg))/1e6 > c.SlowQueryThresholdInMilli {
					event = "slow"
				}
			}
			extra := map[string]interface{}{
				"type":   "unary",
				"method": info.FullMethod,
				"cost":   time.Since(beg),
				"event":  event,
			}
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}

				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				extra["stack"] = stack
				event = "recover"
			}

			for key, val := range getPeer(ctx) {
				extra[key] = val
			}

			if err != nil {
				extra["errors"] = err.Error()
				logger.Error("access", xlog.FieldExtra(extra))
				return
			}

			if c.EnableAccessLog {
				logger.Info("access", xlog.FieldExtra(extra))
			}
		}()
		return handler(ctx, req)
	}
}

func getClientIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("[getClinetIP] invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("[getClientIP] peer.Addr is nil")
	}
	addSlice := strings.Split(pr.Addr.String(), ":")
	return addSlice[0], nil
}

func getPeer(ctx context.Context) map[string]string {
	var peerMeta = make(map[string]string)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val, ok := md["aid"]; ok {
			peerMeta["aid"] = strings.Join(val, ";")
		}
		var clientIP string
		if val, ok := md["client-ip"]; ok {
			clientIP = strings.Join(val, ";")
		} else {
			ip, err := getClientIP(ctx)
			if err == nil {
				clientIP = ip
			}
		}
		peerMeta["clientIP"] = clientIP
		if val, ok := md["client-host"]; ok {
			peerMeta["host"] = strings.Join(val, ";")
		}
	}
	return peerMeta
}

func NewTraceUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := xtrace.NewTracer(trace.SpanKindServer)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (reply interface{}, err error) {
		var remote string
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			md = md.Copy()
		} else {
			md = metadata.MD{}
		}
		operation, mAttrs := xtrace.ParseFullMethod(info.FullMethod)
		attrs := []attribute.KeyValue{
			semconv.RPCServiceKey.String("xgrpc"),
		}
		attrs = append(attrs, mAttrs...)
		if p, ok := peer.FromContext(ctx); ok {
			remote = p.Addr.String()
		}
		attrs = append(attrs, xtrace.PeerAttr(remote)...)
		ctx, span := tracer.Start(ctx, operation, xtrace.MetadataReaderWriter(md), trace.WithAttributes(attrs...))
		defer func() {
			if err != nil {
				span.RecordError(err)
				s, ok := status.FromError(err)
				if ok {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(s.Code())))
				} else {
					span.SetStatus(codes.Error, err.Error())
				}
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()

		return handler(ctx, req)
	}
}

func NewTraceStreamServerInterceptor() grpc.StreamServerInterceptor {
	tracer := xtrace.NewTracer(trace.SpanKindServer)
	attrs := []attribute.KeyValue{
		semconv.RPCServiceKey.String("xgrpc"),
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var remote string
		md, ok := metadata.FromIncomingContext(ss.Context())
		if ok {
			md = md.Copy()
		} else {
			md = metadata.MD{}
		}

		operation, mAttrs := xtrace.ParseFullMethod(info.FullMethod)

		attrs = append(attrs, mAttrs...)
		if p, ok := peer.FromContext(ss.Context()); ok {
			remote = p.Addr.String()
		}
		attrs = append(attrs, xtrace.PeerAttr(remote)...)

		ctx, span := tracer.Start(ss.Context(), operation, xtrace.MetadataReaderWriter(md), trace.WithAttributes(attrs...))
		defer span.End()

		return handler(srv, contextedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

func metricUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	resp, err := handler(ctx, req)
	attrs := []attribute.KeyValue{
		attribute.String("method", info.FullMethod),
	}
	if err != nil {
		xmetric.GRPCServerUnaryFault.Inc(ctx, attrs...)
	}
	xmetric.GRPCServerUnaryDuration.Record(ctx, time.Since(startTime), attrs...)
	return resp, err
}

func metricStreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	startTime := time.Now()
	ctx := context.Background()
	err := handler(srv, ss)
	attrs := []attribute.KeyValue{
		attribute.String("method", info.FullMethod),
	}
	if err != nil {
		xmetric.GRPCServerStreamFault.Inc(ctx, attrs...)
	}
	xmetric.GRPCServerUnaryDuration.Record(ctx, time.Since(startTime), attrs...)
	return err
}

type contextedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context ...
func (css contextedServerStream) Context() context.Context {
	return css.ctx
}
