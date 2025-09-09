// SPDX-License-Identifier: MPL-2.0
// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/
//
// Portions adapted from github.com/tarndt/shardedsingleflight (MPL-2.0).
package singleflight

// ShardedGroup distributes singleflight coordination across multiple shards
// to reduce lock contention for workloads with many distinct keys.
//
// The shard index is derived by hashing the key via newHash() and taking
// modulo shardCount. By default, NewShardedGroup constructs shardCount
// groups using DefaultShardCount and the package's newHash implementation.
type ShardedGroup[T ~string, V any] struct {
	hashFn NewHash
	shards []Group[T, V]

	shardCount uint64
}

// NewShardedGroup constructs a ShardedGroup that uses DefaultShardCount
// shards and the package's newHash function to map keys to shards.
func NewShardedGroup[T ~string, V any](opts ...ShardConfigOption) *ShardedGroup[T, V] {
	config := &ShardConfig{
		hashFn:     newHash,
		shardCount: DefaultShardCount,
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.shardCount < 2 {
		config.shardCount = 2
	}

	s := &ShardedGroup[T, V]{
		hashFn:     config.hashFn,
		shardCount: config.shardCount,
	}

	s.shards = make([]Group[T, V], s.shardCount)

	return s
}

// Do executes and deduplicates the function on the shard determined by key.
//
// Behavior matches Group.Do, but sharding reduces contention between
// unrelated keys under high concurrency.
func (sg *ShardedGroup[T, V]) Do(
	key T, fn func() (V, error),
) (v V, err error, shared bool) {
	return sg.shards[sg.shardIndex(key)].Do(key, fn)
}

// DoChan is the channel-based variant of Do for the sharded group.
//
// Behavior matches Group.DoChan, scoped to the shard determined by key.
func (sg *ShardedGroup[T, V]) DoChan(
	key T, fn func() (V, error),
) <-chan Result[V] {
	return sg.shards[sg.shardIndex(key)].DoChan(key, fn)
}

// Forget clears any in-flight or recently completed state for key on its shard.
//
// After Forget, a subsequent call with the same key will not join an
// in-flight execution started before Forget; it will start a new one.
func (sg *ShardedGroup[T, V]) Forget(key T) {
	sg.shards[sg.shardIndex(key)].Forget(key)
}

// shardIndex returns the shard index for key using the configured hash function.
//
// The hash is computed over the UTF-8 bytes of the key string, and the
// result is reduced modulo shardCount.
func (sg *ShardedGroup[T, V]) shardIndex(key T) uint64 {
	hasher := sg.hashFn()
	hasher.Write([]byte(key))

	return hasher.Sum64() % sg.shardCount
}
