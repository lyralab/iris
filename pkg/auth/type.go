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
	// GenerateToken takes username, role id and whether the password must be changed.
	GenerateToken(string, string, bool) (string, error)
	// ValidateToken returns username, role name and the must-change-password flag.
	ValidateToken(string) (string, string, bool, error)
}
