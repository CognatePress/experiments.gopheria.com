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
| `allocsize/` | [The 30% is 46%, and it stops dead at 80 bytes](../gopheria.com/src/content/posts/the-thirty-percent-smaller-allocation.mdx) |
| `mutexchan/` | [The mutex was 68 times faster, until I added work](../gopheria.com/src/content/posts/mutex-versus-channel-at-contention.mdx) |
| `routers/` | [A 404 from net/http costs 62 allocations](../gopheria.com/src/content/posts/http-router-dispatch-cost.mdx) |
| `gclab/` | [The GOGC I set next to GOMEMLIMIT never fired](../gopheria.com/src/content/posts/gogc-and-gomemlimit-are-different-knobs.mdx) · [Green Tea cut my GC time 63%, and raised it 39%](../gopheria.com/src/content/posts/green-tea-changed-your-gc-profile.mdx) · [Only 54 µs of that 68 ms collection was a pause](../gopheria.com/src/content/posts/reading-gctrace-without-guessing.mdx) |

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

# Four routers on one 200-route table, with the handler removed. router=stdlib-
# control is a second ServeMux built from the same routes; BenchmarkEncode is
# the ruler the dispatch numbers are read against.
go test -run=^$ -bench='Warmup|Dispatch|Encode' -benchmem -count=10 ./routers \
  | grep -v Warmup | benchstat -row /path -col /router -

# A request that matches nothing is the expensive one, and it moves along the
# method axis rather than the route axis.
go test -run=^$ -bench='Warmup|Miss' -benchmem -count=10 ./routers \
  | grep -v Warmup | benchstat -row /routes -col /router -
go test -run=^$ -bench='Warmup|MissMethods' -benchmem -count=10 ./routers \
  | grep -v Warmup | benchstat -row /methods -col /router -

# Where those allocations are. -memprofilerate=1 samples every one, so the
# timings in this run are not comparable with the table above.
go test -run=^$ -bench='Dispatch/router=stdlib/path=miss' -benchmem \
  -memprofile=/tmp/stdlib-miss.prof -memprofilerate=1 ./routers
go tool pprof -top -nodecount=6 -sample_index=alloc_objects /tmp/stdlib-miss.prof

# One steady allocator under each GC knob in turn, then both, then a limit set
# close enough to the live heap to thrash. These are program runs rather than
# benchmarks: each configuration is run five times and read as a median with
# its range, because benchstat has nothing to parse here.
go build -o /tmp/gclab ./gclab
for g in 25 50 100 200 400; do
  GOGC=$g /tmp/gclab -mode=pointer -live=134217728 -dur=6s -label="GOGC=$g"
done
for m in 192MiB 256MiB 384MiB 512MiB; do
  GOGC=off GOMEMLIMIT=$m /tmp/gclab -mode=pointer -live=134217728 -dur=6s
done
GOGC=100 GOMEMLIMIT=256MiB /tmp/gclab -mode=pointer -live=134217728 -dur=6s
GOGC=off GOMEMLIMIT=176MiB /tmp/gclab -mode=pointer -live=134217728 -dur=6s   # thrashes

# The same heap with nothing for the mark phase to follow. Pointer density is
# the only difference; object size and allocation count are equal by design.
/tmp/gclab -mode=scalar -live=134217728 -dur=6s

# Green Tea, on and off, against three heap shapes. Same source, two binaries;
# the flag is a build-time GOEXPERIMENT, so the A/B needs both.
GOEXPERIMENT=nogreenteagc go build -o /tmp/gclab-nogt ./gclab
go version -m /tmp/gclab-nogt | grep GOEXPERIMENT   # build GOEXPERIMENT=nogreenteagc
for rep in 1 2 3 4 5 6 7; do
  for mode in graph pointer scalar; do
    /tmp/gclab      -mode=$mode -live=134217728 -dur=6s
    /tmp/gclab-nogt -mode=$mode -live=134217728 -dur=6s
  done
done
```

```bash
# One gctrace line per collection, and a heap profile written straight after a
# forced one so the line's live figure has something to be checked against.
GODEBUG=gctrace=1 /tmp/gclab -mode=pointer -live=134217728 -dur=6s \
  -heapprofile=/tmp/gc-healthy.pprof
GODEBUG=gctrace=1 GOGC=400 /tmp/gclab -mode=pointer -live=134217728 -dur=6s
GODEBUG=gctrace=1 GOGC=off GOMEMLIMIT=176MiB /tmp/gclab -mode=pointer -live=134217728 -dur=6s
go tool pprof -top -nodecount=4 -sample_index=inuse_space /tmp/gc-healthy.pprof
```

```bash
# One heap allocation at each size class either side of 80 bytes, on both
# allocation paths. Every benchmark allocates a literal size: Go 1.27 picks its
# size-specialised routine from a compile-time constant, and a size read out of
# a slice defeats the thing being measured.
#
# Run the configurations in interleaved rounds and pool the samples. Comparing
# one process against another back-to-back puts every between-process
# difference into the delta — a control pair of two identical runs reported a
# 44.82% difference that way.
for round in 1 2 3 4 5; do
  go test -run=^$ -bench='Warmup|Alloc' -benchmem -count=6 ./allocsize >> on.txt
  GOEXPERIMENT=nosizespecializedmalloc \
    go test -run=^$ -bench='Warmup|Alloc' -benchmem -count=6 ./allocsize >> off.txt
  go test -run=^$ -bench='Warmup|Alloc' -benchmem -count=6 ./allocsize >> ctl.txt
done
benchstat -filter '/shape:noscan' -row /size off=off.txt on=on.txt

# The cross-toolchain leg needs GOTOOLCHAIN=local and a `go` directive the older
# toolchain accepts. With the default of auto, a go.mod requiring 1.27 makes
# go1.26.8 download go1.27.0 and re-execute under it — `go1.26.8 version` then
# prints go1.27.0 and both legs measure the same compiler.
sed -i '' 's/^go 1\.27\.0$/go 1.26.0/' go.mod
GOTOOLCHAIN=local go1.26.8 test -run=^$ -bench='Warmup|Alloc' -benchmem -count=6 ./allocsize
GOTOOLCHAIN=local go1.27.0 test -run=^$ -bench='Warmup|Alloc' -benchmem -count=6 ./allocsize
git checkout go.mod

# Which routines the compiler actually emitted, per toolchain and flag.
go tool nm <binary> | grep -c 'mallocgcSmall.*SC[0-9]'
```

The `graph` mode is the one Green Tea is aimed at: four million 32-byte nodes,
each pointing into a separate 64 MiB pool, so a page holds 256 objects that all
need marking. `pointer` and `scalar` are 768-byte objects — ten to a page — and
they land on the other side of the result.

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
