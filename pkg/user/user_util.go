package user

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"math/big"
)

func generateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", errors.New("ERROR_GENERATING_SALT")
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func hashPassword(password, salt string) (string, error) {
	saltedPassword := []byte(password + salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(saltedPassword, bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ERROR_GENERATING_HASH_PASSWORD")
	}
	return string(hashedPassword), nil
}
func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("password length must be greater than zero")
	}

	// Define the character set for the password.
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

	// Create a byte slice to store the password.
	password := make([]byte, length)

	// Generate random bytes and select characters from the character set.
	for i := 0; i < length; i++ {
		// Generate a random index within the character set.
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			return "", err
		}

		// Select the character at the random index.
		password[i] = charSet[randomIndex.Int64()]
	}

	// Return the password as a string.
	return string(password), nil
}
