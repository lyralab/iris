package apperr

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestIsThroughWrapping(t *testing.T) {
	typed := &ASError{Kind: KindNotFound, Code: "group_not_found", Message: "group not found", Err: ErrGroupNotFound}
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"bare sentinel", ErrGroupNotFound, ErrGroupNotFound, true},
		{"wrapped sentinel", Wrap(ErrGroupNotFound, "get group"), ErrGroupNotFound, true},
		{"double wrapped", Wrap(Wrap(ErrGroupNotFound, "repo"), "service"), ErrGroupNotFound, true},
		{"other sentinel", Wrap(ErrGroupNotFound, "get group"), ErrRoleNotFound, false},
		{"sentinel inside ASError", Wrap(typed, "handler"), ErrGroupNotFound, true},
		{"ASError by code", Wrap(typed, "handler"), &ASError{Code: "group_not_found"}, true},
		{"ASError other code", typed, &ASError{Code: "role_not_found"}, false},
		{"empty code never matches", &ASError{}, &ASError{}, false},
		{"joined", Join(ErrRoleNotFound, ErrGroupNotFound), ErrGroupNotFound, true},
		{"nil", nil, ErrGroupNotFound, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Fatalf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAsAndAsType(t *testing.T) {
	typed := &ASError{Kind: KindConflict, Code: "c"}
	err := Wrap(typed, "ctx")

	var target *ASError
	if !As(err, &target) || target != typed {
		t.Fatalf("As did not find ASError")
	}
	if got, ok := AsType[*ASError](err); !ok || got != typed {
		t.Fatalf("AsType did not find ASError")
	}
	if _, ok := AsType[*ASError](ErrGroupNotFound); ok {
		t.Fatalf("AsType matched a plain sentinel")
	}
	syntax := json.Unmarshal([]byte("{"), &struct{}{})
	if _, ok := AsType[*json.SyntaxError](Wrap(syntax, "decode")); !ok {
		t.Fatalf("AsType did not find *json.SyntaxError")
	}
}

func TestWrapAndUnwrap(t *testing.T) {
	if Wrap(nil, "x") != nil {
		t.Fatalf("Wrap(nil) != nil")
	}
	if got := Wrap(ErrGroupNotFound, "get").Error(); got != "get: group not found" {
		t.Fatalf("Wrap().Error() = %q", got)
	}
	e := &ASError{Message: "safe", Err: ErrRoleNotFound}
	if e.Unwrap() != ErrRoleNotFound {
		t.Fatalf("Unwrap lost cause")
	}
	for _, tt := range []struct {
		e    *ASError
		want string
	}{
		{&ASError{Message: "safe"}, "safe"},
		{&ASError{Err: ErrRoleNotFound}, "role not found"},
		{e, "safe: role not found"},
	} {
		if got := tt.e.Error(); got != tt.want {
			t.Fatalf("Error() = %q, want %q", got, tt.want)
		}
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{&ASError{Kind: KindInvalid}, http.StatusBadRequest},
		{&ASError{Kind: KindUnauthorized}, http.StatusUnauthorized},
		{&ASError{Kind: KindForbidden}, http.StatusForbidden},
		{Wrap(&ASError{Kind: KindNotFound}, "ctx"), http.StatusNotFound},
		{&ASError{Kind: KindConflict}, http.StatusConflict},
		{&ASError{Kind: KindInternal}, http.StatusInternalServerError},
		{ErrGroupNotFound, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		if got := HTTPStatus(tt.err); got != tt.want {
			t.Errorf("HTTPStatus(%v) = %d, want %d", tt.err, got, tt.want)
		}
	}
}
