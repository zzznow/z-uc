package models

import (
	"crypto/rand"
	"fmt"
)

func GenerateSN(userId uint64) string {
	b := make([]byte, 8)
	rand.Read(b)
	suffix := fmt.Sprintf("%x", b)[:8]
	return fmt.Sprintf("U%x-%s", userId, suffix)
}

func RandomNickName() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("fy_%x", b)
}
