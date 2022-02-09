package json

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
)

var ErrorParameter = errors.New("参数错误")

func jsonResult(c *gin.Context, code int, data interface{}, other ...gin.H) {
	okData := make(gin.H)
	for _, m := range other {
		for k, v := range m {
			okData[k] = v
		}
	}
	okData["success"] = true
	okData["result"] = data
	okData["status"] = "ok"
	c.JSON(code, okData)
}

func jsonErr(c *gin.Context, code int, msg error, data ...gin.H) {
	errData := make(gin.H)
	for _, m := range data {
		for k, v := range m {
			errData[k] = v
		}
	}
	errData["success"] = false
	errData["errorCode"] = strconv.Itoa(code)
	errData["errorMessage"] = msg.Error()
	errData["status"] = "error"
	c.JSON(code, errData)

}
func strResult(c *gin.Context, code int, data string) {
	c.String(code, data)
}
