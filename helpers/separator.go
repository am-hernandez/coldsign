package helpers

import (
	"fmt"
	"strings"
)

const separatorWidth = 51

// Separator returns a separator line of '=' characters with optional centered text.
// If text is empty, returns a line of '=' characters.
// The total width is always separatorWidth.
func Separator(text string) string {
	if text == "" {
		return strings.Repeat("=", separatorWidth)
	}

	// Calculate padding needed on each side
	totalPadding := separatorWidth - len(text) - 2 // -2 for spaces around text
	if totalPadding < 0 {
		// If text is too long, just return it with minimal padding
		return fmt.Sprintf("= %s =", text)
	}

	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	return fmt.Sprintf("%s %s %s",
		strings.Repeat("=", leftPadding),
		text,
		strings.Repeat("=", rightPadding),
	)
}

