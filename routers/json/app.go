package json

import (
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

// AppReadiness 只表达当前 HTTP 进程已经能够接收请求。
// 探针会被并发、重复调用，因此这里不能修改进程状态或触发一次性启动逻辑。
func AppReadiness(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ready",
	})
}

var timeStart time.Time

func AppBegin(c *gin.Context) {
	if !timeStart.IsZero() && time.Since(timeStart) < time.Second*10 {
		c.JSON(500, gin.H{
			"message": "not ready",
		})
		return
	}
	if timeStart.IsZero() {
		timeStart = time.Now()
	}
	common.ReadinessNotice()
	c.JSON(200, gin.H{
		"message": "pong",
	})
	return
}
