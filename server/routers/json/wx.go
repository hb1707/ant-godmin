package json

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"net/http"
)

func QyWxConnect(c *gin.Context) {
	if fun.Stripos(c.Request.UserAgent(), "wxwork") > 0 {
		c.Redirect(http.StatusFound, fmt.Sprintf(
			"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s/user/connect&response_type=code&scope=snsapi_base&state=STATE#wechat_redirect",
			setting.Corpid, setting.App.WEBURL))
	} else {
		c.Redirect(http.StatusFound, fmt.Sprintf(
			"https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=%s&agentid=%s&redirect_uri=%s/user/connect&state=STATE",
			setting.Corpid, setting.QyWxAppConfig[setting.AdminAppid].AgentId, setting.App.WEBURL))
	}
	return
}

func QyWxJsConfig(c *gin.Context) {
	url := c.Request.Referer()
	data := auth.GetQyWxConfig(setting.AdminAppid, url)
	jsonResult(c, http.StatusOK, data)
	return
}

func QyWxAgentJsConfig(c *gin.Context) {
	url := c.Request.Referer()
	data := auth.GetQyWxAgentConfig(setting.AdminAppid, url)
	jsonResult(c, http.StatusOK, data)
	return
}

type ReqWxLoginWithPhone struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password"`
	PhoneCode string `json:"phoneCode"`
	WxCode    string `json:"wxCode" binding:"required"`
}

type ReqWxPhone struct {
	SessionKey    string `json:"sessionKey" binding:"required"`
	EncryptedData string `json:"encryptedData" binding:"required"`
	Iv            string `json:"iv" binding:"required"`
	UnionId       string `json:"unionId"`
	OpenId        string `json:"openId" `
	TmpId         uint   `json:"tmpId" form:"tmpId"`
}
