package ui

func ShortHash(hash string) string {
	return hash[:8]
}

func ShortHashLen(hash string, l int) string {
	return hash[:l]
}
