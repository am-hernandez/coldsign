package helpers

// ZeroString clears a string's memory by overwriting it with zeros.
// This is a best-effort attempt to prevent secrets from persisting in memory.
func ZeroString(s *string) {
	b := []byte(*s)
	for i := range b {
		b[i] = 0
	}
	*s = ""
}
