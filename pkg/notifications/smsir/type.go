package smsir

import (
	"net/http"

	"go.uber.org/zap"
)

type Service struct {
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

type SendSmsResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		PackId    string `json:"packId"`
		MessageID []int  `json:"messageIds"`
	} `json:"data"`
	Cost float64 `json:"cost"`
}

type GetStatusResponse struct {
	Status int `json:"status"`
	Data   struct {
		MessageId      int `json:"messageId"`
		Mobile         int `json:"mobile"`
		DeliveryStatus int `json:"deliveryState"`
	} `json:"data"`
}
