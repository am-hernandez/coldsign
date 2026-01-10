package helpers

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ReadHiddenLineFromTTY reads a secret from the terminal with hidden input.
// If optional is false, empty input returns an error.
func ReadHiddenLineFromTTY(prompt string, optional bool) (string, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return "", fmt.Errorf("no TTY available: %w", err)
	}
	defer tty.Close()

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, prompt)
	fmt.Fprint(os.Stderr, "> ")

	b, err := term.ReadPassword(int(tty.Fd()))
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}

	// Convert BEFORE wiping
	s := strings.TrimSpace(string(b))

	// Normalize to avoid paste issues
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.Join(strings.Fields(s), " ")

	// Best-effort wipe of raw bytes
	for i := range b {
		b[i] = 0
	}

	if s == "" && !optional {
		return "", fmt.Errorf("empty input")
	}
	return s, nil
}
