// Package singleflight provides generic helpers around golang.org/x/sync/singleflight
package singleflight

import (
	"golang.org/x/sync/singleflight"
)

// Singleflighter is anything that implements singleflight.Group.
type Singleflighter[T ~string, V any] interface {
	Do(key T, fn func() (V, error)) (V, error, bool)
	DoChan(key T, fn func() (V, error)) <-chan Result[V]
	Forget(key T)
}

// Group wraps singleflight.Group with generics.
//
// T must be a string-like type (constraint ~string) to ensure keys can be
// passed through to the underlying singleflight. V is the result type
// returned by the work function.
type Group[T ~string, V any] struct {
	group singleflight.Group
}

// Result is the typed output sent on channels returned by Group.DoChan and
// ShardedGroup.DoChan.
//
// Val is the value produced by the underlying function. Err is any error
// returned by that function. Shared reports whether this caller received a
// duplicate-suppressed (shared) result, as opposed to being the caller that
// actually executed the function.
type Result[V any] struct {
	Val    V
	Err    error
	Shared bool
}

// Do executes and deduplicates the provided function for the given key.
//
// If multiple goroutines call Do with the same key at the same time, the
// function fn will be invoked exactly once; the other callers will wait for
// that single invocation to complete and will receive the same results.
//
// It returns the function's value V, its error (if any), and a boolean
// shared indicating whether this caller received a shared result.
func (g *Group[T, V]) Do(key T, fn func() (V, error)) (v V, err error, shared bool) {
	result, err, shared := g.group.Do(string(key), func() (any, error) {
		return fn()
	})

	if result != nil {
		v, _ = result.(V) //nolint:errcheck
	}

	return v, err, shared
}

// DoChan is the channel-based variant of Do.
//
// It schedules fn to run once for the given key (deduplicating concurrent
// calls with the same key) and returns a channel that will receive exactly
// one Result[V]. The channel is buffered with capacity 1 so a receiver is
// not strictly required to be ready at completion time.
//
// As with Do, callers that join an in-flight execution receive the same
// result and Err, and the Shared field indicates whether this caller
// received a shared result.
func (g *Group[T, V]) DoChan(key T, fn func() (V, error)) <-chan Result[V] {
	ch := make(chan Result[V], 1)

	upstreamCh := g.group.DoChan(string(key), func() (any, error) {
		return fn()
	})

	go g.toResult(upstreamCh, ch)

	return ch
}

// Forget tells the group to forget about an in-flight or completed entry for key.
//
// If there is a call in flight for key, subsequent Do/DoChan calls with the
// same key will not join that call after Forget has been invoked; instead,
// they will start a new, independent execution. If there is a cached
// result (from a recently completed call), it is also cleared.
func (g *Group[T, V]) Forget(key T) {
	g.group.Forget(string(key))
}

// toResult adapts singleflight.Result into a typed Result[V].
func (g *Group[T, V]) toResult(
	sourceCh <-chan singleflight.Result,
	destCh chan<- Result[V],
) {
	sourceResult := <-sourceCh

	result := Result[V]{
		Err:    sourceResult.Err,
		Shared: sourceResult.Shared,
	}

	if sourceResult.Val != nil {
		result.Val, _ = sourceResult.Val.(V) //nolint:errcheck
	}

	destCh <- result
}
