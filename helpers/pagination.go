package helpers

// GetHasNextAndResponse checks if entity has more data by filter and get corrected response
func GetHasNextAndResponse[T any](limit uint64, response []T) ([]T, bool) {
	responseSize := uint64(len(response))
	return response[:min(responseSize, limit-1)], responseSize == limit
}
