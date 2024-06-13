package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashEncoder(p string) (string, error) {
	encoded, err := bcrypt.GenerateFromPassword([]byte(p), 10)
	return string(encoded), err
}

func HashIsMatched(hashed, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}
