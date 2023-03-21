package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/5idu/pilot/pkg/util/xstring"
	"github.com/5idu/pilot/pkg/xlog"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

type redisContextKeyType struct{}

var (
	ctxBegKey            = redisContextKeyType{}
	redisCmdAttributeKey = attribute.Key("redis.cmd")
)

type interceptor struct {
	dialHook            func(next redis.DialHook) redis.DialHook
	processHook         func(next redis.ProcessHook) redis.ProcessHook
	processPipelineHook func(next redis.ProcessPipelineHook) redis.ProcessPipelineHook
}

func (i *interceptor) DialHook(next redis.DialHook) redis.DialHook {
	return i.dialHook(next)
}
func (i *interceptor) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return i.processHook(next)
}
func (i *interceptor) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return i.processPipelineHook(next)
}

func newInterceptor(compName string, config *Config, logger *xlog.Logger) *interceptor {
	return &interceptor{
		dialHook: func(next redis.DialHook) redis.DialHook {
			return next
		},
		processHook: func(next redis.ProcessHook) redis.ProcessHook {
			return next
		},
		processPipelineHook: func(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
			return next
		},
	}
}

func (i *interceptor) setDialHook(p func(next redis.DialHook) redis.DialHook) *interceptor {
	i.dialHook = p
	return i
}

func (i *interceptor) setProcessHook(p func(next redis.ProcessHook) redis.ProcessHook) *interceptor {
	i.processHook = p
	return i
}

func (i *interceptor) setProcessPipelineHook(p func(next redis.ProcessPipelineHook) redis.ProcessPipelineHook) *interceptor {
	i.processPipelineHook = p
	return i
}

func debugInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {
	return newInterceptor(compName, config, logger).
		setProcessHook(func(next redis.ProcessHook) redis.ProcessHook {
			return func(ctx context.Context, cmd redis.Cmder) error {
				start := time.Now()
				next(ctx, cmd)
				cost := time.Since(start)
				err := cmd.Err()
				fmt.Println(xstring.CallerName(6))
				fmt.Printf("[redis ] %s (%s) :\n", addr, cost) // nolint
				if err != nil {
					fmt.Printf("# %s %+v, ERR=(%s)\n\n", cmd.Name(), cmd.Args(), err.Error())
				} else {
					fmt.Printf("# %s %+v: %s\n\n", cmd.Name(), cmd.Args(), response(cmd))
				}
				return nil
			}
		}).
		setProcessPipelineHook(func(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
			return func(ctx context.Context, cmds []redis.Cmder) error {
				start := time.Now()
				next(ctx, cmds)
				cost := time.Since(start)
				fmt.Println(xstring.CallerName(8))
				fmt.Printf("[redis pipeline] %s (%s) :\n", addr, cost) // nolint
				for _, cmd := range cmds {
					err := cmd.Err()
					if err != nil {
						fmt.Printf("* %s %+v, ERR=<%s>\n", cmd.Name(), cmd.Args(), err.Error())
					} else {
						fmt.Printf("* %s %+v: %s\n", cmd.Name(), cmd.Args(), response(cmd))
					}
				}
				fmt.Print("  \n") // nolint
				return nil
			}
		})
}

func accessInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {
	return newInterceptor(compName, config, logger).
		setProcessHook(func(next redis.ProcessHook) redis.ProcessHook {
			return func(ctx context.Context, cmd redis.Cmder) error {
				start := time.Now()
				next(ctx, cmd)
				cost := time.Since(start)
				err := cmd.Err()
				// error
				extra := map[string]interface{}{
					"key":  compName,
					"name": cmd.Name(),
					"addr": addr,
					"req":  cmd.Args(),
					"cost": cost,
				}
				if err != nil {
					extra["error"] = err.Error()
					if errors.Is(err, redis.Nil) {
						logger.Warn("access", xlog.FieldExtra(extra))
						return nil
					}
					logger.Error("access", xlog.FieldExtra(extra))
					return nil
				}
				// extra["res"] = response(cmd)
				logger.Info("access", xlog.FieldExtra(extra))

				return nil
			}
		}).
		setProcessPipelineHook(func(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
			return func(ctx context.Context, cmds []redis.Cmder) error {
				start := time.Now()
				next(ctx, cmds)
				cost := time.Since(start)
				for _, cmd := range cmds {
					var err = cmd.Err()
					extra := map[string]interface{}{
						"key":  compName,
						"type": "pipeline",
						"name": cmd.Name(),
						"req":  cmd.Args(),
						"cost": cost,
					}

					// error
					if err != nil {
						extra["error"] = err.Error()
						if errors.Is(err, redis.Nil) {
							logger.Warn("access", xlog.FieldExtra(extra))
							continue
						}
						logger.Error("access", xlog.FieldExtra(extra))
						continue
					}
					// extra["res"] = response(cmd)
					logger.Info("access", xlog.FieldExtra(extra))

					continue
				}
				return nil
			}
		})
}

/*
func traceInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {
	tracer := xtrace.NewTracer(trace.SpanKindClient)
	attrs := []attribute.KeyValue{
		semconv.NetHostPortKey.String(addr),
		semconv.DBRedisDBIndexKey.Int(config.DB),
		semconv.DBSystemRedis,
	}

	return newInterceptor(compName, config, logger).
		setBeforeProcess(func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			ctx, span := tracer.Start(ctx, cmd.FullName(), nil, trace.WithAttributes(attrs...))
			span.SetAttributes(
				redisCmdAttributeKey.String(cmd.Name()),
				semconv.DBStatementKey.String(cast.ToString(cmd.Args())),
			)
			return ctx, nil
		}).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			span := trace.SpanFromContext(ctx)
			if err := cmd.Err(); err != nil && err != redis.Nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "ok")
			}

			span.End()
			return nil
		}).
		setBeforeProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
			ctx, span := tracer.Start(ctx, "pipeline", nil, trace.WithAttributes(attrs...))
			span.SetAttributes(
				redisCmdAttributeKey.String(getCmdsName(cmds)),
			)
			return ctx, nil
		}).
		setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
			span := trace.SpanFromContext(ctx)
			for _, cmd := range cmds {
				if err := cmd.Err(); err != nil && err != redis.Nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					span.End()
					return nil
				}
			}
			span.SetStatus(codes.Ok, "ok")
			span.End()
			return nil
		})
}

func metricInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {
	return newInterceptor(compName, config, logger).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			err := cmd.Err()
			attrs := []attribute.KeyValue{
				attribute.String("cmd", strings.ToUpper(cmd.Name())),
				attribute.String("addr", addr),
			}
			if err != nil {
				if errors.Is(err, redis.Nil) {
					attrs = append(attrs, attribute.String("code", "Empty"))
					xmetric.RedisCmdFault.Inc(ctx, attrs...)
				}
				attrs = append(attrs, attribute.String("code", "Error"))
				xmetric.RedisCmdFault.Inc(ctx, attrs...)
			}
			xmetric.RedisCmdDuration.Record(ctx, cost, attrs...)
			return nil
		}).setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
		cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
		for _, cmd := range cmds {
			attrs := []attribute.KeyValue{
				attribute.String("addr", addr),
				attribute.String("cmd", strings.ToUpper(cmd.Name())),
			}
			if err := cmd.Err(); err != nil {
				if errors.Is(err, redis.Nil) {
					attrs = append(attrs, attribute.String("code", "Empty"))
					xmetric.RedisCmdFault.Inc(ctx, attrs...)
				}
				attrs = append(attrs, attribute.String("code", "Error"))
				xmetric.RedisCmdFault.Inc(ctx, attrs...)
			}
		}
		attrs := []attribute.KeyValue{
			attribute.String("addr", addr),
			attribute.String("cmd", strings.ToUpper(getCmdsName(cmds))),
		}
		xmetric.RedisCmdDuration.Record(ctx, cost, attrs...)
		return nil
	})
}

func metricPoolStatsInterceptor(ins *Client) error {
	if ins.master != nil {
		cb, err := xmetric.RedisPoolCountStats.Observe(func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(xmetric.RedisPoolCountStats.Int64ObservableCounter, int64(atomic.LoadUint32(&ins.master.PoolStats().Hits)),
				attribute.String("role", "master"),
				attribute.String("type", "hits"),
				attribute.String("addr", ins.master.Options().Addr),
			)
			o.ObserveInt64(xmetric.RedisPoolCountStats.Int64ObservableCounter, int64(atomic.LoadUint32(&ins.master.PoolStats().Misses)),
				attribute.String("role", "master"),
				attribute.String("type", "misses"),
				attribute.String("addr", ins.master.Options().Addr),
			)
			o.ObserveInt64(xmetric.RedisPoolCountStats.Int64ObservableCounter, int64(atomic.LoadUint32(&ins.master.PoolStats().Timeouts)),
				attribute.String("role", "master"),
				attribute.String("type", "timeouts"),
				attribute.String("addr", ins.master.Options().Addr),
			)
			o.ObserveInt64(xmetric.RedisPoolCountStats.Int64ObservableCounter, int64(atomic.LoadUint32(&ins.master.PoolStats().StaleConns)),
				attribute.String("role", "master"),
				attribute.String("type", "stale_conn"),
				attribute.String("addr", ins.master.Options().Addr),
			)
			return nil
		})
		if err != nil {
			return errors.WithMessage(err, "register master metric callback")
		}
		ins.metricCallbacks = append(ins.metricCallbacks, cb)

		cb, err = xmetric.RedisPoolConnStats.Observe(func(ctx context.Context, o metric.Observer) error {
			o.ObserveInt64(xmetric.RedisPoolConnStats.Int64ObservableUpDownCounter, int64(atomic.LoadUint32(&ins.master.PoolStats().TotalConns)),
				attribute.String("role", "master"),
				attribute.String("type", "total"),
				attribute.String("addr", ins.master.Options().Addr),
			)
			o.ObserveInt64(xmetric.RedisPoolConnStats.Int64ObservableUpDownCounter, int64(atomic.LoadUint32(&ins.master.PoolStats().IdleConns)),
				attribute.String("role", "master"),
				attribute.String("type", "idle"),
				attribute.String("addr", ins.master.Options().Addr),
			)
			return nil
		})
		if err != nil {
			return errors.WithMessage(err, "register master metric callback")
		}
		ins.metricCallbacks = append(ins.metricCallbacks, cb)
	}
	return nil
}
*/

func response(cmd redis.Cmder) string {
	switch recv := cmd.(type) {
	case *redis.Cmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.StringCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.StatusCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.IntCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.DurationCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.BoolCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.CommandsInfoCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.StringSliceCmd:
		return fmt.Sprintf("%v", recv.Val())
	default:
		return ""
	}
}

func getCmdsName(cmds []redis.Cmder) string {
	cmdNameMap := map[string]bool{}
	cmdName := []string{}
	for _, cmd := range cmds {
		if !cmdNameMap[cmd.Name()] {
			cmdName = append(cmdName, cmd.Name())
			cmdNameMap[cmd.Name()] = true
		}
	}
	return strings.Join(cmdName, "_")
}
