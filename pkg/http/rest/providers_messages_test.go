package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/message"
)

type fakeMessageLister struct {
	sender string
	limit  int
}

func (f *fakeMessageLister) ListBySender(sender string, limit int) ([]message.Sent, error) {
	f.sender, f.limit = sender, limit
	return []message.Sent{}, nil
}

func TestGetProviderMessagesLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := map[string]int{
		"":           defaultMessagesLimit,
		"?limit=25":  25,
		"?limit=0":   defaultMessagesLimit,
		"?limit=-3":  defaultMessagesLimit,
		"?limit=abc": defaultMessagesLimit,
		"?limit=500": maxMessagesLimit,
	}
	for query, want := range cases {
		fake := &fakeMessageLister{}
		r := gin.New()
		r.GET("/v0/providers/:name/messages", GetProviderMessagesHandler(fake))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v0/providers/kavenegar/messages"+query, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%q: status %d body %s", query, w.Code, w.Body.String())
		}
		if fake.sender != "kavenegar" || fake.limit != want {
			t.Fatalf("%q: got sender %q limit %d, want kavenegar %d", query, fake.sender, fake.limit, want)
		}
	}
}
