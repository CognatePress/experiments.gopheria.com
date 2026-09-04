// Command stacklab prints where a goroutine's stack is copied, how large each
// new one is, and what a copy does to a pointer aimed at the old one.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopheria.lab/stacks"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: stacklab growth|sizes|rewrite [depth]")
	}
	depth := 4096
	if len(os.Args) > 2 {
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("depth: %v", err)
		}
		depth = n
	}

	switch os.Args[1] {
	case "growth":
		growth(depth)
	case "sizes":
		sizes()
	case "rewrite":
		rewrite(depth)
	default:
		log.Fatalf("unknown mode %q", os.Args[1])
	}
}

func growth(depth int) {
	addrs := make([]uintptr, depth)
	done := make(chan struct{})
	go func() {
		stacks.Descend(0, addrs)
		close(done)
	}()
	<-done

	moves := stacks.Moves(addrs)
	if len(moves) == 0 {
		fmt.Println("no stack copies observed")
		return
	}
	fmt.Printf("local array %d bytes, frame stride %d bytes, depth %d\n",
		stacks.FrameBytes, moves[0].Stride, depth)
	fmt.Printf("%-6s %-9s %-12s %s\n", "copy", "at depth", "frames live", "vs previous")
	prev := 0
	for i, m := range moves {
		ratio := ""
		if prev > 0 {
			ratio = fmt.Sprintf("%.2fx", float64(m.Depth)/float64(prev))
		}
		fmt.Printf("%-6d %-9d %-12s %s\n", i+1, m.Depth, human(m.Used), ratio)
		prev = m.Depth
	}
}

func sizes() {
	fmt.Printf("%-8s %-12s %s\n", "depth", "frames", "stackinuse")
	for depth := 8; depth <= 8192; depth *= 2 {
		fmt.Printf("%-8d %-12s %s\n", depth,
			human(depth*stacks.FrameBytes), human(int(stacks.StackInuseAt(depth))))
	}
}

func rewrite(depth int) {
	done := make(chan stacks.Rewritten)
	go func() { done <- stacks.Rewrite(depth) }()
	r := <-done

	fmt.Printf("stack copies during the descent: %d\n", r.Moves)
	fmt.Printf("address before the descent:      0x%x\n", r.Before)
	fmt.Printf("pointer after the descent:       0x%x  (moved %s)\n",
		r.Pointer, human(abs(int(r.Pointer)-int(r.Before))))
	fmt.Printf("uintptr taken before it:         0x%x  (unchanged: %t)\n",
		r.Stale, r.Stale == r.Before)
	fmt.Printf("value read through the pointer:  %d\n", r.Value)
}

func human(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
