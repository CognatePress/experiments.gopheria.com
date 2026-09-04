// Command spread runs the benchnoise A/B benchmark as a sequence of independent
// `go test -count=1` invocations and reports the distribution of the deltas
// those single runs produced. Two of the three implementations are identical,
// so every non-zero delta it prints for that pair is measurement noise.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type key struct{ work, impl string }

func main() {
	runs := flag.Int("n", 50, "number of independent `go test` invocations")
	pkg := flag.String("pkg", "./benchnoise", "package to benchmark")
	benchtime := flag.String("benchtime", "1s", "value passed to -benchtime")
	orderFlag := flag.String("order", "forward", "implementation order passed to the test binary")
	warmup := flag.Bool("warmup", false, "run a discarded benchmark before the measured ones")
	verbose := flag.Bool("v", false, "print every run's deltas")
	flag.Parse()

	if *runs < 2 {
		log.Fatal("spread: -n must be at least 2")
	}

	start := time.Now()
	samples := make([]map[key]float64, 0, *runs)
	for i := 0; i < *runs; i++ {
		m, err := runOnce(*pkg, *benchtime, *orderFlag, *warmup)
		if err != nil {
			log.Fatalf("spread: run %d: %v", i+1, err)
		}
		samples = append(samples, m)
		if *verbose {
			fmt.Printf("run %2d  sum beta %+6.2f%%  sum plus5 %+6.2f%%  alloc beta %+6.2f%%  alloc plus5 %+6.2f%%\n",
				i+1,
				delta(m, "sum", "beta"), delta(m, "sum", "plus5"),
				delta(m, "alloc", "beta"), delta(m, "alloc", "plus5"))
		}
	}

	label := "-bench=AB"
	if *warmup {
		label = "-bench=Warmup|AB"
	}
	fmt.Printf("%d independent runs of `go test %s -count=1 -benchtime=%s -order=%s`, %s total\n\n",
		*runs, label, *benchtime, *orderFlag, time.Since(start).Round(time.Second))

	fmt.Printf("%-9s %-6s %7s %8s %8s %8s %8s %10s %8s\n",
		"workload", "impl", "truth", "min", "median", "max", "spread", "wrongsign", "|d|>5%")
	for _, work := range []string{"sum", "alloc"} {
		for _, impl := range []string{"beta", "plus5"} {
			truth := 0.0
			if impl == "plus5" {
				truth = 5.0
			}
			ds := make([]float64, 0, len(samples))
			for _, m := range samples {
				ds = append(ds, delta(m, work, impl))
			}
			sort.Float64s(ds)
			lo, hi := ds[0], ds[len(ds)-1]
			wrong, big := 0, 0
			for _, d := range ds {
				if d < 0 {
					wrong++
				}
				if d > 5 || d < -5 {
					big++
				}
			}
			fmt.Printf("%-9s %-6s %+6.1f%% %+7.2f%% %+7.2f%% %+7.2f%% %7.2f%% %7d/%-3d %5d/%-3d\n",
				work, impl, truth, lo, median(ds), hi, hi-lo, wrong, len(ds), big, len(ds))
		}
	}
	fmt.Println()
	fmt.Println("truth     = the delta the code carries by construction")
	fmt.Println("wrongsign = runs that reported this implementation as faster than alpha")
	fmt.Println("|d|>5%    = runs whose reported delta exceeded 5% in either direction")
}

// runOnce executes one `go test` invocation and returns its ns/op per benchmark.
func runOnce(pkg, benchtime, order string, warmup bool) (map[key]float64, error) {
	bench := "-bench=AB"
	if warmup {
		bench = "-bench=Warmup|AB"
	}
	cmd := exec.Command("go", "test", "-run=^$", bench, "-benchmem",
		"-count=1", "-benchtime="+benchtime, pkg, "-order="+order)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go test: %w", err)
	}
	m := make(map[key]float64)
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "BenchmarkAB/") {
			continue
		}
		k, ok := parseName(fields[0])
		if !ok {
			continue
		}
		found := false
		for i, f := range fields {
			if f == "ns/op" && i > 0 {
				v, err := strconv.ParseFloat(fields[i-1], 64)
				if err != nil {
					return nil, fmt.Errorf("parse ns/op in %q: %w", sc.Text(), err)
				}
				m[k], found = v, true
			}
		}
		if !found {
			return nil, fmt.Errorf("no ns/op in %q", sc.Text())
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(m) != 6 {
		return nil, fmt.Errorf("expected 6 benchmark lines, got %d", len(m))
	}
	return m, nil
}

// parseName turns BenchmarkAB/work=sum/impl=alpha-12 into {sum, alpha}.
func parseName(name string) (key, bool) {
	if i := strings.LastIndex(name, "-"); i > 0 {
		name = name[:i]
	}
	var k key
	for _, part := range strings.Split(name, "/") {
		switch {
		case strings.HasPrefix(part, "work="):
			k.work = strings.TrimPrefix(part, "work=")
		case strings.HasPrefix(part, "impl="):
			k.impl = strings.TrimPrefix(part, "impl=")
		}
	}
	return k, k.work != "" && k.impl != ""
}

func delta(m map[key]float64, work, impl string) float64 {
	base := m[key{work, "alpha"}]
	return (m[key{work, impl}]/base - 1) * 100
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}
