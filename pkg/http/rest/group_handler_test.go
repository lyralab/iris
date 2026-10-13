package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/groups"
	"go.uber.org/zap"
)

func getGroup(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := &membershipGroups{byID: map[string]*groups.Group{"abc": {ID: "abc", Name: "ops"}}}
	r := gin.New()
	r.GET("/v0/groups/:group_id", GetGroupHandler(svc, zap.NewNop().Sugar()))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestGetGroupByPathID(t *testing.T) {
	w := getGroup(t, "/v0/groups/abc")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var body struct {
		Data groups.Group `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.ID != "abc" {
		t.Fatalf("id %q, want abc", body.Data.ID)
	}
}

func TestGetGroupMissingIDNotFound(t *testing.T) {
	w := getGroup(t, "/v0/groups/nope")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}
