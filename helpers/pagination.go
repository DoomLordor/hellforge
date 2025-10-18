package helpers

// GetHasNextAndResponseSize checks if entity has more data by filter and get corrected response size
func GetHasNextAndResponseSize(limit, responseSize uint64, needPagination bool) (bool, uint64) {
	if !needPagination {
		return false, responseSize
	}

	if limit >= responseSize {
		return false, responseSize
	}

	return true, responseSize - 1
}
