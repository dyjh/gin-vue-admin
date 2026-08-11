package experience

import (
	"crypto/sha256"
	"encoding/hex"
)

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
