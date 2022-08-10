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

func init() {
	confQyWxAdmin()
}
func confQyWxAdmin() {
	app, err := Cfg.GetSection("wx")
	if err == nil {
		Corpid = app.Key("QYWX_CORPID").MustString("")
		QyWxAppConfig[AdminAppid] = QyWxApp{
			AgentId: app.Key("QYWX_AGENT_ID").MustString(""),
			Secret:  app.Key("QYWX_SECRET").MustString(""),
		}
	}
}
