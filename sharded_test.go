package singleflight

import "testing"

func TestShardedGroupDo(t *testing.T) {
	sg := NewShardedGroup[string, int]()
	doDedupe(t, sg, keyA)
}

func TestShardedGroupDoChan(t *testing.T) {
	sg := NewShardedGroup[string, string]()
	doChanDedupe(t, sg, keyB)
}

func TestShardedGroupForget(t *testing.T) {
	sg := NewShardedGroup[string, int]()
	forgetCreatesNewExecution(t, sg, keyA)
}

func TestShardedGroupError(t *testing.T) {
	sg := NewShardedGroup[string, int]()
	doErrorPropagates(t, sg, keyB, 0)
}
