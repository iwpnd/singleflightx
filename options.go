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

type ShardConfig struct {
	hashFn     NewHash
	shardCount uint64
}

type ShardConfigOptions = func(*ShardConfig) func(*ShardConfig) error

func WithShardCount(shardCount uint64) func(*ShardConfig) {
	return func(config *ShardConfig) {
		config.shardCount = shardCount
	}
}

func WithHashFn(hashFn NewHash) func(*ShardConfig) {
	return func(config *ShardConfig) {
		config.hashFn = hashFn
	}
}
