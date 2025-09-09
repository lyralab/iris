package smsir

import (
	"net/http"

	"github.com/root-ali/iris/pkg/scheduler/cache_receptors"

	"go.uber.org/zap"
)

type SmsirService struct {
	Client     *http.Client
	LineNumber string
	Priority   int
	Logger     *zap.SugaredLogger
	cache      cache_receptors.CacheService
}

type customTransport struct {
	Base   http.RoundTripper
	Header http.Header
}

type SendSMSRequestBody struct {
	Mobiles     []string `json:"mobiles"`
	MessageText string   `json:"messageText"`
	LineNumber  int      `json:"lineNumber,omitempty"`
}

type VerifyResponseBody struct {
	Status  int     `json:"status"`
	Message string  `json:"message"`
	Data    float64 `json:"data"`
}

type SendSmsResponseBody struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []int  `json:"data"`
}
