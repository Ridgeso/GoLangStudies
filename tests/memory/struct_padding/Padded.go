package main

import "fmt"

type PaddedStruct struct {
	A bool  // 1 + 7(pad)
	B int64 // 8
	C bool  // 1 + 7(pad)
}
// sizeof = 24 bytes

func main() {
	const N = 5_000_000
	items := make([]PaddedStruct, N)
	for i := range items {
		items[i] = PaddedStruct{A: i%2 == 0, B: int64(i), C: i%3 == 0}
	}
	sum := int64(0)
	for i := range items {
		sum += items[i].B
	}
	fmt.Println(sum)
}
