package kavenegar

import (
	kn "github.com/kavenegar/kavenegar-go"
	"go.uber.org/zap"
)

type kavenegarService struct {
	API    *kn.Kavenegar
	Sender string
	Logger *zap.SugaredLogger
}
