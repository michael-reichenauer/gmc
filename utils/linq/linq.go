package linq

import "github.com/samber/lo"

// Find search an element in a slice based on a predicate. It returns element and true if element was found.
func Find[V any](collection []V, predicate func(V) bool) (V, bool) {
	return lo.Find(collection, predicate)
}

func Contains[V any](collection []V, predicate func(V) bool) bool {
	_, ok := Find(collection, predicate)
	return ok
}

// Filter iterates over elements of collection, returning an array of all elements predicate returns truthy for.
func Filter[V any](collection []V, predicate func(V) bool) []V {
	return lo.Filter(collection, func(v V, _ int) bool { return predicate(v) })
}

// Map manipulates a slice and transforms it to a slice of another type.
func Map[V any, R any](collection []V, mapFunc func(V) R) []R {
	return lo.Map(collection, func(v V, _ int) R { return mapFunc(v) })
}

// FilterMap returns a slice which obtained after both filtering and mapping using the given callback function.
// The callback function should return two values:
//   - the result of the mapping operation and
//   - whether the result element should be included or not.
func FilterMap[V any, R any](collection []V, predicate func(V) bool, mapFunc func(V) R) []R {
	return lo.FilterMap(collection, func(v V, _ int) (R, bool) {
		var r R
		if !predicate(v) {
			return r, false
		}
		return mapFunc(v), true
	})
}
