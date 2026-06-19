package main

import (
	"fmt"
	"sort"
)

func main() {
	const N = 10_000_000
	data := make([]int, N)
	for i := range data {
		data[i] = (i*1_000_003 + 7) % N
	}
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	fmt.Println(data[0], data[N-1])
}
