package main

import (
	"encoding/json"
	"fmt"
)

type Item struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Value float64  `json:"value"`
	Tags  []string `json:"tags"`
}

func main() {
	const N = 200_000
	items := make([]Item, N)
	for i := range items {
		items[i] = Item{
			ID:    i,
			Name:  fmt.Sprintf("item_%d", i),
			Value: float64(i) * 1.23456,
			Tags:  []string{"alpha", "beta", "gamma"},
		}
	}
	data, _ := json.Marshal(items)
	var out []Item
	json.Unmarshal(data, &out)
	fmt.Println(len(out))
}
