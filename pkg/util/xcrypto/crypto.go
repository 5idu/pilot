package xcrypto

import (
	"errors"
	"fmt"

	cryptorand "crypto/rand"
)

// Bytes gets a slice of cryptographically random bytes of the given length and
// enforces that we check for short reads to avoid entropy exhaustion.
func Bytes(length int) ([]byte, error) {
	key := make([]byte, length)
	n, err := cryptorand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("could not read from random source: %v", err)
	}
	if n < length {
		return nil, errors.New("entropy exhausted")
	}
	return key, nil
}
