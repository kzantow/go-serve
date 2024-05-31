package serve

import (
	"reflect"
	"slices"

	"golang.org/x/exp/maps"
)

func IsErrorType(out reflect.Type) bool {
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	return out.Implements(errorInterface)
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetOrPanic[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func SortedMapValues[T any](m map[string]T) []T {
	var out []T
	sorted := maps.Keys(m)
	slices.Sort(sorted)
	for _, k := range sorted {
		out = append(out, m[k])
	}
	return out
}

func Map[From any, To any](slice []From, f func(f From) To) []To {
	var to []To
	for _, from := range slice {
		to = append(to, f(from))
	}
	return to
}

func Filter[T any](slice []T, f func(f T) bool) []T {
	var out []T
	for _, t := range slice {
		if !f(t) {
			continue
		}
		out = append(out, t)
	}
	return out
}
