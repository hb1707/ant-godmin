package setting

type WxApp struct {
	AppSecret string
}
type QyWxApp struct {
	Secret   string
	AgentId  string
	KfSecret string
}

var Corpid = ""
var AdminAppid = "admin"
var (
	WxAppConfig   = map[string]WxApp{}
	QyWxAppConfig = map[string]QyWxApp{}
)
