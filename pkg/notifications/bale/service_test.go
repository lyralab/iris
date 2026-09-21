package bale

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/root-ali/iris/pkg/notifications"
	"go.uber.org/zap"
)

func getLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

func TestGetName(t *testing.T) {
	service := NewBaleService("http://localhost", "test-token", 10, getLogger())
	if service.GetName() != "Bale" {
		t.Errorf("Expected name 'Bale', got %s", service.GetName())
	}
}

func TestGetFlag(t *testing.T) {
	service := NewBaleService("http://localhost", "test-token", 10, getLogger())
	if service.GetFlag() != "bale" {
		t.Errorf("Expected flag 'bale', got %s", service.GetFlag())
	}
}

func TestGetPriority(t *testing.T) {
	priority := 25
	service := NewBaleService("http://localhost", "test-token", priority, getLogger())
	if service.GetPriority() != priority {
		t.Errorf("Expected priority %d, got %d", priority, service.GetPriority())
	}
}

func TestStatus(t *testing.T) {
	service := NewBaleService("http://localhost", "test-token", 10, getLogger())
	status, err := service.Status("123")
	if err != nil {
		t.Errorf("Status returned error: %v", err)
	}
	if status != notifications.TypeMessageStatusDelivered {
		t.Errorf("Expected TypeMessageStatusDelivered, got %v", status)
	}
}

func TestVerifySuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bottest-token/getMe" {
			t.Errorf("Expected path /bottest-token/getMe, got %s", r.URL.Path)
		}
		response := GetMeResponse{OK: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	result, err := service.Verify()
	if err != nil {
		t.Errorf("Verify returned error: %v", err)
	}
	if result != "Evertything good to go" {
		t.Errorf("Expected 'Evertything good to go', got %s", result)
	}
}

func TestVerifyFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GetMeResponse{OK: false}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	result, err := service.Verify()
	if err == nil {
		t.Error("Expected Verify to return error for failed response")
	}
	if result != "Bot Not Active" {
		t.Errorf("Expected 'Bot Not Active', got %s", result)
	}
}

func TestSendSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/bottest-token/sendMessage" {
			t.Errorf("Expected path /bottest-token/sendMessage, got %s", r.URL.Path)
		}

		var sendMsg SendMessage
		if err := json.NewDecoder(r.Body).Decode(&sendMsg); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		response := SendMessageResponse{OK: true, Message: "Message sent"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	message := notifications.Message{
		State:     "firing",
		Subject:   "Alert Title",
		Message:   "Alert Description",
		Time:      "2025-08-02 09:55:06",
		Receptors: []string{"123456789"},
	}

	responses, err := service.Send(message)
	if err != nil {
		t.Errorf("Send returned error: %v", err)
	}
	if len(responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responses))
	}
}

func TestSendResolved(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sendMsg SendMessage
		json.NewDecoder(r.Body).Decode(&sendMsg)

		if sendMsg.ChatID != 123456789 {
			t.Errorf("Expected ChatID 123456789, got %d", sendMsg.ChatID)
		}

		response := SendMessageResponse{OK: true, Message: "Message sent"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	message := notifications.Message{
		State:     "resolved",
		Subject:   "Alert Title",
		Message:   "Alert Description",
		Time:      "2025-08-02 09:55:06",
		Receptors: []string{"123456789"},
	}

	responses, err := service.Send(message)
	if err != nil {
		t.Errorf("Send returned error: %v", err)
	}
	if len(responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responses))
	}
}

func TestSendMultipleReceptors(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		response := SendMessageResponse{OK: true, Message: "Message sent"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	message := notifications.Message{
		State:     "firing",
		Subject:   "Alert Title",
		Message:   "Alert Description",
		Time:      "2025-08-02 09:55:06",
		Receptors: []string{"123456789", "987654321", "555555555"},
	}

	responses, err := service.Send(message)
	if err != nil {
		t.Errorf("Send returned error: %v", err)
	}
	if len(responses) != 3 {
		t.Errorf("Expected 3 responses, got %d", len(responses))
	}
	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestSendInvalidReceptor(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := SendMessageResponse{OK: true, Message: "Message sent"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	message := notifications.Message{
		State:     "firing",
		Subject:   "Alert Title",
		Message:   "Alert Description",
		Time:      "2025-08-02 09:55:06",
		Receptors: []string{"not-a-number", "123456789"},
	}

	responses, err := service.Send(message)
	if err == nil {
		t.Error("Expected error for invalid receptor")
	}
	if len(responses) != 1 {
		t.Errorf("Expected 1 successful response, got %d", len(responses))
	}
}

func TestSendServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	message := notifications.Message{
		State:     "firing",
		Subject:   "Alert Title",
		Message:   "Alert Description",
		Time:      "2025-08-02 09:55:06",
		Receptors: []string{"123456789"},
	}

	responses, err := service.Send(message)
	if err == nil {
		t.Error("Expected error for server error response")
	}
	if len(responses) != 0 {
		t.Errorf("Expected 0 successful responses, got %d", len(responses))
	}
}

func TestSendAPIFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := SendMessageResponse{OK: false, Message: "Invalid chat ID"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	service := NewBaleService(mockServer.URL, "test-token", 10, getLogger())
	message := notifications.Message{
		State:     "firing",
		Subject:   "Alert Title",
		Message:   "Alert Description",
		Time:      "2025-08-02 09:55:06",
		Receptors: []string{"123456789"},
	}

	_, err := service.Send(message)
	if err == nil {
		t.Error("Expected error for API failure response")
	}
}

func TestRealServerVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real server test in short mode")
	}

	url := os.Getenv("BALE_URL")
	token := os.Getenv("BALE_TOKEN")

	if url == "" || token == "" {
		t.Skip("BALE_URL or BALE_TOKEN not set, skipping real server test")
	}

	service := NewBaleService(url, token, 10, getLogger())
	result, err := service.Verify()
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	t.Logf("Verify result: %s", result)
}
