package json

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/common"
)

func AppTest(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
	return
}

const readinessDelay = 10 * time.Second

var (
	timeStart           time.Time
	readinessMu         sync.Mutex
	readinessNoticeOnce sync.Once
	readinessNotice     = common.ReadinessNotice
)

func AppBegin(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// AppReadiness 在首次探测后保留一段启动缓冲时间，避免进程刚监听端口就被切入业务流量。
// 启动探针与就绪探针职责分离：/begin 只确认进程已启动，流量切换仅以这里的结果为准。
func AppReadiness(c *gin.Context) {
	if !readinessReady(time.Now()) {
		c.JSON(500, gin.H{
			"message": "not ready",
		})
		return
	}

	// 通知属于“已经真正就绪”的副作用；并发探针和后续轮询都不能重复发送。
	readinessNoticeOnce.Do(readinessNotice)
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func readinessReady(now time.Time) bool {
	readinessMu.Lock()
	defer readinessMu.Unlock()

	if timeStart.IsZero() {
		timeStart = now
		return false
	}
	return now.Sub(timeStart) >= readinessDelay
}
