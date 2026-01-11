package helpers

import "strings"

// ShortAddr formats an address as first 6 chars + … + last 4 chars.
// If the address is 12 chars or less, returns it unchanged.
func ShortAddr(a string) string {
	a = strings.TrimSpace(a)
	if len(a) <= 12 {
		return a
	}
	return a[:6] + "…" + a[len(a)-4:]
}

