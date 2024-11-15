package setting

type WxApp struct {
	AppSecret      string
	Token          string // 接收消息时的token
	EncodingAESKey string // 接收消息时的EncodingAESKey
}
type QyWxApp struct {
	Corpid         string
	Secret         string
	AgentId        int
	KfSecret       string
	Token          string // 接收消息时的token
	EncodingAESKey string // 接收消息时的EncodingAESKey
}

var AdminAppid = "qywx"

var (
	WxAppConfig   = map[string]WxApp{}
	QyWxAppConfig = map[string]QyWxApp{}
)

func init() {
	confQyWxAdmin()
}
func confQyWxAdmin() {
	app, err := Cfg.GetSection(AdminAppid)
	if err == nil {
		QyWxAppConfig[AdminAppid] = QyWxApp{
			Corpid:         app.Key("QYWX_CORPID").MustString(""),
			AgentId:        app.Key("QYWX_AGENT_ID").MustInt(0),
			Secret:         app.Key("QYWX_SECRET").MustString(""),
			Token:          app.Key("QYWX_TOKEN").MustString(""),
			EncodingAESKey: app.Key("QYWX_ENCODING_AES_KEY").MustString(""),
		}
	}
}

func ConfWxApp(section string, appid string) {
	app, err := Cfg.GetSection(section)
	if err == nil {
		WxAppConfig[appid] = WxApp{
			AppSecret:      app.Key("WX_SECRET").MustString(""),
			Token:          app.Key("WX_TOKEN").MustString(""),
			EncodingAESKey: app.Key("WX_ENCODING_AES_KEY").MustString(""),
		}
	}
}
