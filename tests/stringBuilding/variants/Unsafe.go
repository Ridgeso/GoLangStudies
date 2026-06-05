package main

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"unsafe"
)

func main() {
	debug.SetGCPercent(-1)
	buf := make([]byte, 0, 8_000_000) //prealokacja
	for i := 0; i < 500_000; i++ {
		buf = append(buf, "word"...)
		buf = strconv.AppendInt(buf, int64(i), 10)
		buf = append(buf, ' ')
	}
	s := unsafe.String(
		unsafe.SliceData(buf),
		len(buf),
	)
	fmt.Println(len(s))
}
