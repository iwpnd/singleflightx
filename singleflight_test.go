package singleflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	keyA         = "key-a"
	keyB         = "key-b"
	wantValueInt = 42
	wantValueStr = "ok"
	numCallers   = 5
	sleepJoin    = 30 * time.Millisecond
	sleepHold    = 50 * time.Millisecond
)

type tcase struct {
	n    int
	name string
}

func TestGroupDo(t *testing.T) {
	var g Group[string, int]
	doDedupe(t, &g, keyA)
}

func TestGroupDoChan(t *testing.T) {
	var g Group[string, string]
	doChanDedupe(t, &g, keyB)
}

func TestGroupForget(t *testing.T) {
	var g Group[string, int]
	forgetCreatesNewExecution(t, &g, keyA)
}

func TestGroupError(t *testing.T) {
	var g Group[string, int]
	doErrorPropagates(t, &g, keyB, 0)
}

type doer[T ~string, V any] interface {
	Do(T, func() (V, error)) (V, error, bool)
	DoChan(T, func() (V, error)) <-chan Result[V]
	Forget(T)
}

func forgetCreatesNewExecution[T ~string](t *testing.T, d doer[T, int], key T) {
	t.Helper()

	start := make(chan struct{})
	done := make(chan struct{})

	var total int32
	fn1 := func() (int, error) {
		atomic.AddInt32(&total, 1)
		<-start
		time.Sleep(sleepHold)
		close(done)
		return 1, nil
	}

	// begin first call
	var wg sync.WaitGroup
	wg.Add(1)
	var v1 int
	var e1 error
	var s1 bool
	go func() {
		defer wg.Done()
		v1, e1, s1 = d.Do(key, fn1)
	}()

	// let the first register
	time.Sleep(sleepJoin)

	// forget and start a fresh, independent call
	d.Forget(key)
	fn2 := func() (int, error) {
		atomic.AddInt32(&total, 1)
		return 2, nil
	}
	v2, e2, s2 := d.Do(key, fn2)

	// release the first
	close(start)
	<-done
	wg.Wait()

	if got := atomic.LoadInt32(&total); got != 2 {
		t.Fatalf("underlying calls = %d, want 2", got)
	}
	if e1 != nil || e2 != nil {
		t.Fatalf("unexpected errors: e1=%v e2=%v", e1, e2)
	}
	if v1 != 1 || v2 != 2 {
		t.Fatalf("values = (%d,%d), want (1,2)", v1, v2)
	}
	if s1 || s2 {
		t.Fatalf("shared flags = (%v,%v), want both false", s1, s2)
	}
}

func doDedupe[T ~string](t *testing.T, d doer[T, int], key T) {
	t.Helper()

	var calls int32
	fn := func() (int, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(sleepJoin)
		return wantValueInt, nil
	}

	tests := []tcase{
		{n: 1, name: "single caller"},
		{n: numCallers, name: "multiple callers"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// reset to avoid carry-over from previous run.
			atomic.StoreInt32(&calls, 0)

			var wg sync.WaitGroup
			wg.Add(tc.n)

			vals := make([]int, tc.n)
			errs := make([]error, tc.n)
			shared := make([]bool, tc.n)

			for i := range tc.n {
				go func(i int) {
					defer wg.Done()
					v, err, s := d.Do(key, fn)
					vals[i], errs[i], shared[i] = v, err, s
				}(i)
			}
			wg.Wait()

			// exactly one call per subtest.
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("underlying calls = %d, want 1", got)
			}

			for i := range tc.n {
				if errs[i] != nil {
					t.Fatalf("errs[%d]=%v, want nil", i, errs[i])
				}
				if vals[i] != wantValueInt {
					t.Fatalf("vals[%d]=%d, want %d", i, vals[i], wantValueInt)
				}
				if tc.n > 1 && shared[i] == false {
					t.Fatalf("expected calls to be shared")
				}
				if tc.n == 1 && shared[i] == true {
					t.Fatal("expected un-shared call, but got shared")
				}
			}
		})
	}
}

func doChanDedupe[T ~string](t *testing.T, d doer[T, string], key T) {
	t.Helper()

	var calls int32
	fn := func() (string, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(sleepJoin)
		return wantValueStr, nil
	}

	tests := []tcase{
		{n: 1, name: "single caller"},
		{n: numCallers, name: "multiple callers"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// reset to avoid carry-over from previous run.
			atomic.StoreInt32(&calls, 0)

			chans := make([]<-chan Result[string], 0, tc.n)
			for i := 0; i < tc.n; i++ {
				chans = append(chans, d.DoChan(key, fn))
			}

			for i := range tc.n {
				res := <-chans[i]
				if res.Err != nil {
					t.Fatalf("res.Err[%d]=%v, want nil", i, res.Err)
				}
				if res.Val != wantValueStr {
					t.Fatalf("res.Val[%d]=%q, want %q", i, res.Val, wantValueStr)
				}
				if tc.n > 1 && res.Shared == false {
					t.Fatalf("expected calls to be shared")
				}
				if tc.n == 1 && res.Shared == true {
					t.Fatal("expected un-shared call, but got shared")
				}
			}

			// exactly one call per subtest.
			if got := atomic.LoadInt32(&calls); got != 1 {
				t.Fatalf("underlying calls = %d, want 1", got)
			}
		})
	}
}

func doErrorPropagates[T ~string, V any](t *testing.T, d doer[T, V], key T, zero V) {
	t.Helper()
	wantErr := errors.New("boom")
	fn := func() (V, error) { return zero, wantErr }

	v, err, shared := d.Do(key, fn)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err=%v, want %v", err, wantErr)
	}
	_ = v // zero is fine; type-specific equality not required here
	if shared {
		t.Fatalf("shared=%v, want false", shared)
	}
}
