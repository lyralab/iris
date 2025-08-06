package captcha

import (
	"math/rand"
	"time"

	"github.com/mojocn/base64Captcha"

	"go.uber.org/zap"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type CaptchaService struct {
	Captcha *base64Captcha.Captcha
	logger  *zap.SugaredLogger
}

func NewCaptchaService(logger *zap.SugaredLogger) *CaptchaService {
	store := base64Captcha.NewMemoryStore(100, 10*time.Minute)
	driver := base64Captcha.NewDriverDigit(100, 240, 2, 0.2, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	service := &CaptchaService{
		Captcha: captcha,
		logger:  logger,
	}

	return service
}

// GenerateCaptcha creates a new math captcha
func (cs *CaptchaService) GenerateCaptcha() (string, string, error) {
	id, b64s, answer, err := cs.Captcha.Generate()
	if err != nil {
		cs.logger.Errorw("Failed to generate captcha", "error", err)
		return "", "", err
	}
	err = cs.Captcha.Store.Set(id, answer)
	if err != nil {
		cs.logger.Errorw("Failed to store captcha", "error", err)
		return "", "", err
	}
	cs.logger.Infow("Captcha generated", "captcha_id", id)
	return id, b64s, nil
}

func (cs *CaptchaService) VerifyCaptcha(id, answer string) bool {
	if answer == "" {
		cs.logger.Warnw("Empty captcha answer", "captcha_id", id)
		return false
	}
	isValid := cs.Captcha.Verify(id, answer, true)
	if isValid {
		cs.logger.Infow("Captcha verified", "captcha_id", id)
	} else {
		cs.logger.Warnw("Captcha verification failed", "captcha_id", id)
	}
	return isValid
}
