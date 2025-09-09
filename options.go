package singleflight

import (
	"hash"
	"hash/fnv"
)

const (
	// DefaultShardCount defines the default number of shards used
	// when no custom shard count is provided.
	DefaultShardCount = 2
)

// NewHash is a function type that returns a new hash.Hash64.
// It allows for customizing the hash function used in sharding.
type NewHash = func() hash.Hash64

// newHash returns a new 64-bit FNV-1a hash function.
// This is the default hashing function used for sharding.
func newHash() hash.Hash64 {
	return fnv.New64a()
}

// ShardConfig configures sharding behavior for singleflight groups.
// It determines the hash function to use and the number of shards
// across which requests will be distributed.
type ShardConfig struct {
	hashFn     NewHash
	shardCount uint64
}

// ShardConfigOption defines a functional option for configuring ShardConfig.
// Options can be passed to customize shard count or hashing function.
type ShardConfigOption = func(*ShardConfig)

// WithShardCount returns a ShardConfigOption that sets the shard count.
// By default, the shard count is set to DefaultShardCount.
func WithShardCount(shardCount uint64) ShardConfigOption {
	return func(config *ShardConfig) {
		config.shardCount = shardCount
	}
}

// WithHashFn returns a ShardConfigOption that sets a custom hash function
// for computing shard indices. By default, fnv.New64a is used.
func WithHashFn(hashFn NewHash) ShardConfigOption {
	return func(config *ShardConfig) {
		config.hashFn = hashFn
	}
}
