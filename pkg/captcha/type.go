package captcha

// CaptchaServiceInterface defines the contract for captcha service
type CaptchaServiceInterface interface {
	GenerateCaptcha() (string, string, error)
	VerifyCaptcha(string, string) bool
}
