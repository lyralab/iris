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
	"strconv"
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
	if u, err := url.Parse(s.host); err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("asiatech host %q must include a scheme, e.g. https://smsapi.asiatech.ir", s.host)
	}
	// Get asiatech token
	token, err := s.getAuthenticationToken()
	if err != nil {
		s.logger.Errorw("Asiatech: cannot send SMS, authentication failed", "host", s.host, "error", err)
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
		s.logger.Errorw("Asiatech: cannot build SMS request body", "error", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.host+sendSms, bytes.NewReader(resObj))
	if err != nil {
		s.logger.Errorw("Asiatech: cannot create SMS request",
			"receptors", message.Receptors, "error", err)
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		s.logger.Errorw("Asiatech: SMS request failed",
			"host", s.host, "receptors", message.Receptors, "error", err)
		return nil, err
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
		s.logger.Errorw("Asiatech: cannot parse SMS response",
			"httpStatus", resp.StatusCode, "error", err)
		return nil, fmt.Errorf("cannot parse asiatech response (http %d): %w", resp.StatusCode, err)
	}

	if msgResp.ResultCode != 100 {
		s.logger.Errorw("Asiatech: SMS rejected by provider",
			"resultCode", msgResp.ResultCode, "receptors", message.Receptors)
		return nil, fmt.Errorf("cannot send message via asiatech provider: %d", msgResp.ResultCode)
	}

	// A negative id (e.g. -11) is a per-receptor rejection: nothing was sent to that receptor.
	// Report it as "r1;r2;" with "nil" for accepted receptors, the shape the alert scheduler reads.
	results := make([]string, len(msgResp.Data))
	rejected := 0
	for i, id := range msgResp.Data {
		results[i] = "nil"
		if n, err := strconv.Atoi(id); err == nil && n < 0 {
			results[i] = "asiatech rejected receptor with code " + id
			rejected++
			s.logger.Errorw("Asiatech: receptor rejected by provider, SMS not sent",
				"code", id, "receptor", receptorAt(message.Receptors, i))
		}
	}
	if rejected > 0 {
		return msgResp.Data, errors.New(strings.Join(results, ";") + ";")
	}
	s.logger.Infow("Asiatech: SMS accepted", "receptors", message.Receptors, "messageIDs", msgResp.Data)

	return msgResp.Data, nil
}

func (s *Service) Status(messageID string) (notifications.MessageStatusType, error) {
	msgBody := DeliveryMessageRequest{messageID}
	s.logger.Infow("Asiatech: checking delivery status", "messageID", messageID)
	token, err := s.getAuthenticationToken()
	if err != nil {
		s.logger.Errorw("Asiatech: cannot check delivery status, authentication failed", "messageID", messageID, "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	client := s.createApiHandler(token)

	jsonBody, err := json.Marshal(msgBody)
	if err != nil {
		s.logger.Errorw("Asiatech: cannot build delivery status request", "messageID", messageID, "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	req, err := http.NewRequest(http.MethodPost, s.host+deliveryStatus, bytes.NewReader(jsonBody))
	if err != nil {
		s.logger.Errorw("Asiatech: cannot create delivery status request", "messageID", messageID, "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Errorw("Asiatech: delivery status request failed", "messageID", messageID, "host", s.host, "error", err)
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
		s.logger.Errorw("Asiatech: cannot parse delivery status response", "messageID", messageID, "error", err)
		return notifications.TypeMessageStatusFailed, err
	}
	msgCode := msgResp.Data[0].DeliveryStatus
	switch msgCode {
	case 1:
		s.logger.Infow("Asiatech: message delivered", "messageID", messageID)
		return notifications.TypeMessageStatusDelivered, nil
	case 5, 11, 13, 14, 15, 16:
		s.logger.Warnw("Asiatech: message delivery failed", "messageID", messageID, "deliveryStatus", msgCode)
		return notifications.TypeMessageStatusFailed, nil
	default:
		s.logger.Infow("Asiatech: message still in progress", "messageID", messageID, "deliveryStatus", msgCode)
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
			s.logger.Errorw("Asiatech: cannot create token request", "host", s.host, "error", err)
			return "", err

		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		s.logger.Infow("Asiatech: requesting new access token", "host", s.host)
		resp, err := client.Do(req)
		if err != nil {
			s.logger.Errorw("Asiatech: token request failed", "host", s.host, "error", err)
			return "", fmt.Errorf("cannot get token from asiatech: %w", err)
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			s.logger.Errorw("Asiatech: token request rejected (check username, password and scope)",
				"host", s.host, "statusCode", resp.StatusCode)
			return "", fmt.Errorf("cannot get token from asiatech: http %d", resp.StatusCode)
		}
		tknRsp := TokenResponse{}
		err = json.NewDecoder(resp.Body).Decode(&tknRsp)
		if err != nil {
			s.logger.Errorw("Asiatech: cannot decode token response", "error", err)
			return "", fmt.Errorf("cannot get token from asiatech: invalid response: %w", err)
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

func receptorAt(receptors []string, i int) string {
	if i < len(receptors) {
		return receptors[i]
	}
	return ""
}
