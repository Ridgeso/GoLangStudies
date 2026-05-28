package main

import (
	"fmt"
	"sort"
)

func main() {
	n := 10_000_000
	arr := make([]int, n)
	for i := range arr {
		arr[i] = (i*1_000_003 + 7) % n
	}
	sort.Ints(arr)
	fmt.Println(arr[0], arr[n-1])
}
