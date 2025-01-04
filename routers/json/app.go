package json

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/common"
)

func AppTest(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
	return
}

func AppBegin(c *gin.Context) {
	common.ReadinessNotice()
	c.JSON(200, gin.H{
		"message": "pong",
	})
	return
}
