package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = 12

func Hash(plaintext string) (string, error) {
	if len(plaintext) == 0 {
		return "", errors.New("password must not be empty")
	}
	if len(plaintext) > 72 {
		return "", errors.New("password must not exceed 72 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

func Verify(hash, plaintext string) error {
	if hash == "" || plaintext == "" {
		return errors.New("hash and plaintext must not be empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid credentials")
		}
		return fmt.Errorf("password verify: %w", err)
	}
	return nil
}
