package smsir

import (
	"net/http"

	"go.uber.org/zap"
)

type SmsirService struct {
	Client     *http.Client
	LineNumber string
	Priority   int
	Logger     *zap.SugaredLogger
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
