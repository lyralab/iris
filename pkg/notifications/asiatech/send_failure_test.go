package asiatech

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Part of the SMS fallback issue (#2): a failing Asiatech must report an error so the scheduler can move
// on to the next sms provider. It must not panic.
func TestSend_TransportErrorReturnsError(t *testing.T) {
	mockCache := &MockCache{}
	mockCache.On("Get", "asiatech_token").Return("cached_token", true)
	// Nothing listens on this port, so the send request fails before any response exists.
	service := &Service{
		host:   "http://127.0.0.1:1",
		sender: "9890000476",
		c:      mockCache,
		logger: zap.NewNop().Sugar(),
	}

	var err error
	assert.NotPanics(t, func() {
		_, err = service.Send(notifications.Message{
			State: "firing", Subject: "s", Message: "m", Time: time.Now().Format(time.DateTime),
			Receptors: []string{"09120000000"},
		})
	})
	assert.Error(t, err)
}

// The configured host has no scheme ("smsapi.asiatech.ir"); every call fails with "unsupported protocol scheme".
// Fails today with a token error that hides the real cause, so the message must name the host problem.
func TestSend_HostWithoutSchemeErrorNamesTheHost(t *testing.T) {
	mockCache := &MockCache{}
	mockCache.On("Get", "asiatech_token").Return("", false)
	service := &Service{
		host:   "smsapi.asiatech.ir",
		c:      mockCache,
		logger: zap.NewNop().Sugar(),
	}

	_, err := service.Send(notifications.Message{State: "firing", Receptors: []string{"09120000000"}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")
}

// Asiatech answers resultCode 100 but puts a negative code (e.g. "-11") in data for a receptor it rejected.
// Such a receptor was not sent anything, so Send must report it as failed.
func asiatechReturning(t *testing.T, data string) *Service {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"resultCode":100,"succeeded":true,"data":` + data + `}`))
	}))
	t.Cleanup(server.Close)
	mockCache := &MockCache{}
	mockCache.On("Get", "asiatech_token").Return("cached_token", true)
	return &Service{host: server.URL, sender: "9890000476", c: mockCache, logger: zap.NewNop().Sugar()}
}

func TestSend_RejectedReceptorCodeReturnsError(t *testing.T) {
	service := asiatechReturning(t, `["-11"]`)

	ids, err := service.Send(notifications.Message{State: "firing", Receptors: []string{"09120000000"}})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "-11")
	assert.Len(t, ids, 1)
}

func TestSend_PartiallyRejectedReceptorsReportPerRecipient(t *testing.T) {
	service := asiatechReturning(t, `["12345","-11"]`)

	ids, err := service.Send(notifications.Message{State: "firing", Receptors: []string{"09120000000", "09130000000"}})

	// one token per receptor in send order, "nil" = accepted: the shape the alert scheduler parses.
	assert.EqualError(t, err, "nil;asiatech rejected receptor with code -11;")
	assert.Equal(t, []string{"12345", "-11"}, ids)
}

func TestSend_AcceptedReceptorsReturnIDs(t *testing.T) {
	service := asiatechReturning(t, `["12345"]`)

	ids, err := service.Send(notifications.Message{State: "firing", Receptors: []string{"09120000000"}})

	assert.NoError(t, err)
	assert.Equal(t, []string{"12345"}, ids)
}
