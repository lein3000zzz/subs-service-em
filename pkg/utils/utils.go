package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateID() (string, error) {
	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
