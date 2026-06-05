package main

import "fmt"

type Point struct{ X, Y, Z float64 }

//go:noinline
func newPointValue(x, y, z float64) Point {
	return Point{X: x, Y: y, Z: z}
}

func main() {
	const N = 10_000_000
	sum := 0.0
	for i := 0; i < N; i++ {
		p := newPointValue(float64(i), float64(i+1), float64(i+2))
		sum += p.X + p.Y + p.Z
	}
	fmt.Println(sum)
}
