package xstring

import (
	"hash/fnv"
)

func Hash(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}
