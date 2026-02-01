package asiatech

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/root-ali/iris/pkg/notifications"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := NewService("testUser",
		"testUser",
		"testscope",
		"http://asiatech.sms",
		"testsender",
		1,
		mockCache,
		logger)

	assert.NotNil(t, service)
	assert.Equal(t, "Asiatech", service.GetName())
	assert.Equal(t, "sms", service.GetFlag())
	assert.Equal(t, 1, service.GetPriority())
}

func TestGetAuthenticationToken_FromCache(t *testing.T) {
	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     "http://test.com",
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("cached_token", true)

	token, err := service.getAuthenticationToken()

	assert.NoError(t, err)
	assert.Equal(t, "cached_token", token)
	mockCache.AssertExpectations(t)
}

func TestGetAuthenticationToken_FromAPI_MockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/connect/token" {
			w.Header().Set("Content-Type", "application/json")
			response := TokenResponse{
				AccessToken: "new_token",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("", false)
	mockCache.On("Set", "asiatech_token", "new_token", 4*time.Minute).Return(nil)

	token, err := service.getAuthenticationToken()

	assert.NoError(t, err)
	assert.Equal(t, "new_token", token)
	mockCache.AssertExpectations(t)
}

func TestGetAuthenticationToken_APIError_MockServer(t *testing.T) {
	// Create mock server that returns 400
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/connect/token" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("", false)

	token, err := service.getAuthenticationToken()

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "cannot get token from asiatech")
	mockCache.AssertExpectations(t)
}

func TestGetAuthenticationToken_CacheSetError_MockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/connect/token" {
			w.Header().Set("Content-Type", "application/json")
			response := TokenResponse{
				AccessToken: "new_token",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("", false)
	mockCache.On("Set", "asiatech_token", "new_token", 4*time.Minute).Return(assert.AnError)

	token, err := service.getAuthenticationToken()

	assert.Error(t, err)
	assert.Empty(t, token)
	mockCache.AssertExpectations(t)
}

func TestSend_FiringAlert_MockServer(t *testing.T) {
	// Create mock server for token and send message endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			w.Header().Set("Content-Type", "application/json")
			response := TokenResponse{AccessToken: "test_token"}
			json.NewEncoder(w).Encode(response)
		case "/api/1/message/send":
			w.Header().Set("Content-Type", "application/json")
			response := SendMessageResponse{
				ResultCode: 100,
				Data:       []string{"msg_id_1"},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		sender:   "testsender",
		priority: 1,
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("", false)
	mockCache.On("Set", "asiatech_token", "test_token", 4*time.Minute).Return(nil)

	message := notifications.Message{
		State:     "firing",
		Subject:   "Test Alert",
		Message:   "Test message content",
		Time:      "2023-12-14 10:00:00",
		Receptors: []string{"09123456789"},
	}

	messageIDs, err := service.Send(message)

	assert.NoError(t, err)
	assert.Equal(t, []string{"msg_id_1"}, messageIDs)
	mockCache.AssertExpectations(t)
}

func TestSend_ResolvedAlert_MockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			w.Header().Set("Content-Type", "application/json")
			response := TokenResponse{AccessToken: "test_token"}
			json.NewEncoder(w).Encode(response)
		case "/api/1/message/send":
			// Verify the message format for resolved alert
			body, _ := io.ReadAll(r.Body)
			var msgBody SendMessage
			json.Unmarshal(body, &msgBody)

			assert.Contains(t, msgBody.Text, "✅ Resolved")

			w.Header().Set("Content-Type", "application/json")
			response := SendMessageResponse{
				ResultCode: 100,
				Data:       []string{"msg_id_1"},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		sender:   "testsender",
		priority: 1,
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("test_token", true)

	message := notifications.Message{
		State:     "resolved",
		Subject:   "Test Alert",
		Message:   "Test message content",
		Time:      "2023-12-14 10:00:00",
		Receptors: []string{"09123456789"},
	}

	messageIDs, err := service.Send(message)

	assert.NoError(t, err)
	assert.Equal(t, []string{"msg_id_1"}, messageIDs)
	mockCache.AssertExpectations(t)
}

func TestSend_MessageFormatting_MockServer(t *testing.T) {
	tests := []struct {
		name         string
		state        string
		expectedText string
	}{
		{
			name:         "Firing Alert",
			state:        "firing",
			expectedText: "🚨 Firing",
		},
		{
			name:         "Resolved Alert",
			state:        "resolved",
			expectedText: "✅ Resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/connect/token":
					w.Header().Set("Content-Type", "application/json")
					response := TokenResponse{AccessToken: "test_token"}
					json.NewEncoder(w).Encode(response)
				case "/api/1/message/send":
					// Verify the message format
					body, _ := io.ReadAll(r.Body)
					var msgBody SendMessage
					json.Unmarshal(body, &msgBody)

					assert.Contains(t, msgBody.Text, tt.expectedText)
					assert.Contains(t, msgBody.Text, "Test Alert")
					assert.Contains(t, msgBody.Text, "Test message content")
					assert.Contains(t, msgBody.Text, "Time: 2023-12-14 10:00:00")
					assert.Equal(t, "testsender", msgBody.Sender)
					assert.Equal(t, []string{"09123456789"}, msgBody.Receptor)

					w.Header().Set("Content-Type", "application/json")
					response := SendMessageResponse{
						ResultCode: 100,
						Data:       []string{"msg_id_1"},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			mockCache := &MockCache{}
			logger := zap.NewNop().Sugar()

			service := &Service{
				host:     server.URL,
				username: "testuser",
				password: "testpass",
				scope:    "testscope",
				sender:   "testsender",
				priority: 1,
				c:        mockCache,
				logger:   logger,
			}

			mockCache.On("Get", "asiatech_token").Return("test_token", true)

			message := notifications.Message{
				State:     tt.state,
				Subject:   "Test Alert",
				Message:   "Test message content",
				Time:      "2023-12-14 10:00:00",
				Receptors: []string{"09123456789"},
			}

			messageIDs, err := service.Send(message)

			assert.NoError(t, err)
			assert.Equal(t, []string{"msg_id_1"}, messageIDs)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestSend_APIError_MockServer(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/connect/token":
			w.Header().Set("Content-Type", "application/json")
			response := TokenResponse{AccessToken: "test_token"}
			json.NewEncoder(w).Encode(response)
		case "/api/1/message/send":
			w.Header().Set("Content-Type", "application/json")
			response := SendMessageResponse{
				ResultCode: 400, // Error code
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		sender:   "testsender",
		priority: 1,
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("test_token", true)

	message := notifications.Message{
		State:     "firing",
		Subject:   "Test Alert",
		Message:   "Test message content",
		Time:      "2023-12-14 10:00:00",
		Receptors: []string{"09123456789"},
	}

	_, err := service.Send(message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot send message via asiatech provider: 400")
	mockCache.AssertExpectations(t)
}

func TestSend_AuthTokenError_MockServer(t *testing.T) {
	// Create mock server that returns auth error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/connect/token" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	logger := zap.NewNop().Sugar()

	service := &Service{
		host:     server.URL,
		username: "testuser",
		password: "testpass",
		scope:    "testscope",
		sender:   "testsender",
		priority: 1,
		c:        mockCache,
		logger:   logger,
	}

	mockCache.On("Get", "asiatech_token").Return("", false)

	message := notifications.Message{
		State:     "firing",
		Subject:   "Test Alert",
		Message:   "Test message content",
		Time:      "2023-12-14 10:00:00",
		Receptors: []string{"09123456789"},
	}

	_, err := service.Send(message)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot get token from asiatech")
	mockCache.AssertExpectations(t)
}

func TestGetStatus_FromAPI_GetRejectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/connect/token" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			response := TokenResponse{
				AccessToken: "test_token",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		if r.URL.Path == deliveryStatus {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			response := DeliveryMessageResponse{
				ResultCode: 100,
				Data: []struct {
					ID             string `json:"id"`
					DeliveryStatus int    `json:"deliveryStatus"`
					DeliveryTime   string `json:"deliveryDate"`
				}{
					{
						ID:             "message_ids",
						DeliveryStatus: 5,
						DeliveryTime:   "2023-12-14 12:00:00",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}))
	defer server.Close()

	mockCache := &MockCache{}
	mockCache.On("Get", "asiatech_token").Return("", false)
	mockCache.On("Set", "asiatech_token", "test_token", 4*time.Minute).Return(nil)
	logger := zap.NewNop().Sugar()

	service := &Service{
		server.URL,
		"testUser",
		"testPassword",
		"98900090",
		"scope",
		1,
		mockCache,
		logger,
	}

	status, err := service.Status("message_ids")
	if err != nil {
		t.Errorf("Failed to get status: %v", err)
		t.FailNow()
	}
	t.Log("message status:", status)
}

func TestStatus(t *testing.T) {
	service := &Service{}

	status, err := service.Status("test_message_id")

	assert.NoError(t, err)
	assert.Equal(t, notifications.TypeMessageStatusDelivered, status)
}

func TestVerify(t *testing.T) {
	service := &Service{}

	message, err := service.Verify()

	assert.NoError(t, err)
	assert.Equal(t, "Asiatech service is up and running...", message)
}

func TestCreateApiHandler(t *testing.T) {
	service := &Service{}

	client := service.createApiHandler("test_token")

	assert.NotNil(t, client)
	assert.Equal(t, 10*time.Second, client.Timeout)
}

func TestCustomTransport_RoundTrip(t *testing.T) {
	// Create a test server to capture the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &customTransport{
		Base: http.DefaultTransport,
		Header: http.Header{
			"Content-Type":  {"application/json"},
			"Authorization": {"Bearer test_token"},
		},
	}

	req, _ := http.NewRequest("GET", server.URL, bytes.NewReader([]byte{}))

	resp, err := transport.RoundTrip(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
