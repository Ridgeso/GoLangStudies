package main

import "fmt"

func main() {
	const N = 100_000
	const queries = 2_000

	data := make([]int, N)
	for i := 0; i < N; i++ {
		data[i] = i * 2
	}

	hits := 0
	for i := 0; i < queries; i++ {
		key := (i * 7) % (N * 2)
		for _, v := range data {
			if v == key {
				hits++
				break
			}
		}
	}
	fmt.Println(hits)
}
