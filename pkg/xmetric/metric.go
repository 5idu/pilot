package xmetric

var (
	// RedisCmdFault ...
	RedisCmdFault = NewInt64CounterVecOpts("redis.cmd.faults", "The number of redis cmd faults.")
	// RedisCmdDuration ...
	RedisCmdDuration = NewHistogramVec("redis.cmd.duration", "The duration of redis cmd.")
	// RedisPoolCountStats ...
	RedisPoolCountStats = NewInt64CounterObserverVecOpts("redis.pool.count", "The count stats of redis pool.")
	// RedisPoolCountStats ...
	RedisPoolConnStats = NewUpDownCounterObserverVecOpts("redis.pool.conn", "The conn stats of redis pool.")
	// HttpRequestFault ...
	HttpRequestFault = NewInt64CounterVecOpts("http.request.faults", "The number of http request faults.")
	// HttpRequestDuration ...
	HttpRequestDuration = NewHistogramVec("http.request.duration", "The duration of http request.")
	// GRPCCallFault ...
	GRPCCallFault = NewInt64CounterVecOpts("grpc.call.faults", "The number of grpc call faults.")
	// GRPCCallDuration ...
	GRPCCallDuration = NewHistogramVec("grpc.call.duration", "The duration of grpc call.")
	// RedisPoolCountStats ...
	MongoDBClientSession = NewUpDownCounterObserverVecOpts("mongodb.client.session", "The number of client current session of mongodb.")
	// EchoServerDuration ...
	EchoServerDuration = NewHistogramVec("echo.server.duration", "The duration of http echo server.")
	// GRPCServerUnaryFault ...
	GRPCServerUnaryFault = NewInt64CounterVecOpts("grpc.server.unary.faults", "The number of grpc server unary faults.")
	// GRPCServerUnaryDuration ...
	GRPCServerUnaryDuration = NewHistogramVec("grpc.server.unary.duration", "The duration of grpc server unary.")
	// GRPCServerStreamFault ...
	GRPCServerStreamFault = NewInt64CounterVecOpts("grpc.server.stream.faults", "The number of grpc server stream faults.")
	// GRPCServerStreamDuration ...
	GRPCServerStreamDuration = NewHistogramVec("grpc.server.stream.duration", "The duration of grpc server stream.")
)