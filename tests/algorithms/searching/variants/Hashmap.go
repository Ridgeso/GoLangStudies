package main

import "fmt"

func main() {
	const N = 100_000
	const queries = 5_000_000

	m := make(map[int]int, N)
	for i := 0; i < N; i++ {
		m[i*2] = i
	}

	hits := 0
	for i := 0; i < queries; i++ {
		key := (i * 7) % (N * 2)
		if _, ok := m[key]; ok {
			hits++
		}
	}
	fmt.Println(hits)
}
