package bale

import (
	"net/http"

	"go.uber.org/zap"
)

type Service struct {
	Url      string
	Token    string
	Client   *http.Client
	Priority int
	Logger   *zap.SugaredLogger
}

type SendMessage struct {
	ChatID  int64  `json:"chat_id"`
	Message string `json:"text"`
}

type SendMessageResponse struct {
	OK      bool   `json:"ok"`
	Date    int    `json:"date"`
	Message string `json:"text"`
}

type GetMeResponse struct {
	OK bool `json:"ok"`
}

type customTransport struct {
	Base   http.RoundTripper
	Header http.Header
}

type errorStack []error

func (e *errorStack) Append(err error) {
	*e = append(*e, err)
}

func (e errorStack) Error() string {
	errMsg := ""
	for _, err := range e {
		if err == nil {
			errMsg += "nil;"
		} else {
			errMsg += err.Error() + ";"
		}
	}
	return errMsg
}
