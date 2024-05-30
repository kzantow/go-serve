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
