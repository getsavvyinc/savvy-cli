package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

const CommandPrefix = "cmd-"

func New(prefix string) string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return prefix + hex.EncodeToString(bytes)
}
