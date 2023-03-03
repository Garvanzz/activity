package algorithm

import (
	"math/rand"
)

func Samples(collection []interface{}, count int) []interface{} {
	size := len(collection)

	ts := append([]interface{}{}, collection...)

	results := []interface{}{}

	for i := 0; i < size && i < count; i++ {
		copyLength := size - i

		index := rand.Intn(size - i)
		results = append(results, ts[index])

		// Removes element.
		// It is faster to swap with last element and remove it.
		ts[index] = ts[copyLength-1]
		ts = ts[:copyLength-1]
	}

	return results
}
