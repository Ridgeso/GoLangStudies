package main

import "fmt"

func main() {
	const N = 5_000_000
	s := make([]int, N)
	for i := 0; i < N; i++ {
		s[i] = i
	}
	fmt.Println(s[N-1])
}
