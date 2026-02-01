package asiatech

import (
	"net/http"

	"github.com/root-ali/iris/pkg/cache"
	"go.uber.org/zap"
)

type Service struct {
	host     string
	username string
	password string
	sender   string
	scope    string
	priority int

	c cache.Interface[string, string]

	logger *zap.SugaredLogger
}

type customTransport struct {
	Base   http.RoundTripper
	Header http.Header
}

type SendMessage struct {
	Sender   string `json:"sourceAddress"`
	Receptor string `json:"destinationAddress"`
	Text     string `json:"messageText"`
	Coding   int    `json:"dataCoding"`
}

type SendMessageResponse struct {
	Message    string   `json:"message"`
	Succeeded  bool     `json:"succeeded"`
	Data       []string `json:"data"`
	ResultCode int      `json:"resultCode"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpireAt    string `json:"expires_at"`
}

type DeliveryMessageRequest []string

type DeliveryMessageResponse struct {
	Message    string `json:"message"`
	ResultCode int    `json:"resultCode"`
	Data       []struct {
		ID             string `json:"id"`
		DeliveryStatus int    `json:"deliveryStatus"`
		DeliveryTime   string `json:"deliveryDate"`
	}
}

var (
	getTokenPath   = "/connect/token"
	sendSms        = "/api/1/message/send"
	deliveryStatus = "/api/message/getdlr"
)
