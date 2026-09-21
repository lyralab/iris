package bale

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func NewBaleService(url, token string, priority int, logger *zap.SugaredLogger) notifications.NotificationInterface {
	client := createClient()
	return &Service{Url: url, Token: token, Client: client, Priority: priority, Logger: logger}
}

func (s *Service) GetName() string {
	return "Bale"
}

func (s *Service) GetFlag() string {
	return "bale"
}

func (s *Service) GetPriority() int {
	return s.Priority
}

func (s *Service) Send(message notifications.Message) ([]string, error) {
	var errStack errorStack
	responses := make([]string, 0)
	text := ""
	if message.State == "firing" {
		text = `🚨 * Firing * 🚨` + "\n\n * " + message.Subject + " * \n\n" + message.Message + "\n\n" + message.Time
	} else if message.State == "resolved" {
		text = `✅ * Resolved * ✅` + "\n\n * " + message.Subject + " * \n\n" + message.Message + "\n\n" + message.Time
	}
	for _, receptor := range message.Receptors {
		chatID, err := strconv.ParseInt(receptor, 10, 64)
		if err != nil {
			s.Logger.Errorw("Cannot parse chat id", "receptor", receptor, "error", err)
			errStack.Append(err)
			continue
		}
		sendMessage := SendMessage{
			ChatID:  chatID,
			Message: text,
		}

		requestBody, err := json.Marshal(sendMessage)
		if err != nil {
			s.Logger.Errorw("Cannot marshal send message", "message", sendMessage, "error", err)
			errStack.Append(err)
			continue
		}
		req, err := http.NewRequest("POST", s.Url+"/bot"+s.Token+"/sendMessage", bytes.NewBuffer(requestBody))
		if err != nil {
			s.Logger.Errorw("Cannot create request", "message", sendMessage, "error", err)
			errStack.Append(err)
			continue
		}
		resp, err := s.Client.Do(req)
		if err != nil {
			s.Logger.Errorw("Cannot send request", "message", sendMessage, "error", err)
			errStack.Append(err)
			continue
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				s.Logger.Errorw("Cannot close body", "message", sendMessage, "error", err)
			}
		}(resp.Body)
		if resp.StatusCode != http.StatusOK {
			s.Logger.Errorw("Cannot send message", "message", sendMessage, "statusCode", resp.StatusCode)
			errStack.Append(errors.New(resp.Status))
			continue
		}
		var sendResponse SendMessageResponse
		err = json.NewDecoder(resp.Body).Decode(&sendResponse)
		if err != nil {
			s.Logger.Errorw("Cannot parse response", "message", sendMessage, "error", err)
			errStack.Append(err)
			continue
		}
		if !sendResponse.OK {
			s.Logger.Errorw("Cannot send message", "message", sendMessage)
			errStack.Append(errors.New(sendResponse.Message))
			continue
		}
		responses = append(responses, sendResponse.Message)
	}
	if len(errStack) > 0 {
		return responses, errStack
	}
	return responses, nil
}

func (s *Service) Status(messageID string) (notifications.MessageStatusType, error) {
	return notifications.TypeMessageStatusDelivered, nil
}

func (s *Service) Verify() (string, error) {
	verifyPath := "getMe"
	req, err := http.NewRequest("GET", s.Url+"/bot"+s.Token+"/"+verifyPath, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.Logger.Errorw("Cannot close body", "error", err)
		}
	}(resp.Body)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var verifyResponse GetMeResponse
	if err := json.Unmarshal(respBody, &verifyResponse); err != nil {
		s.Logger.Errorw("Cannot unmarshal response body", "error", err)
		return "", err
	}
	if !verifyResponse.OK {
		return "Bot Not Active", errors.New("bot not active")
	}
	return "Evertything good to go", nil
}

func createClient() *http.Client {
	headers := http.Header{
		"Content-Type": {"application/json"},
	}
	return &http.Client{
		Timeout: time.Second * 10,
		Transport: &customTransport{
			Base:   http.DefaultTransport,
			Header: headers,
		},
	}
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	return t.Base.RoundTrip(req)
}
