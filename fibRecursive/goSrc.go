package main

import "fmt"

func fibRec(n int) int64 {
	if n <= 1 {
		return int64(n)
	}
	return fibRec(n-1) + fibRec(n-2)
}

func main() {
	fmt.Println(fibRec(42))
}
