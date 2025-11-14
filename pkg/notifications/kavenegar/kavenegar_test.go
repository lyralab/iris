package kavenegar

import (
	"os"
	"testing"

	"github.com/root-ali/iris/internal/logging"
	notifications "github.com/root-ali/iris/pkg/notifications"
)

func TestSend(t *testing.T) {
	zl, err := logging.New("Debug")
	if err != nil {
		t.Fatal(err)
	}
	logger := zl.Sugar()
	kv := NewKavenegarService(os.Getenv("KAVENEGAR_API_TOKEN"),
		1, os.Getenv("KAVENEGAR_SENDER"), logger)

	message := notifications.Message{
		Subject: "Test Subject",
		Message: "Test Message",
		Receptors: []string{
			"09305243718",
		},
	}
	_, err = kv.Send(message)
	if err != nil {
		t.Fatal(err)
	}
}
