package smsir

import "go.uber.org/zap"

type smsirService struct {
	ApiKey string
	Logger *zap.SugaredLogger
}
