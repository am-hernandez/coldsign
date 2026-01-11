package logo

import (
	"fmt"
	"os"
)

// Print prints the coldsign ASCII art logo to stderr.
func Print() {
	fmt.Fprint(os.Stderr, `
░█▀▀░█▀█░█░░░█▀▄░█▀▀░▀█▀░█▀▀░█▀█
░█░░░█░█░█░░░█░█░▀▀█░░█░░█░█░█░█
░▀▀▀░▀▀▀░▀▀▀░▀▀░░▀▀▀░▀▀▀░▀▀▀░▀░▀
air-gapped Ethereum transaction signer
`)
}

