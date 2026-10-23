// Package apperr holds the errors shared across iris layers: sentinels, a typed
// ASError carrying a Kind, and thin wrappers over stdlib errors so callers only
// import one errors package.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinels. Messages of non-internal kinds are safe to return to clients;
// KindInternal messages are for logs only.
var (
	ErrPasswordNotMatch    = sentinel(KindUnauthorized, "password_not_match", "password does not match")
	ErrPasswordReused      = sentinel(KindInvalid, "password_reused", "new password must differ from the current one")
	ErrUserAlreadyExists   = sentinel(KindConflict, "user_already_exists", "user already exists")
	ErrUserAlreadyVerified = sentinel(KindConflict, "user_already_verified", "user already verified")
	ErrUserNotVerified     = sentinel(KindForbidden, "user_not_verified", "user is not verified yet")
	ErrUserNotFound        = sentinel(KindNotFound, "user_not_found", "user not found")
	ErrUsernameRequired    = sentinel(KindInvalid, "username_required", "username is required")
	ErrUserIDRequired      = sentinel(KindInvalid, "user_id_required", "user id is required")
	ErrRoleNotFound        = sentinel(KindNotFound, "role_not_found", "role not found")

	ErrGenerateToken            = sentinel(KindInternal, "generate_token", "failed to generate jwt token")
	ErrTokenExpired             = sentinel(KindUnauthorized, "token_expired", "token expired")
	ErrTokenFailedToGetUserName = sentinel(KindUnauthorized, "token_no_username", "failed to get username from token")
	ErrTokenFailedToGetRole     = sentinel(KindUnauthorized, "token_no_role", "failed to get role from token")
	ErrInvalidToken             = sentinel(KindUnauthorized, "invalid_token", "invalid token")

	ErrGroupAlreadyExists = sentinel(KindConflict, "group_already_exists", "group already exists")
	ErrGroupNotFound      = sentinel(KindNotFound, "group_not_found", "group not found")

	ErrAlertNotFound = sentinel(KindNotFound, "alert_not_found", "alert not found")

	ErrProviderNotFound      = sentinel(KindNotFound, "provider_not_found", "provider not found")
	ErrNoActiveProvider      = sentinel(KindInternal, "no_active_provider", "no active provider found")
	ErrNoAlternativeProvider = sentinel(KindInternal, "no_alternative_provider", "no alternative provider found")
	ErrBotNotActive          = sentinel(KindInternal, "bot_not_active", "bot not active")

	ErrServiceAlreadyStarted = sentinel(KindInternal, "service_already_started", "service already started")
)

func sentinel(kind Kind, code, msg string) error {
	return &ASError{Kind: kind, Code: code, Message: msg}
}

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

// PublicMessage returns text safe to show a client: the Message of the first
// non-internal ASError in err's chain, or a generic text otherwise.
func PublicMessage(err error) string {
	if e, ok := AsType[*ASError](err); ok && e.Kind != KindInternal && e.Message != "" {
		return e.Message
	}
	return "internal server error"
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
