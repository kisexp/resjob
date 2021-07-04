package main

import (
	"fmt"
	"sort"
)

type ByLength []string

func (b ByLength) Len() int {
	return len(b)
}

func (b ByLength) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByLength) Less(i, j int) bool {
	return len(b[i]) < len(b[j])
}

func main() {
	fruits := []string{"apple", "banana", "kiwi", "orage"}
	sort.Sort(ByLength(fruits))
	fmt.Println(fruits)
}
