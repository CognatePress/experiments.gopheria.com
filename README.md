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
| `gomaxprocs/` | [GOMAXPROCS is not your thread count](../gopheria.com/src/content/posts/gomaxprocs-is-not-thread-count.mdx) |

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

# GOMAXPROCS: the thread count it does not cap, then the quota that sets it
go build -o /tmp/gmp ./gomaxprocs
for p in 1 2 4 12; do GOMAXPROCS=$p /tmp/gmp threads 200; done
GOMAXPROCS=2 /tmp/gmp maxthreads 20 200          # crashes: thread exhaustion
go test -run=^$ -bench=Mix -benchmem -count=10 \
  -cpu=1,2,3,4,6,8,10,12,14,16,20,24 ./gomaxprocs | benchstat -

# The container half needs Linux, so the same source is cross-compiled and run
# under Docker. --cpus is the cgroup CPU quota the runtime reads.
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /tmp/gmp-linux ./gomaxprocs
for q in 1 2 2.5 0.5; do
  docker run --rm --cpus=$q -v /tmp/gmp-linux:/gmp:ro alpine:3.22 /gmp report
done
docker run --rm --cpus=2 -v /tmp/gmp-linux:/gmp:ro alpine:3.22 /gmp work 200000 64
docker run --rm --cpus=2 -e GODEBUG=containermaxprocs=0 \
  -v /tmp/gmp-linux:/gmp:ro alpine:3.22 /gmp work 200000 64
```

`benchstat` comes from `go install golang.org/x/perf/cmd/benchstat@latest`.

Numbers will differ on your machine. The *shapes* — 203 threads against 4, a
`does not escape` verdict next to 104 KiB per operation, zero misses after one
GC and total loss after two, a thread count that does not move when `GOMAXPROCS`
moves twelvefold — are what the articles claim, and those should reproduce
anywhere.

The container runs are the one exception to the machine line above: they are the
same source cross-compiled to `linux/arm64` and run under Docker Desktop 4.88.1
(engine 29.7.2) on that same host, because the cgroup quota the runtime reads
does not exist on darwin.
