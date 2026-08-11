package common

import "github.com/rs/xid"

// NewPublicID creates a prefixed public identifier.
func NewPublicID(prefix string) string {
	return prefix + "_" + xid.New().String()
}

// TruncateRunes limits a string by Unicode code points.
func TruncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
