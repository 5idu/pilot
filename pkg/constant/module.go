package constant

type Module int

const (
	ModuleInvalid = Module(iota)

	ModuleClientGrpc
	ModuleClientRedis
	ModuleClientEtcd

	ModuleRegistryEtcd

	ModuleStoreMongoDB
)
