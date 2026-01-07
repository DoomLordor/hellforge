package helpers

// GetHasNextAndResponse checks if entity has more data by filter and get corrected response
func GetHasNextAndResponse[T any](response []T, limit uint64) ([]T, bool) {
	responseSize := uint64(len(response))
	return response[:min(responseSize, limit-1)], responseSize == limit
}
