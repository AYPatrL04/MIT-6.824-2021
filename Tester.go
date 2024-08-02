package MIT6_824_2021

import (
	"strings"
)

func main() {

}

type Pair struct {
	key   string
	value int
}

func mapper(key string, value string) []Pair {
	// key: document name
	// value: document contents
	words := strings.Fields(value)
	kvs := make([]Pair, 0)
	for _, w := range words {
		kvs = append(kvs, Pair{w, 1})
	}
	return kvs
}

func reducer(key string, values []Pair) int {
	// key: word
	// values: list of word counts
	count := 0
	for _, v := range values {
		if v.key == key {
			count += v.value
		}
	}
	return count
}
