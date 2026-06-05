package main

import (
	"fmt"
	"sync"
)

func main() {
	const workers = 500
	const iters = 100_000

	var mu sync.Mutex
	total := int64(0)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				mu.Lock()
				total++
				mu.Unlock()
			}
		}(w)
	}
	wg.Wait()
	fmt.Println(total)
}
