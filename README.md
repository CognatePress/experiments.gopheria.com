# Experiments

The harnesses behind the measurements published on Gopheria. Every benchmark or
behaviour claim in an article was produced by code in here, so a reader — or a
future version of the author — can rerun it instead of trusting the number.

Measured on `go1.27.0 darwin/arm64`, Apple M4 Pro, 12 cores, macOS 26.6.2.

| Directory | Article |
|---|---|
| `threads/` | [What the Go scheduler does when your goroutine blocks](../gopheria.com/src/content/posts/go-scheduler-blocking-goroutine.mdx) |
| `escape/` | [Escape analysis is not a rule of thumb](../gopheria.com/src/content/posts/escape-analysis-is-not-a-rule-of-thumb.mdx) |
| `pool/`, `poolbench/` | [sync.Pool survives one GC, not two](../gopheria.com/src/content/posts/sync-pool-under-gc-pressure.mdx) |

## Running them

```bash
cd experiments

# Scheduler: OS threads consumed by 200 blocked goroutines, three ways
go build -o /tmp/threadslab ./threads
GOMAXPROCS=2 /tmp/threadslab syscall    # blocking read(2) on a raw pipe
GOMAXPROCS=2 /tmp/threadslab netpoll    # read on an idle TCP connection
GOMAXPROCS=2 /tmp/threadslab fast       # a tight loop of fast write(2) calls
GOMAXPROCS=2 GODEBUG=schedtrace=1000 /tmp/threadslab syscall

# Escape analysis: the compiler's verdict, then what actually happened
go build -gcflags='-m -l' ./escape
go test -run=^$ -bench=. -benchmem -count=10 ./escape | benchstat -

# sync.Pool: eviction schedule, then cost under GC pressure
go run ./pool
go test -run=^$ -bench=. -benchmem -count=10 ./poolbench | benchstat -
```

`benchstat` comes from `go install golang.org/x/perf/cmd/benchstat@latest`.

Numbers will differ on your machine. The *shapes* — 203 threads against 4, a
`does not escape` verdict next to 104 KiB per operation, zero misses after one
GC and total loss after two — are what the articles claim, and those should
reproduce anywhere.
