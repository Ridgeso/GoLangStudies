package main

import (
	"fmt"
	"strings"
)

func main() {
	var sb strings.Builder
	for i := 0; i < 500_000; i++ {
		sb.WriteString(fmt.Sprintf("word%d ", i))
	}
	fmt.Println(len(sb.String()))
}
