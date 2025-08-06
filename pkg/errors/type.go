package errors

import "errors"

var (
	ErrPasswordNotMatch         = errors.New("password not valid")
	ErrUserAlreadyExists        = errors.New("cannot signup new user because use already exists")
	ErrUserAlreadyVerified      = errors.New("user already verified")
	ErrGenerateToken            = errors.New("failed to Generate jwt token")
	ErrTokenExpired             = errors.New("token Expired")
	ErrTokenFailedToGetUserName = errors.New("failed to get username from token")
	ErrTokenFailedToGetRole     = errors.New("failed to get role from token")
	ErrInvalidToken             = errors.New("invalid token")
	ErrUserNotVerified          = errors.New("user is not verified yet")
	ErrRoleNotFound             = errors.New("role not found")

	ErrGroupAlreadyexisted = errors.New("group already exists")
)
