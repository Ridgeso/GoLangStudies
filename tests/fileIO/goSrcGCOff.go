package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"runtime/debug"
)

func main() {
    debug.SetGCPercent(-1)
	path := "tmp/bench_io_go.txt"

	f, _ := os.Create(path)
	w := bufio.NewWriterSize(f, 1<<20)
	for i := 0; i < 1_000_000; i++ {
		w.WriteString(strconv.Itoa(i))
		w.WriteByte('\n')
	}
	w.Flush()
	f.Close()

	f2, _ := os.Open(path)
	scanner := bufio.NewScanner(f2)
	count := 0
	for scanner.Scan() {
		count++
	}
	f2.Close()
	os.Remove(path)
	fmt.Println(count)
}
