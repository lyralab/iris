package jwtvalidation

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
	"time"
)

func NewJWTIssuerService(secretKey []byte, issuer string, t time.Duration, l *zap.SugaredLogger) *JWTIssue {
	return &JWTIssue{
		SecretKey: secretKey,
		Issuer:    issuer,
		Expire:    t,
		l:         l,
	}
}

func (ji *JWTIssue) CreateToken(user user.User) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Name,
		"iss": ji.Issuer,
		"aud": user.Role,
		"exp": time.Now().Add(ji.Expire).Unix(),
		"iat": time.Now().Unix(),
	})
	tokenString, err := claims.SignedString(ji.SecretKey)

	if err != nil {
		ji.l.Errorf("jwt issuer create token err: %v", err)
		return "", err
	}
	ji.l.Info("jwt issuer create token for user", zap.String("user", user.Name))
	return tokenString, nil
}

func (ji *JWTIssue) ValidateToken(tokenString string, _ string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return ji.SecretKey, nil
	})
	if err != nil {
		ji.l.Error("jwt issuer parse token err: %v", err)
		return err
	}
	if !token.Valid {
		return fmt.Errorf("jwt issuer invalid token")
	}
	var u *user.User
	u.Name = token.Claims.(jwt.MapClaims)["sub"].(string)
	u.Role = token.Claims.(jwt.MapClaims)["aud"].(string)
	return nil
}
