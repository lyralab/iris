package smsir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/root-ali/iris/pkg/scheduler/cache_receptors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

var (
	Host = "https://api.sms.ir"
)

func NewSmsirService(apikey string, lineNumber int, p int, logger *zap.SugaredLogger,
	cacheService cache_receptors.CacheService) notifications.NotificationInterface {
	client := createAPIHandler(apikey)
	return &smsirService{cache: cacheService,Client: client, Priority: p, LineNumber: lineNumber, Logger: logger}
}

func (s *smsirService) Send(message notifications.Message) ([]string, error) {
	var sendGroupNumbers []string
	for _, group := range message.Receptors {
		nums, err := s.cache.GetNumbers(group)
		if err != nil {
			return nil, err
		}
		sendGroupNumbers = append(sendGroupNumbers, nums...)
	}
	requestBody := SendSMSRequestBody{
		Mobiles:     sendGroupNumbers,
		MessageText: message.Message,
		LineNumber:  s.LineNumber,
	}

	rbj, err := json.Marshal(requestBody)
	if err != nil {
		s.Logger.Errorw("Cannot marshal request body", "error", err)
		return nil, err
	}
	req, _ := http.NewRequest("POST", Host+"/v1/send/bulk", io.NopCloser(bytes.NewReader(rbj)))
	resp, err := s.Client.Do(req)
	if err != nil {
		s.Logger.Errorw("Cannot send request", "error", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)
		s.Logger.Errorw("Smsir returned non-200 status code",
			"body", string(respBody),
			"status", resp.StatusCode,
			"error", fmt.Errorf("http status code %d", resp.StatusCode),
		)
		return nil, fmt.Errorf("http status code %d", resp.StatusCode)
	}
	s.Logger.Infow("Smsir returned 200 status code", "status", resp.StatusCode)
	return nil, nil
}

func (s *smsirService) Status(string) (notifications.MessageStatusType, error) {
	var messageStatus notifications.MessageStatusType

	return messageStatus, nil
}

func createAPIHandler(apikey string) *http.Client {
	headers := http.Header{
		"Accept":       {"application/json"},
		"Content-Type": {"application/json"},
		"X-API-KEY":    {apikey},
	}
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &customTransport{
			Base:   http.DefaultTransport,
			Header: headers,
		},
	}
	return client
}

func (s *smsirService) GetName() string {
	return "Smsir"
}

func (s *smsirService) GetFlag() string {
	return "sms"
}

func (s *smsirService) GetPriority() int {
	return s.Priority
}

func (s *smsirService) Verify() (string, error) {
	req, _ := http.NewRequest("GET", Host+"/v1/credit", nil)
	resp, err := s.Client.Do(req)
	if err != nil {
		s.Logger.Errorw("Cannot send request", "error", err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)
		s.Logger.Errorw("Smsir returned non-200 status code",
			"body", string(respBody),
			"status", resp.StatusCode,
			"error", fmt.Errorf("http status code %d", resp.StatusCode),
		)
		return "", fmt.Errorf("http status code %d", resp.StatusCode)
	}
	respBody, _ := io.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	var verifyResponse VerifyResponseBody
	if err := json.Unmarshal(respBody, &verifyResponse); err != nil {
		s.Logger.Errorw("Cannot unmarshal response body", "error", err)
		return "", err
	}
	if verifyResponse.Status != 1 {
		s.Logger.Errorw("Smsir returned non-200 status code",
			"status", verifyResponse.Status,
			"message", verifyResponse.Message,
		)
		return "", fmt.Errorf("smsir returned non-200 status code: %d, message: %s", verifyResponse.Status, verifyResponse.Message)
	}
	s.Logger.Infow("Smsir verify response", "status", verifyResponse.Status, "message", verifyResponse.Message)
	return strconv.FormatFloat(verifyResponse.Data, 'f', -1, 64), nil
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	return t.Base.RoundTrip(req)
}
