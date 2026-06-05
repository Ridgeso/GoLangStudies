package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder
	sb.Grow(8_000_000)
	for i := 0; i < 500_000; i++ {
		sb.WriteString("word")
        sb.WriteString(strconv.Itoa(i))
        sb.WriteByte(' ')
	}
	fmt.Println(len(sb.String()))
}
