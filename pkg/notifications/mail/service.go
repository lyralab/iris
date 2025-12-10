package mail

import (
	"bytes"
	"html/template"
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"github.com/wneessen/go-mail"
)

func (s *service) Send(message notifications.Message) ([]string, error) {
	msg := mail.NewMsg()

	messageSubject := message.Subject
	timeParse, err := time.Parse(time.DateTime, message.Time)
	als := AlertNotification{
		State:     message.State,
		Title:     message.Subject,
		Message:   message.Message,
		Timestamp: timeParse,
	}
	messageBody, err := generateEmail(als)
	if err != nil {
		s.logger.Errorf("failed to generate email body: %v", err)
		return nil, err
	}
	if err := msg.From(s.fromAddr); err != nil {
		s.logger.Errorf("failed to set from address: %v", err)
		return nil, err
	}
	if err := msg.To(message.Receptors...); err != nil {
		s.logger.Errorf("failed to set subject: %v", err)
		return nil, err
	}

	msg.Subject(messageSubject)
	msg.SetBodyString("text/html", messageBody)

	if err := s.client.DialAndSend(msg); err != nil {
		s.logger.Errorf("failed to send email: %v", err)
		return nil, err
	}
	s.logger.Infof("email sent successfully to: %v", message.Receptors)

	return message.Receptors, nil
}
func (s *service) Status(messageID string) (notifications.MessageStatusType, error) {
	return notifications.TypeMessageStatusDelivered, nil
}
func (s *service) Verify() (string, error) {
	s.logger.Info("Mail service verification is not implemented")
	return "", nil
}

func (s *service) GetName() string {
	return s.name
}

func (s *service) GetFlag() string {
	return "mail"
}

func (s *service) GetPriority() int {
	return s.priority
}

func generateEmail(als AlertNotification) (string, error) {
	tmpl, err := template.New("iris").Parse(irisTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, als)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
