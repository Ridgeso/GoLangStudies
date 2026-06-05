package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	const tasks = 500_000

	var result atomic.Int64
	var wg sync.WaitGroup
	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func(j int) {
			defer wg.Done()
			result.Add(int64(j % 7))
		}(i)
	}
	wg.Wait()
	fmt.Println(result.Load())
}
