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

var timeStart time.Time

func AppBegin(c *gin.Context) {
	if !timeStart.IsZero() && time.Since(timeStart) < time.Second*10 {
		c.JSON(500, gin.H{
			"message": "not ready",
		})
		return
	}
	timeStart = time.Now()
	common.ReadinessNotice()
	c.JSON(200, gin.H{
		"message": "pong",
	})
	return
}
