package kavenegar

import (
	kn "github.com/kavenegar/kavenegar-go"
	"go.uber.org/zap"
)

type KavenegarService struct {
	API      *kn.Kavenegar
	Sender   string
	Priority int
	Logger   *zap.SugaredLogger
}
