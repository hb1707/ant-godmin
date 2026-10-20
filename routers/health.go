package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/routers/json"
)

// RegisterHealthRoutes 统一注册启动、就绪和存活探针，避免业务模块遗漏入口。
// 可传入 Engine 或 RouterGroup；同一前缀只注册一次，并在业务鉴权中间件之前调用。
// 保留原处理函数及响应格式：/begin 只确认启动，/readiness 承担就绪缓冲和通知。
func RegisterHealthRoutes(r gin.IRoutes) {
	r.GET("/test", json.AppTest)
	r.GET("/begin", json.AppBegin)
	r.GET("/readiness", json.AppReadiness)
}
