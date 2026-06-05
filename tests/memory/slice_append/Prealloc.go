package main

import "fmt"

func main() {
	const N = 5_000_000
	s := make([]int, 0, N)
	for i := 0; i < N; i++ {
		s = append(s, i)
	}
	fmt.Println(s[N-1])
}
