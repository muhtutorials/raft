package main

// compareWords returns true if "a" is longer (lexicography breaking ties)
func compareWords(a, b string) bool {
	if len(a) == len(b) {
		return a < b
	}
	return len(a) > len(b)
}

func cloneWords(words [3]string) []string {
	var w [3]string
	copy(w[:], words[:])
	return w[:]
}
