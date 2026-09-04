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
| `benchnoise/` | [benchstat reported p=0.000 between a function and itself](../gopheria.com/src/content/posts/benchstat-or-it-didnt-happen.mdx) |
| `stacks/` | [A 1 MiB goroutine stack costs 136µs and reports 144 B/op](../gopheria.com/src/content/posts/goroutine-stacks-grow-by-copying.mdx) |
| `mutexchan/` | [The mutex was 68 times faster, until I added work](../gopheria.com/src/content/posts/mutex-versus-channel-at-contention.mdx) |

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

# Benchmark method: two identical implementations and one that does 5% more
# work, so every reported delta has a known correct answer to be judged against.
go test -run=^$ -bench=AB -benchmem -count=10 ./benchnoise/ | benchstat -row /work -col /impl -
go test -run=^$ -bench=AB -benchmem -count=10 ./benchnoise/ -order=reverse | benchstat -row /work -col /impl -

# The distribution 30 single runs produce, with and without a discarded warm-up
go build -o /tmp/spread ./benchnoise/spread
/tmp/spread -n 30 -order=forward
/tmp/spread -n 30 -order=forward -warmup

# Goroutine stacks: where they are copied, how big each one is, and what a copy
# does to a pointer aimed at the old one.
go build -o /tmp/stacklab ./stacks/main
/tmp/stacklab growth 4096
/tmp/stacklab sizes
/tmp/stacklab rewrite 4096
go test -run=^$ -bench='Warmup|Descend' -benchmem -count=10 ./stacks/ \
  | grep -v Warmup | benchstat -row /depth -col /goroutine -

# Mutex against channel, swept from 1 to 64 goroutines, first with an empty
# critical section and then with 146ns of work in it. mutex-control is a second
# copy of the baseline, so whatever benchstat reports for it is the run's own
# error bar.
go test -run=^$ -bench='Warmup|Counter' -benchmem -count=10 ./mutexchan \
  | grep -v Warmup | benchstat -row /goroutines -col /impl -
go test -run=^$ -bench='Warmup|Counter' -benchmem -count=10 ./mutexchan -spin=128 \
  | grep -v Warmup | benchstat -row /goroutines -col /impl -
go test -run=^$ -bench=Spin -count=10 ./mutexchan | benchstat -row /iters -

# Whether that mutex ever reaches starvation mode, via its two symptoms: how
# evenly the lock was shared, and how long the longest waiter waited.
go build -o /tmp/fairness ./mutexchan/fairness
/tmp/fairness -dur=3s
/tmp/fairness -dur=3s -timed
```

`benchstat` comes from `go install golang.org/x/perf/cmd/benchstat@latest`; the
numbers published from `benchnoise/` were taken with
`golang.org/x/perf@v0.0.0-20260825160852-19be9d8e6c70`.

Numbers will differ on your machine. The *shapes* — 203 threads against 4, a
`does not escape` verdict next to 104 KiB per operation, zero misses after one
GC and total loss after two, a thread count that does not move when `GOMAXPROCS`
moves twelvefold — are what the articles claim, and those should reproduce
anywhere.

The container runs are the one exception to the machine line above: they are the
same source cross-compiled to `linux/arm64` and run under Docker Desktop 4.88.1
(engine 29.7.2) on that same host, because the cgroup quota the runtime reads
does not exist on darwin.
