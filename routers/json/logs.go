package json

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/pkg/log"
	"time"
)

type ReqLog struct {
	Msg    string    `json:"msg" form:"msg"`
	Level  int       `json:"level" form:"level"`
	Client string    `json:"client" form:"client"`
	Time   time.Time `json:"time" form:"time"`
}

func AddLog(c *gin.Context) {
	var req ReqLog
	err := c.BindJSON(&req)
	if err != nil {
		log.Error(err)
	}
	if req.Level == 0 {
		log.Info("["+req.Client+"]", req.Time, req.Msg)
	} else if req.Level == 1 {
		log.Warning("["+req.Client+"]", req.Time, req.Msg)
	} else if req.Level == 2 {
		log.Error("["+req.Client+"]", req.Time, req.Msg)
	} else if req.Level == 3 {
		log.Error("["+req.Client+"]", req.Time, req.Msg)
	}
}
