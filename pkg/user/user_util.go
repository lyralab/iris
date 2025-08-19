package user

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"

	"golang.org/x/crypto/bcrypt"
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
	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	password := make([]byte, length)
	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			return "", err
		}
		password[i] = charSet[randomIndex.Int64()]
	}

	return string(password), nil
}
