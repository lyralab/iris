package user

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/root-ali/iris/internal/apperr"
	"golang.org/x/crypto/bcrypt"
)

func generateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", apperr.New("ERROR_GENERATING_SALT")
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func hashPassword(password, salt string) (string, error) {
	saltedPassword := []byte(password + salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(saltedPassword, bcrypt.DefaultCost)
	if err != nil {
		return "", apperr.New("ERROR_GENERATING_HASH_PASSWORD")
	}
	return string(hashedPassword), nil
}

// meetsSigninPolicy mirrors the signin passwordStrength rule, so a generated password can log in.
func meetsSigninPolicy(p string) bool {
	return strings.ContainsAny(p, "abcdefghijklmnopqrstuvwxyz") &&
		strings.ContainsAny(p, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") &&
		strings.ContainsAny(p, "0123456789") &&
		strings.ContainsAny(p, "@$!%*?&#^")
}

func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", apperr.New("password length must be greater than zero")
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
