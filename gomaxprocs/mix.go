package main

// mix is splitmix64, chosen because it touches no memory: the sweep below
// measures the scheduler's willingness to run Go code in parallel, not the
// memory subsystem's willingness to keep up.
func mix(seed uint64, rounds int) uint64 {
	x := seed
	for i := 0; i < rounds; i++ {
		x += 0x9e3779b97f4a7c15
		z := x
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		x ^= z >> 31
	}
	return x
}
