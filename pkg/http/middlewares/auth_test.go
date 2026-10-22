package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/root-ali/iris/pkg/auth"
	"go.uber.org/zap"
)

type fakeRoleRepo struct{}

func (fakeRoleRepo) GetRoleByID(string) (string, string, error) { return "viewer", "", nil }

func TestFlaggedTokenOnlyReachesPasswordChangeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop().Sugar()
	aths := auth.NewAuthService([]byte("test-secret"), fakeRoleRepo{}, logger)

	r := gin.New()
	ok := func(c *gin.Context) { c.Status(200) }
	r.GET("/v0/users/me", ValidateJWTToken(aths, "", logger), ok)
	r.PUT("/v0/users/password", ValidateJWTToken(aths, "", logger), ok)
	r.GET("/v0/alerts", ValidateJWTToken(aths, "admin,viewer", logger), ok)

	flagged, _ := aths.GenerateToken("bob", "role-id", true)
	clean, _ := aths.GenerateToken("bob", "role-id", false)

	tests := []struct {
		token, method, path string
		want                int
	}{
		{flagged, "GET", "/v0/users/me", 200},
		{flagged, "PUT", "/v0/users/password", 200},
		{flagged, "GET", "/v0/alerts", 403},
		{clean, "GET", "/v0/alerts", 200},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		req.Header.Set("Authorization", "Bearer "+tt.token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tt.want {
			t.Errorf("%s %s (token flagged=%v): got %d, want %d", tt.method, tt.path, tt.token == flagged, w.Code, tt.want)
		}
	}
}
