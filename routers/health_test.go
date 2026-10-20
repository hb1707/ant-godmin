package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterHealthRoutesRootAndOnlineCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 与服务启动一致：先注册库内 /online 别名，再注册业务模块根路径。
	RegisterHealthRoutes(r.Group("/online"))
	RegisterHealthRoutes(r)
	// 业务鉴权必须不影响无凭证的探针请求。
	r.Group("/api").Use(func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }).GET("/private", func(c *gin.Context) {})
	if got := len(r.Routes()); got != 7 {
		t.Fatalf("expected six health routes and one business route, got %d", got)
	}
	for _, prefix := range []string{"", "/online"} {
		for _, path := range []string{"/test", "/begin", "/readiness"} {
			t.Run(prefix+path, func(t *testing.T) {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, prefix+path, nil))
				want := http.StatusOK
				body := `{"message":"pong"}`
				if path == "/readiness" {
					// 首次探测仍需缓冲，不能因统一注册变成无条件就绪；不触发真实通知。
					want = http.StatusInternalServerError
					body = `{"message":"not ready"}`
				}
				if w.Code != want || w.Body.String() != body {
					t.Fatalf("got %d %s, want %d %s", w.Code, w.Body.String(), want, body)
				}
			})
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/private", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("business authentication changed: status %d", w.Code)
	}
}
