package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	iriserror "github.com/root-ali/iris/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type AuthRoleRepository interface {
	GetRoleByID(string) (string, string, error)
}

func NewAuthService(secretKey []byte, rs AuthRoleRepository, logger *zap.SugaredLogger) AuthServiceInterface {
	return &authService{
		secretKey: secretKey,
		rs:        rs,
		logger:    logger,
	}
}

func (as *authService) GenerateToken(username string, role string, mustChangePassword bool) (string, error) {
	roleName, roleAccess, err := as.rs.GetRoleByID(role)
	as.logger.Infow("role information", "name", roleName, "access", roleAccess)
	if err != nil {
		as.logger.Info("error is ", err)
		return "", err
	}
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &jwt.MapClaims{
		"username":             username,
		"role":                 roleName,
		"accesses":             roleAccess,
		"exp":                  expirationTime.Unix(),
		"must_change_password": mustChangePassword,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(as.secretKey)
	if err != nil {
		as.logger.Errorw("Failed to generate token",
			"user", username, "error", err)
		return "", iriserror.ErrGenerateToken
	}
	return tokenString, nil
}

func (as *authService) ValidateToken(tokenString string) (string, string, bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return as.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			as.logger.Errorw("Token expired", "error", err)
			return "", "", false, iriserror.ErrTokenExpired
		} else {
			as.logger.Errorw("Failed to parse token", "error", err)
			return "", "", false, err

		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			as.logger.Errorw("Token expired", "error", err)
			return "", "", false, iriserror.ErrTokenExpired
		}
		username, ok := claims["username"].(string)
		if !ok {
			as.logger.Errorw("Failed to get username from token")
			return "", "", false, iriserror.ErrTokenFailedToGetUserName
		}
		role, ok := claims["role"].(string)
		if !ok {
			as.logger.Errorw("Failed to get role from token")
			return "", "", false, iriserror.ErrTokenFailedToGetRole
		}
		// Tokens issued before this claim existed carry no flag.
		mustChangePassword, _ := claims["must_change_password"].(bool)
		return username, role, mustChangePassword, nil
	} else {
		as.logger.Errorw("Invalid token")
		return "", "", false, iriserror.ErrInvalidToken
	}

}
