package asiatech

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/root-ali/iris/pkg/cache"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func NewService(username, password, scope, host, sender string, priority int, c cache.Interface[string, string],
	logger *zap.SugaredLogger) notifications.NotificationInterface {
	return &Service{
		host:     host,
		username: username,
		password: password,
		sender:   sender,
		scope:    scope,
		priority: priority,
		c:        c,
		logger:   logger,
	}
}

func (s *Service) GetName() string {
	return "Asiatech"
}

func (s *Service) GetFlag() string {
	return "sms"
}

func (s *Service) GetPriority() int {
	return s.priority
}

func (s *Service) Send(message notifications.Message) ([]string, error) {
	text := ""
	if message.State == "firing" {
		text = "🚨 Firing \n" + message.Subject + "\n" + message.Message + "\nTime: " + message.Time
	} else if message.State == "resolved" {
		text = "✅ Resolved \n" + message.Subject + "\n" + message.Message + "\nTime: " + message.Time
	}
	// Get asiatech token
	token, err := s.getAuthenticationToken()
	if err != nil {
		s.logger.Errorw("Failed to sent message with asiatech", "error", err)
		return nil, err
	}

	httpClient := s.createApiHandler(token)
	var msgBody []SendMessage
	for _, mobile := range message.Receptors {
		re := regexp.MustCompile("^0")
		mobile := re.ReplaceAllString(mobile, "98")
		msgBody = append(msgBody, SendMessage{
			Sender:   s.sender,
			Receptor: mobile,
			Text:     text,
			Coding:   0,
		})
	}

	resObj, err := json.Marshal(msgBody)
	if err != nil {
		s.logger.Errorw("cannot marshal request body", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.host+sendSms, bytes.NewReader(resObj))
	if err != nil {
		s.logger.Errorw("cannot create request to send message via asiatech",
			"receptor", message.Receptors)
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		s.logger.Errorw("cannot send message via asiatech provider",
			"receptor", message.Receptors,
			"message", message.Message,
			"error", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	msgResp := SendMessageResponse{}
	err = json.NewDecoder(resp.Body).Decode(&msgResp)
	if err != nil {
		s.logger.Errorw("cannot parse response from asiatech", "error", err)
		return nil, err
	}

	if msgResp.ResultCode != 100 {
		s.logger.Errorw("cannot send message via asiatech provider",
			"resultCode", msgResp.ResultCode)
		return nil, fmt.Errorf("cannot send message via asiatech provider: %d", msgResp.ResultCode)
	}

	return msgResp.Data, nil
}

func (s *Service) Status(messageID string) (notifications.MessageStatusType, error) {
	msgBody := DeliveryMessageRequest{messageID}
	s.logger.Infow("sending message to asiatech", "message", msgBody)
	token, err := s.getAuthenticationToken()
	if err != nil {
		s.logger.Errorw("Failed to send message to asiatech", "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	client := s.createApiHandler(token)

	jsonBody, err := json.Marshal(msgBody)
	if err != nil {
		s.logger.Errorw("cannot marshal request body", "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	req, err := http.NewRequest(http.MethodPost, s.host+deliveryStatus, bytes.NewReader(jsonBody))
	if err != nil {
		s.logger.Errorw("cannot create request to get delivery status via asiatech provider",
			"ID", messageID)
		return notifications.TypeMessageStatusFailed, err
	}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Errorw("cannot get delivery status from asiatech", "ID", messageID)
		return notifications.TypeMessageStatusFailed, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	msgResp := DeliveryMessageResponse{}
	err = json.NewDecoder(resp.Body).Decode(&msgResp)
	if err != nil {
		s.logger.Errorw("cannot parse response from asiatech", "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	msgCode := msgResp.Data[0].DeliveryStatus
	switch msgCode {
	case 1:
		s.logger.Infow("delivery status message received", "id", messageID)
		return notifications.TypeMessageStatusDelivered, nil
	case 5, 11, 13, 14, 15, 16:
		s.logger.Infow("delivery status message failed", "id", messageID)
		return notifications.TypeMessageStatusFailed, nil
	default:
		s.logger.Infow("message still is in progress...", "id", messageID)
		return notifications.TypeMessageStatusSent, nil
	}

}

func (s *Service) Verify() (string, error) {
	return "Asiatech service is up and running...", nil
}

func (s *Service) getAuthenticationToken() (string, error) {
	token, ok := s.c.Get("asiatech_token")
	if ok {
		return token, nil
	} else {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		form := url.Values{}
		form.Set("scope", s.scope)
		form.Set("username", s.username)
		form.Set("password", s.password)
		body := strings.NewReader(form.Encode())

		req, err := http.NewRequest(http.MethodPost, s.host+getTokenPath, body)
		if err != nil {
			s.logger.Error("Cannot create request for getting token in asiatech provider", "error", err)
			return "", err

		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		s.logger.Info("getting token from asiatech provider")
		resp, err := client.Do(req)
		if err != nil {
			s.logger.Errorw("cannot request to asiatech provider for getting token")
			return "", errors.New("cannot get token from asiatech")
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			s.logger.Errorw("getting not OK response from asiatech when trying to get token", "statusCode", resp.StatusCode)
			return "", errors.New("cannot get token from asiatech")
		}
		tknRsp := TokenResponse{}
		err = json.NewDecoder(resp.Body).Decode(&tknRsp)
		if err != nil {
			s.logger.Errorw("Cannot decode token response from asiatech provider", "error", err)
		}

		err = s.c.Set("asiatech_token", tknRsp.AccessToken, 4*time.Minute)
		if err != nil {
			return "", err
		}

		return tknRsp.AccessToken, nil
	}

}

func (s *Service) createApiHandler(token string) *http.Client {
	headers := http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + token},
		"scope":         {"ApiAccess"},
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &customTransport{
			Base:   http.DefaultTransport,
			Header: headers,
		},
	}

	return client
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	return t.Base.RoundTrip(req)
}
