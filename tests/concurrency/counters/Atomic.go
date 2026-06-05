package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	const workers = 500
	const iters = 100_000

	var total atomic.Int64
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				total.Add(1)
			}
		}(w)
	}
	wg.Wait()
	fmt.Println(total.Load())
}
