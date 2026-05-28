package main

import "fmt"

func main() {
	const N = 2_000_000
	m := make(map[int]int, N)
	for i := 0; i < N; i++ {
		m[i] = i * 2
	}
	sum := 0
	for i := 0; i < N; i++ {
		sum += m[i]
	}
	fmt.Println(sum)
}
