package jwtvalidation

import (
	"go.uber.org/zap"
	"time"
)

type JWTIssue struct {
	SecretKey []byte
	Issuer    string
	Expire    time.Duration
	l         *zap.SugaredLogger
}
