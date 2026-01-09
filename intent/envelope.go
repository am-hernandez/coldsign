package intent

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const EnvelopePrefixV1 = "coldintent:v1:"

// DecodeEnvelopeOrJSON takes either:
//   - raw JSON bytes (as string)
//   - or an envelope: "coldintent:v1:<base64url(json)>"
//
// It returns the underlying JSON bytes.
func DecodeEnvelopeOrJSON(input string) ([]byte, error) {
	s := strings.TrimSpace(input)
	s = strings.Trim(s, "\"") // tolerate scanners that wrap in quotes

	if strings.HasPrefix(s, EnvelopePrefixV1) {
		b64 := strings.TrimPrefix(s, EnvelopePrefixV1)
		b64 = strings.TrimSpace(b64)

		// base64 URL encoding WITHOUT padding is ideal for QR
		// RawURLEncoding expects no '=' padding
		decoded, err := base64.RawURLEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("invalid envelope base64url: %w", err)
		}
		return decoded, nil
	}

	// Fallback: assume raw JSON
	return []byte(s), nil
}
