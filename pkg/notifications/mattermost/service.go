package mattermost

import (
	"context"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/root-ali/iris/pkg/notifications"
)

func (s service) Send(message notifications.Message) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	results := make([]string, 0, len(message.Receptors))
	errStack := make(errorStack, 0)
	for _, recipient := range message.Receptors {
		post := &model.Post{
			ChannelId: recipient,
			Message:   message.Message,
		}

		_, r, err := s.client.CreatePost(ctx, post)
		if err != nil {
			errStack.Append(err)
			results = append(results, recipient)
			continue
		}

		results = append(results, r.RequestId)
	}
	if len(errStack) > 0 {
		return results, errStack
	}
	return results, nil
}

func (s service) Status(_ string) (notifications.MessageStatusType, error) {
	return notifications.TypeMessageStatusDelivered, nil
}

func (s service) Verify() (string, error) {
	return "Service is operational", nil
}

func (s service) GetName() string {
	return "Mattermost"
}

func (s service) GetFlag() string {
	return "mattermost"
}

func (s service) GetPriority() int {
	return s.priority
}
