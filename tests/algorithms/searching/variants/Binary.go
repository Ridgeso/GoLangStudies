package main

import (
	"fmt"
	"sort"
)

func main() {
	const N = 100_000
	const queries = 5_000_000

	data := make([]int, N)
	for i := 0; i < N; i++ {
		data[i] = i * 2
	}

	hits := 0
	for i := 0; i < queries; i++ {
		key := (i * 7) % (N * 2)
		idx := sort.SearchInts(data, key)
		if idx < N && data[idx] == key {
			hits++
		}
	}
	fmt.Println(hits)
}
