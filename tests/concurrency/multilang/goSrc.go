package main

import (
	"fmt"
	"sync"
)

func worker(id int, wg *sync.WaitGroup, ch chan<- int) {
	defer wg.Done()
	sum := 0
	for i := 0; i < 10_000; i++ {
		sum += i * id
	}
	ch <- sum
}

func main() {
	const numWorkers = 1000
	ch := make(chan int, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, ch)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	total := int64(0)
	for v := range ch {
		total += int64(v)
	}
	fmt.Println(total)
}
