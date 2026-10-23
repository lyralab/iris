package jwtvalidation

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/user"
	"go.uber.org/zap"
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
		"sub": user.UserName,
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
	ji.l.Info("jwt issuer create token for user", zap.String("user", user.UserName))
	return tokenString, nil
}

func (ji *JWTIssue) ValidateToken(tokenString string, _ string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return ji.SecretKey, nil
	})
	if err != nil {
		ji.l.Errorw("jwt issuer parse token err", "error", err)
		return err
	}
	if !token.Valid {
		return apperr.ErrInvalidToken
	}
	return nil
}
