package smsir

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func NewSmsirService(apikey string, lineNumber string, p int, logger *zap.SugaredLogger) *Service {
	client := createAPIHandler(apikey)
	return &Service{Client: client, Priority: p, LineNumber: lineNumber, Logger: logger}
}

func (s *Service) Send(message notifications.Message) ([]string, error) {
	lineNumber, _ := strconv.Atoi(s.LineNumber)
	text := ""
	if message.State == "firing" {
		text = "🚨" + message.Subject + "\n" + message.Message + "\nTime: " + message.Time
	} else if message.State == "resolved" {
		text = "✅" + message.Subject + "\n" + message.Message + "\nTime: " + message.Time
	}
	requestBody := SendSMSRequestBody{
		Mobiles:     message.Receptors,
		MessageText: text,
		LineNumber:  lineNumber,
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

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Errorw("Cannot read response body", "error", err)
		return nil, err
	}
	var rsp SendSmsResponse
	err = json.Unmarshal(bodyData, &rsp)
	s.Logger.Infow("Smsir returned response", " body", string(bodyData))
	s.Logger.Infow("Smsir response status", "status",
		rsp.Status, "message", rsp.Message, "data", rsp.Data)
	if err != nil {
		s.Logger.Errorw("Cannot unmarshal response body", "error", rsp)
		return nil, err
	}
	var messageIDs []string
	for _, id := range rsp.Data.MessageID {
		messageIDs = append(messageIDs, strconv.Itoa(id))
	}

	s.Logger.Infow("Smsir service sent message successfully", "status", resp.StatusCode)
	return messageIDs, nil
}

func (s *Service) Status(messageId string) (notifications.MessageStatusType, error) {
	messageStatus := notifications.MessageStatusType(1)
	req, err := http.NewRequest("GET", Host+"/v1/send/"+messageId, nil)
	if err != nil {
		s.Logger.Errorw("Cannot build request", "error", err)
		messageStatus = notifications.MessageStatusType(-1)
		return messageStatus, err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		s.Logger.Errorw("Cannot send request", "error", err)
		messageStatus = notifications.MessageStatusType(-1)
		return messageStatus, err
	}
	s.Logger.Infow("Smsir status response", "status_code", resp.StatusCode, "messageId", messageId)
	if resp.StatusCode != http.StatusOK {
		messageStatus = notifications.MessageStatusType(-1)
		return messageStatus, fmt.Errorf("http status code %d", resp.StatusCode)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.Logger.Errorw("Cannot read response body", "error", err)
			return
		}
	}(resp.Body)
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Errorw("Cannot read response body", "error", err)
		messageStatus = notifications.MessageStatusType(-1)
		return messageStatus, err
	}

	var rsp GetStatusResponse
	err = json.Unmarshal(bodyData, &rsp)
	if err != nil {
		s.Logger.Errorw("Cannot unmarshal response body", "error", err)
		messageStatus = notifications.MessageStatusType(-1)
		return messageStatus, err
	}
	s.Logger.Infow("Smsir status response", " body", rsp.Data.Mobile, "messageId", messageId)
	if rsp.Status == 1 {
		messageStatus = notifications.MessageStatusType(10)
		return messageStatus, nil
	}
	if rsp.Data.DeliveryStatus == 3 || rsp.Data.DeliveryStatus == 5 {
		messageStatus = notifications.MessageStatusType(1)
		return messageStatus, nil
	}
	if rsp.Data.DeliveryStatus == 6 || rsp.Data.DeliveryStatus == 7 || rsp.Data.DeliveryStatus == 2 || rsp.Data.DeliveryStatus == 4 {
		messageStatus = notifications.MessageStatusType(6)
		return messageStatus, errors.New("failed to deliver message")
	}
	messageStatus = notifications.TypeMessageStatusDelivered
	s.Logger.Infow("Smsir status message", "status", messageStatus, "messageId", messageId)
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

func (s *Service) GetName() string {
	return "Smsir"
}

func (s *Service) GetFlag() string {
	return "sms"
}

func (s *Service) GetPriority() int {
	return s.Priority
}

func (s *Service) Verify() (string, error) {
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
