// Command gomaxprocs separates the two numbers people conflate: GOMAXPROCS,
// which is permission to run Go code, and the OS thread count, which is not
// capped by it.
//
//	go run ./gomaxprocs report            # what the runtime decided, and why
//	go run ./gomaxprocs threads 200       # threads created by N blocked syscalls
//	go run ./gomaxprocs maxthreads 20     # the ceiling that crashes instead of throttling
//	go run ./gomaxprocs work              # throughput and tail latency at this GOMAXPROCS
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func osThreads() int { return pprof.Lookup("threadcreate").Count() }

// report prints the two numbers side by side, plus the cgroup file the runtime
// reads on Linux to arrive at the default. On darwin the file does not exist,
// which is itself the answer for why the container case has to run in a VM.
func report() {
	fmt.Printf("GOOS=%s GOARCH=%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("runtime.NumCPU()   = %d\n", runtime.NumCPU())
	fmt.Printf("runtime.GOMAXPROCS = %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("GOMAXPROCS env     = %q\n", os.Getenv("GOMAXPROCS"))
	fmt.Printf("GODEBUG env        = %q\n", os.Getenv("GODEBUG"))
	fmt.Printf("threads now        = %d\n", osThreads())
	for _, p := range []string{"/sys/fs/cgroup/cpu.max", "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"} {
		if b, err := os.ReadFile(p); err == nil {
			fmt.Printf("%-18s = %q\n", p, string(b))
		}
	}
}

// blockInSyscall parks n goroutines in read(2) on the read end of a raw pipe
// nobody writes to. Raw fds from syscall.Pipe never reach the netpoller, so
// each call blocks the thread it is running on and the runtime must make a new
// one for the P.
func blockInSyscall(n int) {
	for i := 0; i < n; i++ {
		var fds [2]int
		if err := syscall.Pipe(fds[:]); err != nil {
			panic(err)
		}
		go func(fd int) {
			buf := make([]byte, 1)
			syscall.Read(fd, buf)
		}(fds[0])
	}
}

func threads(n int) {
	before := osThreads()
	blockInSyscall(n)
	time.Sleep(2 * time.Second) // sysmon retakes a P only after 20µs; give it room
	fmt.Printf("%d\t%d\t%d\t%d\n", runtime.GOMAXPROCS(0), n, before, osThreads())
}

// maxthreads shows what the runtime-wide ceiling actually does when it is hit.
func maxthreads(limit, blockers int) {
	debug.SetMaxThreads(limit)
	fmt.Printf("SetMaxThreads(%d), GOMAXPROCS=%d, blocking %d goroutines in read(2)\n",
		limit, runtime.GOMAXPROCS(0), blockers)
	blockInSyscall(blockers)
	time.Sleep(2 * time.Second)
	fmt.Printf("survived: threads=%d\n", osThreads())
}

// work is the container-side load: a fixed number of requests, each doing CPU
// work over a short-lived allocation so the GC runs, spread over a fixed
// number of concurrent workers. Concurrency is held constant across runs; only
// GOMAXPROCS changes.
func work(requests, concurrency int) {
	lat := make([]time.Duration, requests)
	jobs := make(chan int, concurrency)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	start := time.Now()
	for w := 0; w < concurrency; w++ {
		go func() {
			defer wg.Done()
			for i := range jobs {
				t0 := time.Now()
				buf := make([]byte, 4096)
				var acc uint64
				for r := 0; r < 400; r++ {
					acc = mix(acc+uint64(r), 64)
					buf[acc%4096] = byte(acc)
				}
				sinkByte = buf[acc%4096]
				lat[i] = time.Since(t0)
			}
		}()
	}
	for i := 0; i < requests; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	total := time.Since(start)

	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	var st runtime.MemStats
	runtime.ReadMemStats(&st)
	fmt.Printf("gomaxprocs=%d conc=%d requests=%d wall=%.3fs ops/sec=%.0f p50=%v p99=%v gc=%d threads=%d\n",
		runtime.GOMAXPROCS(0), concurrency, requests, total.Seconds(),
		float64(requests)/total.Seconds(),
		lat[requests*50/100].Round(time.Microsecond),
		lat[requests*99/100].Round(time.Microsecond),
		st.NumGC, osThreads())
}

var sinkByte byte

func atoi(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func arg(i, def int) int {
	if len(os.Args) > i {
		return atoi(os.Args[i], def)
	}
	return def
}

func main() {
	mode := "report"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "report":
		report()
	case "threads":
		threads(arg(2, 200))
	case "maxthreads":
		maxthreads(arg(2, 20), arg(3, 200))
	case "work":
		work(arg(2, 200000), arg(3, 64))
	default:
		fmt.Fprintln(os.Stderr, "usage: gomaxprocs report|threads|maxthreads|work")
		os.Exit(2)
	}
}
