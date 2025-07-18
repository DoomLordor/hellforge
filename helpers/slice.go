package helpers

// Map returns slice with elements of source slice processed by mapper func
func Map[I, O any](data []I, fn func(I) O) []O {
	out := make([]O, len(data))

	for i, t := range data {
		out[i] = fn(t)
	}

	return out
}

// MapValuesToSlice returns map values as slice
func MapValuesToSlice[T any, D comparable](data map[D]*T) []*T {
	slice := make([]*T, 0, len(data))
	for _, obj := range data {
		slice = append(slice, obj)
	}

	return slice
}

// MapKeysToSlice convert map keys to slice
func MapKeysToSlice[K comparable, T any](items map[K]T) []K {
	keys := make([]K, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	return keys
}

// SliceElementsToMap convert slice elements to map struct
func SliceElementsToMap[K comparable](items []K) map[K]struct{} {
	result := make(map[K]struct{}, len(items))
	for _, item := range items {
		result[item] = struct{}{}
	}

	return result
}

// SliceElementsNotOrIn returns first slice elements contains or not in second slice
func SliceElementsNotOrIn[K comparable](firstSlice, secondSlice []K, in bool) []K {
	uniqueMap := SliceElementsToMap(secondSlice)

	result := make([]K, 0, len(firstSlice))
	for _, v := range firstSlice {
		if _, ok := uniqueMap[v]; ok == in {
			result = append(result, v)
		}
	}

	return result[:len(result):len(result)]
}

// EqualSliceElements check to slice equal elements
func EqualSliceElements[K comparable](firstSlice, secondSlice []K) bool {
	if len(firstSlice) != len(secondSlice) {
		return false
	}

	return len(SliceElementsNotOrIn(firstSlice, secondSlice, false)) == 0
}

// UniqueSliceElementsByKeyFunc returns slice with unique elements of source slice by keyFunc
func UniqueSliceElementsByKeyFunc[T any, K comparable](items []T, keyFunc func(T) K) []T {
	uniqueKeys := make(map[K]struct{}, len(items))
	var result []T

	for _, item := range items {
		key := keyFunc(item)
		if _, exists := uniqueKeys[key]; !exists {
			uniqueKeys[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// UniqueSliceElements returns unique slice elements
func UniqueSliceElements[K comparable](items []K) []K {
	uniqueKeys := make(map[K]struct{})
	var result []K

	for _, item := range items {
		if _, exists := uniqueKeys[item]; !exists {
			uniqueKeys[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// NotNilSlice returns slice without nil values
func NotNilSlice[T any](slice []*T) []T {
	result := make([]T, 0, len(slice))

	for _, v := range slice {
		if v != nil {
			result = append(result, *v)
		}
	}

	return result
}
