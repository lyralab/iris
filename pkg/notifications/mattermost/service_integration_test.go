package mattermost

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

// TestSendToReceptor sends one alert through the real Mattermost server, the same way the alert
// scheduler does, and checks that the post reached the channel stored in a user's mattermost_id.
//
// It is skipped unless the server details are set:
//
//	MATTERMOST_URL=https://chat.jibit.cloud \
//	MATTERMOST_BOT_TOKEN=... \
//	MATTERMOST_TEST_RECEPTOR=<value of users.mattermost_id> \
//	go test ./pkg/notifications/mattermost/ -run TestSendToReceptor -v -count=1
//
// The Go client honours HTTPS_PROXY / NO_PROXY, so run it with the same proxy environment as the server
// to reproduce a proxy that cannot reach Mattermost.
func TestSendToReceptor(t *testing.T) {
	url, token, receptor := os.Getenv("MATTERMOST_URL"), os.Getenv("MATTERMOST_BOT_TOKEN"), os.Getenv("MATTERMOST_TEST_RECEPTOR")
	if url == "" || token == "" || receptor == "" {
		t.Skip("set MATTERMOST_URL, MATTERMOST_BOT_TOKEN and MATTERMOST_TEST_RECEPTOR to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := model.NewAPIv4Client(url)
	client.SetToken(token)

	// Each check names one cause, so a failure says which setting is wrong.
	me, _, err := client.GetMe(ctx, "")
	if err != nil {
		t.Fatalf("bot token rejected or server unreachable (check proxy / NO_PROXY): %v", err)
	}
	if _, _, err := client.GetChannel(ctx, receptor); err != nil {
		if _, _, uerr := client.GetUser(ctx, receptor, ""); uerr == nil {
			t.Fatalf("%s is a user id; mattermost_id must be a channel id, e.g. the DM channel between %s and that user", receptor, me.Username)
		}
		t.Fatalf("bot %s cannot access channel %s (not a member, or not a channel id): %v", me.Username, receptor, err)
	}

	subject := "iris integration test " + time.Now().Format(time.RFC3339Nano)
	svc := NewService(Config{Url: url, BotToken: token}, zap.NewNop().Sugar())
	if _, err := svc.Send(notifications.Message{
		Subject:   subject,
		Message:   "TestSendToReceptor",
		State:     "firing",
		Time:      time.Now().Format(time.DateTime),
		Receptors: []string{receptor},
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// ponytail: Send drops CreatePost errors today, so confirm delivery by reading the channel back.
	posts, _, err := client.GetPostsForChannel(ctx, receptor, 0, 5, "", false, false)
	if err != nil {
		t.Fatalf("read channel %s: %v", receptor, err)
	}
	for _, p := range posts.Posts {
		if p.UserId == me.Id && p.Message == "🚨 "+subject+" 🚨" {
			return
		}
	}
	t.Fatalf("Send returned no error but no post %q from %s is in channel %s", subject, me.Username, receptor)
}
