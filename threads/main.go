// Command threads shows what a blocked goroutine costs in OS threads,
// depending on how it blocks.
//
//	go run ./threads syscall   # 200 goroutines in a blocking read(2)
//	go run ./threads netpoll   # 200 goroutines in a network read
package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

const blockers = 200

func osThreads() int { return pprof.Lookup("threadcreate").Count() }

// blockInSyscall parks each goroutine in read(2) on the read end of a raw pipe
// that nobody writes to. Raw fds from syscall.Pipe never reach the netpoller,
// so the syscall really does block the thread it is running on.
func blockInSyscall() {
	for i := 0; i < blockers; i++ {
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

// blockInNetpoll parks each goroutine in a read on an idle TCP connection.
// os/net file descriptors are registered with the netpoller, so the goroutine
// parks and the thread goes back to running other work.
func blockInNetpoll() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c // held open, never written to
		}
	}()
	for i := 0; i < blockers; i++ {
		c, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			panic(err)
		}
		go func(c net.Conn) {
			buf := make([]byte, 1)
			c.Read(buf)
		}(c)
	}
}

// blockInFastSyscalls keeps every goroutine inside a syscall almost all the
// time — but each individual call returns immediately. sysmon only retakes a P
// whose syscall outlasted one sysmon tick (at least 20µs, and longer once
// sysmon backs its polling interval off), so these never trigger a handoff.
func blockInFastSyscalls(stop chan struct{}) {
	null, err := syscall.Open("/dev/null", syscall.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	buf := []byte{0}
	for i := 0; i < blockers; i++ {
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				syscall.Write(null, buf)
			}
		}()
	}
}

func main() {
	mode := "syscall"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	fmt.Printf("GOMAXPROCS=%d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("threads before: %d\n", osThreads())

	switch mode {
	case "syscall":
		blockInSyscall()
	case "netpoll":
		blockInNetpoll()
	case "fast":
		stop := make(chan struct{})
		defer close(stop)
		blockInFastSyscalls(stop)
	default:
		fmt.Fprintln(os.Stderr, "usage: threads syscall|netpoll|fast")
		os.Exit(2)
	}

	// Give sysmon time to retake the Ps and the runtime time to settle.
	time.Sleep(2 * time.Second)

	fmt.Printf("goroutines blocked: %d\n", blockers)
	fmt.Printf("threads after:  %d\n", osThreads())
}
