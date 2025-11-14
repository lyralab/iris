package smsir

import (
	"os"
	"testing"

	"github.com/root-ali/iris/internal/logging"
	"github.com/root-ali/iris/pkg/notifications"
)

func TestSend(t *testing.T) {
	zl, err := logging.New("Debug")
	if err != nil {
		t.Fatal(err)
	}
	logger := zl.Sugar()

	smsir := NewSmsirService(os.Getenv("SMSIR_API_TOKEN"),
		os.Getenv("SMSIR_LINE_NUMBER"), 2, logger)

	message := notifications.Message{
		Subject: "Test Subject",
		Message: "Test Message",
		Receptors: []string{
			"09305243718",
		},
	}

	_, err = smsir.Send(message)
	if err != nil {
		t.Fatal(err)
	}
}
