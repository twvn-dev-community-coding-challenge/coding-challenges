// Package ids generates unique, log-friendly identifiers such as R-3f9c1a2b7d4e.
package ids

import (
	"crypto/rand"
	"encoding/hex"
)

func New(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error
	return prefix + "-" + hex.EncodeToString(b)
}
