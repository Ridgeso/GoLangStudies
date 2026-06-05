package main

import "fmt"

type PackedStruct struct {
	B int64 // 8
	A bool  // 1
	C bool  // 1 + 6(pad to align)
}
// sizeof = 16 bytes (vs 24 padded)

func main() {
	const N = 5_000_000
	items := make([]PackedStruct, N)
	for i := range items {
		items[i] = PackedStruct{B: int64(i), A: i%2 == 0, C: i%3 == 0}
	}
	sum := int64(0)
	for i := range items {
		sum += items[i].B
	}
	fmt.Println(sum)
}
