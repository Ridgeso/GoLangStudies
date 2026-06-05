package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

func main() {
	const tasks = 500_000
	numWorkers := runtime.GOMAXPROCS(0)

	jobs := make(chan int, tasks)
	for i := 0; i < tasks; i++ {
		jobs <- i
	}
	close(jobs)

	var result atomic.Int64
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local := int64(0)
			for j := range jobs {
				local += int64(j % 7)
			}
			result.Add(local)
		}()
	}
	wg.Wait()
	fmt.Println(result.Load())
}
