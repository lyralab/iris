package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/alerts"
)

type silenceAlertService struct {
	alerts.Service
	called bool
	id, by string
	d      time.Duration
	err    error
}

func (s *silenceAlertService) SilenceAlert(id, by string, d time.Duration) (*alerts.Alert, error) {
	s.called, s.id, s.by, s.d = true, id, by, d
	if s.err != nil {
		return nil, s.err
	}
	var until *time.Time
	if d > 0 {
		t := time.Now().Add(d)
		until = &t
	}
	return &alerts.Alert{Id: id, Status: "firing", Silenced: true, SilencedUntil: until, SilencedBy: by}, nil
}

func postSilence(t *testing.T, svc alerts.Service, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v0/alerts/:id/silence", func(c *gin.Context) { c.Set("username", "boss") }, SilenceAlert(svc))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v0/alerts/a1/silence", strings.NewReader(body)))
	return w
}

func TestSilenceAlertHandler(t *testing.T) {
	cases := []struct {
		name, body string
		wantCode   int
		wantD      time.Duration
		wantUntil  bool
	}{
		{"timed", `{"duration":"4h"}`, http.StatusOK, 4 * time.Hour, true},
		{"empty body is indefinite", ``, http.StatusOK, 0, false},
		{"empty duration is indefinite", `{}`, http.StatusOK, 0, false},
		{"bad duration", `{"duration":"soon"}`, http.StatusBadRequest, 0, false},
		{"zero duration", `{"duration":"0s"}`, http.StatusBadRequest, 0, false},
		{"bad json", `{`, http.StatusBadRequest, 0, false},
	}
	for _, c := range cases {
		svc := &silenceAlertService{}
		w := postSilence(t, svc, c.body)
		if w.Code != c.wantCode {
			t.Fatalf("%s: code = %d, want %d (%s)", c.name, w.Code, c.wantCode, w.Body)
		}
		if c.wantCode != http.StatusOK {
			if svc.called {
				t.Fatalf("%s: service called on a bad request", c.name)
			}
			continue
		}
		if svc.id != "a1" || svc.by != "boss" || svc.d != c.wantD {
			t.Fatalf("%s: service got id=%q by=%q d=%v", c.name, svc.id, svc.by, svc.d)
		}
		var resp struct{ Alert AlertResponse }
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if !resp.Alert.Silenced || resp.Alert.SilencedBy != "boss" || (resp.Alert.SilencedUntil != "") != c.wantUntil {
			t.Fatalf("%s: response alert = %+v", c.name, resp.Alert)
		}
	}
}

func TestSilenceAlertHandlerMapsErrors(t *testing.T) {
	for err, code := range map[error]int{
		apperr.ErrAlertNotFiring: http.StatusConflict,
		apperr.ErrAlertNotFound:  http.StatusNotFound,
	} {
		if w := postSilence(t, &silenceAlertService{err: err}, `{"duration":"1h"}`); w.Code != code {
			t.Errorf("%v: code = %d, want %d", err, w.Code, code)
		}
	}
}

func TestAlertResponseHidesExpiredSilence(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	got := toAlertResponse([]*alerts.Alert{{Silenced: true, SilencedUntil: &past, SilencedBy: "boss"}})[0]
	if got.Silenced || got.SilencedBy != "" || got.SilencedUntil != "" {
		t.Fatalf("expired silence shown: %+v", got)
	}
}

type listAlertService struct {
	alerts.Service
	called       bool
	silencedOnly bool
}

func (s *listAlertService) GetAlerts(_, _ string, silencedOnly bool, _, _ int) ([]*alerts.Alert, error) {
	s.called, s.silencedOnly = true, silencedOnly
	return nil, nil
}

func TestGetAlertsSilencedParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for query, want := range map[string]struct {
		code     int
		silenced bool
	}{
		"":               {http.StatusOK, false},
		"?silenced=true": {http.StatusOK, true},
		"?silenced=0":    {http.StatusOK, false},
		"?silenced=yes":  {http.StatusBadRequest, false},
	} {
		svc := &listAlertService{}
		r := gin.New()
		r.GET("/v0/alerts", GetAlerts(svc))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v0/alerts"+query, nil))
		if w.Code != want.code || svc.silencedOnly != want.silenced || svc.called != (want.code == http.StatusOK) {
			t.Errorf("%q: code=%d silencedOnly=%v called=%v, want %+v", query, w.Code, svc.silencedOnly, svc.called, want)
		}
	}
}
