<br />
<div align="center">
  <h3 align="center">singleflightx</h3>

  <p align="center">
    singleflight.Group extension - generic and (optionally) sharded.
    <br />
    <a href="https://github.com/iwpnd/gssf/issues">Report Bug</a>
    ·
    <a href="https://github.com/iwpnd/gssf/issues">Request Feature</a>
  </p>
</div>

## About the project

This package adds generics to [`singleflight.Group`](https://pkg.go.dev/golang.org/x/sync/singleflight) by wrapping the original implementation. It also extends it with a sharded variant as per [shardedsingleflight](https://github.com/tarndt/shardedsingleflight/) that spreads the coordination across shards to reduce contention in very busy systems.

## Installation

```bash
go get github.com/iwpnd/singleflightx
```

## Quick start

### Deduplication with `Group`

```go
package main

import (
    "fmt"
    "sync"
    "time"
    sfx "github.com/iwpnd/singleflightx"
)

type key string

func main() {
    var g sfx.Group[key, int]

    fn := func() (int, error) {
        time.Sleep(50 * time.Millisecond) // pretend work
        return 42, nil
    }

    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            v, err, shared := g.Do(key("answer"), fn)
            fmt.Printf("goroutine %d => v=%d err=%v shared=%v\n", id, v, err, shared)
        }(i)
    }

    wg.Wait()
}
```

What you’ll see is exactly one underlying call runs; all callers receive the same value with `shared=true`.

### Channel variant with `DoChan`

```go
resCh := g.DoChan(key("answer"), fn)
res := <-resCh
fmt.Println(res.Val, res.Err, res.Shared)
```

This is useful when you want to compose with `select` or timers.

### Forcing a fresh execution with `Forget`

```go
g.Forget(key("answer"))
// The next Do/DoChan with the same key won’t join an in-flight call started before Forget.
```

## Deduplication with `ShardedGroup`

`ShardedGroup[T, V]` reduces lock contention by hashing keys to shards.

```go
sg := sf.NewShardedGroup[string, string]()

fetch := func() (string, error) {
    time.Sleep(25 * time.Millisecond)
    return "ok", nil
}

ch1 := sg.DoChan("/users", fetch)
ch2 := sg.DoChan("/teams", fetch)

r1, r2 := <-ch1, <-ch2
fmt.Println(r1.Val, r1.Shared, r2.Val, r2.Shared)
```

Each key maps to a shard via an internal hash, so unrelated keys don’t contend on the same mutex.

## Development

Run tests:

```bash
make test
```

Lint and format:

```bash
make lint
```

## Contributing

Issues and PRs are welcome. Please open an issue first if you plan substantial changes to the API.

## License

Distributed under MIT License. See `LICENSE.md`

## Contact

**Maintainer**: Your Name
Email: [iwpnd@posteo.de](mailto:iwpnd@posteo.de)

---

## Acknowledgments

* The original singleflight package from `golang.org/x/sync`
* Idea and README structure inspired by pragmatic repos like rip-ts
* [tarndt](https://github.com/tarndt) for his work on [shardedsingleflight](https://github.com/tarndt/shardedsingleflight)

