package main

import "fmt"

func matMul(a, b [][]float64, n int) [][]float64 {
	c := make([][]float64, n)
	for i := range c {
		c[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for k := 0; k < n; k++ {
			aik := a[i][k]
			for j := 0; j < n; j++ {
				c[i][j] += aik * b[k][j]
			}
		}
	}
	return c
}

func main() {
	n := 400
	a := make([][]float64, n)
	b := make([][]float64, n)
	for i := 0; i < n; i++ {
		a[i] = make([]float64, n)
		b[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			a[i][j] = float64(i+j) * 0.01
			b[i][j] = float64(i-j) * 0.01
		}
	}
	c := matMul(a, b, n)
	fmt.Printf("%.6f\n", c[n-1][n-1])
}
