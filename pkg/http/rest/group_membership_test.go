package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/internal/apperr"
	"github.com/root-ali/iris/pkg/groups"
	"go.uber.org/zap"
)

type membershipGroups struct {
	addedUser   string
	removedUser string
	byID        map[string]*groups.Group
}

func (m *membershipGroups) CreateGroup(*groups.Group) error        { return nil }
func (m *membershipGroups) GetGroup(string) (*groups.Group, error) { return nil, nil }
func (m *membershipGroups) GetGroupById(id string) (*groups.Group, error) {
	if g, ok := m.byID[id]; ok {
		return g, nil
	}
	return nil, apperr.ErrGroupNotFound
}
func (m *membershipGroups) GetAllGroups() ([]*groups.Group, error)    { return nil, nil }
func (m *membershipGroups) DeleteGroup(*groups.Group) error           { return nil }
func (m *membershipGroups) ListUsers(*groups.Group) ([]string, error) { return nil, nil }
func (m *membershipGroups) ListGroupByUser(string) ([]*groups.Group, error) {
	return nil, nil
}
func (m *membershipGroups) AddUser(_ *groups.Group, userID string) error {
	m.addedUser = userID
	return nil
}
func (m *membershipGroups) RemoveUser(_ *groups.Group, userID string) error {
	m.removedUser = userID
	return nil
}

func TestAddUserToGroupReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &membershipGroups{}
	r := gin.New()
	r.POST("/v0/groups/:group_id/users", AddUserToGroupHandler(svc, zap.NewNop().Sugar()))

	req := httptest.NewRequest(http.MethodPost, "/v0/groups/g1/users", strings.NewReader(`{"user_id":"u1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if w.Body.Len() == 0 {
		t.Fatal("empty body")
	}
	if svc.addedUser != "u1" {
		t.Fatalf("added %q", svc.addedUser)
	}
}

func TestRemoveUserFromGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &membershipGroups{}
	r := gin.New()
	r.DELETE("/v0/groups/:group_id/users/:user_id", RemoveUserFromGroupHandler(svc, zap.NewNop().Sugar()))

	req := httptest.NewRequest(http.MethodDelete, "/v0/groups/g1/users/u1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if svc.removedUser != "u1" {
		t.Fatalf("removed %q", svc.removedUser)
	}
}
