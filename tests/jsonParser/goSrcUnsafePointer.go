package main
//unsafe.Pointer oraz wyłączony GC
import (
	json "github.com/goccy/go-json" //korzysta z unsafe.Pointer
	"fmt"
	"strconv"
	"runtime/debug"
)

type Item struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Value float64  `json:"value"`
	Tags  []string `json:"tags"`
}

func main() {
    debug.SetGCPercent(-1)
    var tags = []string{"alpha", "beta", "gamma"}
	const N = 200_000
	items := make([]Item, N)
	for i := range items {
		items[i] = Item{
			ID:    i,
			Name:  "item_" + strconv.Itoa(i),
			Value: float64(i) * 1.23456,
			Tags:  tags,
		}
	}
	data, _ := json.Marshal(items)
	out := make([]Item, 0, N)
	json.Unmarshal(data, &out)
	fmt.Println(len(out))
}
