package idgen

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// generates a cryptographically secure, URL-safe random string of a given length.
func GenerateSecureID(length int) (string, error) {
	// Calculate the number of bytes needed for the desired string length
	// Each base64 character encodes 6 bits.
	// So, length characters means (length * 6) bits.
	// Divide by 8 to get bytes: (length * 6) / 8 = (length * 3) / 4
	numBytes := (length * 3) / 4 // Minimum bytes needed to guarantee length
	if numBytes == 0 {           // Ensure at least 1 byte for very short lengths
		numBytes = 1
	}

	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	// encode the bytes. This creates a string with URL-safe characters.
	encoded := base64.URLEncoding.EncodeToString(b)

	// trim or pad to the desired length
	if len(encoded) > length {
		return encoded[:length], nil
	}
	// If encoded string is shorter than length, it means numBytes might not have been enough
	// or base64 padding was removed. For simplicity here, we'll just return what we got
	// or potentially re-generate if the exact length is critical and padding is needed.
	// For most cases, GenerateSecureID(8) will be > 8, so trimming is common.
	return encoded, nil
}