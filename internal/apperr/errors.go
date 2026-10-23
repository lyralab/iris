// Package apperr holds the errors shared across iris layers: sentinels, a typed
// ASError carrying a Kind, and thin wrappers over stdlib errors so callers only
// import one errors package.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

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
	ErrPasswordReused           = errors.New("new password must differ from the current one")

	ErrGroupAlreadyexisted = errors.New("group already exists")
	ErrGroupNotFound       = errors.New("group not found")

	ErrProviderNotFound = errors.New("provider not found")
)

func New(text string) error { return errors.New(text) }

func Is(err, target error) bool { return errors.Is(err, target) }

func As(err error, target any) bool { return errors.As(err, target) }

func Join(errs ...error) error { return errors.Join(errs...) }

// AsType finds the first error in err's chain of type T.
func AsType[T error](err error) (T, bool) { return errors.AsType[T](err) }

// Wrap annotates err with msg, keeping it matchable with Is/As. Nil stays nil.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Kind classifies an error so transports can map it without knowing every sentinel.
type Kind int

const (
	KindInternal Kind = iota
	KindInvalid
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
)

// ASError is a typed error. Message is safe to show clients; Err is the
// internal cause and stays reachable through Unwrap.
type ASError struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *ASError) Error() string {
	switch {
	case e.Err == nil:
		return e.Message
	case e.Message == "":
		return e.Err.Error()
	default:
		return e.Message + ": " + e.Err.Error()
	}
}

func (e *ASError) Unwrap() error { return e.Err }

// Is matches another *ASError with the same non-empty Code, so a wrapped copy
// still matches its template.
func (e *ASError) Is(target error) bool {
	t, ok := target.(*ASError)
	return ok && e.Code != "" && e.Code == t.Code
}

// HTTPStatus maps the Kind of the first ASError in err's chain to a status code.
// Errors without one are internal.
func HTTPStatus(err error) int {
	e, ok := AsType[*ASError](err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch e.Kind {
	case KindInvalid:
		return http.StatusBadRequest
	case KindUnauthorized:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
