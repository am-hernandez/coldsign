package helpers

import (
	"fmt"
	"math/big"
	"strings"
)

type Unit int

const (
	Wei Unit = iota
	Gwei
	Eth
)

// FormatWeiString converts a base-10 wei string into the requested unit,
// returning a decimal string with `decimals` digits after the decimal point.
// For Unit=Wei, decimals is ignored and the original integer is returned normalized.
func FormatWeiString(weiStr string, unit Unit, decimals int) (string, error) {
	weiStr = strings.TrimSpace(weiStr)
	if weiStr == "" {
		return "", fmt.Errorf("empty value")
	}

	wei, ok := new(big.Int).SetString(weiStr, 10)
	if !ok {
		return "", fmt.Errorf("invalid base-10 integer")
	}

	// Guardrail: keep output sane
	if decimals < 0 {
		return "", fmt.Errorf("invalid decimals")
	}
	if decimals > 30 {
		return "", fmt.Errorf("decimals too large")
	}

	switch unit {
	case Wei:
		return wei.String(), nil
	case Gwei:
		return formatByExp(wei, 9, decimals), nil
	case Eth:
		return formatByExp(wei, 18, decimals), nil
	default:
		return "", fmt.Errorf("unknown unit")
	}
}

// formatByExp formats integer `n` as n / 10^exp with `decimals` fractional digits.
func formatByExp(n *big.Int, exp int, decimals int) string {
	den := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exp)), nil)
	r := new(big.Rat).SetFrac(n, den)
	return r.FloatString(decimals)
}

// Convenience wrappers (encode conventions)
func FormatETH(weiStr string) (string, error)  { return FormatWeiString(weiStr, Eth, 18) }
func FormatETH6(weiStr string) (string, error) { return FormatWeiString(weiStr, Eth, 6) }
func FormatGwei(weiStr string) (string, error) { return FormatWeiString(weiStr, Gwei, 2) }
