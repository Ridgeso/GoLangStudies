package main

import (
	"fmt"
	"sync"
)

func main() {
	const workers = 500
	const iters = 100_000

	ch := make(chan int64, workers*2)
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			local := int64(0)
			for i := 0; i < iters; i++ {
				local++
			}
			ch <- local
		}(w)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	total := int64(0)
	for v := range ch {
		total += v
	}
	fmt.Println(total)
}
