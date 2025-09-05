package helpers

// GetHasNextAndResponseSize checks if entity has more data by filter and get corrected response size
func GetHasNextAndResponseSize(limit, responseSize int) (bool, int) {
	if limit >= responseSize {
		return false, responseSize
	}

	return true, responseSize - 1
}
