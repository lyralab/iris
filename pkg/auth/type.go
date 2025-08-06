package auth

import (
	"go.uber.org/zap"
)

type authService struct {
	secretKey []byte
	rs        AuthRoleRepository
	logger    *zap.SugaredLogger
}

type AuthServiceInterface interface {
	GenerateToken(string, string) (string, error)
	ValidateToken(string) (string, string, error)
}
