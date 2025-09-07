package singleflight

import (
	"hash"
	"hash/fnv"
)

const (
	DefaultShardCount = 2
)

type NewHash = func() hash.Hash64

func newHash() hash.Hash64 {
	return fnv.New64a()
}
