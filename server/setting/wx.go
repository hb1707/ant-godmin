package setting

type WxApp struct {
	AppSecret string
}
type WxOffiaccount struct {
	Secret string
}
type QyWxApp struct {
	Secret   string
	AgentId  string
	KfSecret string
}

var Corpid = ""
var AdminAppid = "admin"

var (
	WxOaConfig    = map[string]WxOffiaccount{}
	WxAppConfig   = map[string]WxApp{}
	QyWxAppConfig = map[string]QyWxApp{}

	WxWorkToken           = ""
	WxWorkEncodingAESKey  = ""
	SecretExternalContact = ""
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
