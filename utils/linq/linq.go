package linq

import "github.com/samber/lo"

// Find search an element in a slice based on a predicate. It returns element and true if element was found.
func Find[V any](collection []V, predicate func(V) bool) (V, bool) {
	return lo.Find(collection, predicate)
}

// Contains returns true if an element is present in a collection.
func Contains[V comparable](collection []V, element V) bool {
	return lo.Contains(collection, element)
}

// ContainsBy returns true if predicate function return true.
func ContainsBy[V any](collection []V, predicate func(V) bool) bool {
	return lo.ContainsBy(collection, predicate)
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
