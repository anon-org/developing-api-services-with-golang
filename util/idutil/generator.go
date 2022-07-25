package idutil

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func MustGenerateID(n uint8) string {
	b := make([]byte, n/2)
	_, err := rand.Read(b)

	if err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(b)
}
