package passc

import (
	"crypto/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

func GenerateRandomPassword(length int) (*string, error) {
	password := make([]byte, length)

	_, err := rand.Read(password)
	if err != nil {
		return nil, err
	}

	for i := 0; i < length; i++ {
		password[i] = charset[int(password[i])%len(charset)]
	}

	str := string(password)
	return &str, nil
}

func alignPassword(password string) string {
	length := len(password)
	if length < 16 {
		return password + strings.Repeat("*", 16-length)
	}
	if length > 16 {
		return password[:16]
	}
	return password
}
