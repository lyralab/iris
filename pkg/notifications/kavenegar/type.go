package kavenegar

import (
	kn "github.com/kavenegar/kavenegar-go"
	"github.com/root-ali/iris/pkg/scheduler/cache_receptors"
	"go.uber.org/zap"
)

type KavenegarService struct {
	API      *kn.Kavenegar
	Sender   string
	Priority int
	Logger   *zap.SugaredLogger
	cache    cache_receptors.CacheService
}
