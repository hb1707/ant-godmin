package json

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/common"
)

func TestAppReadinessWaitsBeforeBecomingReadyAndNotifiesOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)

	readinessMu.Lock()
	timeStart = time.Time{}
	readinessMu.Unlock()
	readinessNoticeOnce = sync.Once{}
	noticeCount := 0
	readinessNotice = func() {
		noticeCount++
	}
	t.Cleanup(func() {
		readinessMu.Lock()
		timeStart = time.Time{}
		readinessMu.Unlock()
		readinessNoticeOnce = sync.Once{}
		readinessNotice = common.ReadinessNotice
	})

	router := gin.New()
	router.GET("/begin", AppBegin)
	router.GET("/readiness", AppReadiness)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/begin", nil)
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("/begin: expected status %d, got %d", http.StatusOK, response.Code)
	}

	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/readiness", nil)
		router.ServeHTTP(response, request)

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("/readiness before delay: expected status %d, got %d", http.StatusInternalServerError, response.Code)
		}
	}
	if noticeCount != 0 {
		t.Fatalf("before delay: expected no readiness notice, got %d", noticeCount)
	}

	readinessMu.Lock()
	timeStart = time.Now().Add(-readinessDelay)
	readinessMu.Unlock()

	for i := 0; i < 2; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/readiness", nil)
		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("/readiness after delay: expected status %d, got %d", http.StatusOK, response.Code)
		}
	}
	if noticeCount != 1 {
		t.Fatalf("after delay: expected one readiness notice, got %d", noticeCount)
	}
}
