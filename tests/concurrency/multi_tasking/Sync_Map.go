package main

import (
	"fmt"
	"sync"
)

func main() {
	const N = 500_000
	const readers = 8

	var sm sync.Map
	for i := 0; i < N; i++ {
		sm.Store(i, i*2)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	hits := int64(0)

	for r := 0; r < readers; r++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			local := int64(0)
			for i := id; i < N; i += readers {
				if v, ok := sm.Load(i); ok {
					local += int64(v.(int))
				}
			}
			mu.Lock()
			hits += local
			mu.Unlock()
		}(r)
	}
	wg.Wait()
	fmt.Println(hits)
}
