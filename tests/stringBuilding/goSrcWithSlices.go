package main

import (
	"fmt"
	"strconv"
	"runtime/debug"
)

func main() {
    debug.SetGCPercent(-1)
	buf := make([]byte, 0, 8_000_000) //prealokacja
	for i := 0; i < 500_000; i++ {
		buf = append(buf, "word"...)
        buf = strconv.AppendInt(buf, int64(i), 10)
        buf = append(buf, ' ')
	}
//kopia:
    s := string(buf)
    fmt.Println(len(s))
}
