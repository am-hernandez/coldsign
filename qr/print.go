package qr

import (
	"os"

	"github.com/mdp/qrterminal/v3"
)

func PrintToTerminal(payload string) {
	cfg := qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     os.Stderr, // keep stdout clean for piping
		HalfBlocks: true,
	}
	qrterminal.GenerateWithConfig(payload, cfg)
}
