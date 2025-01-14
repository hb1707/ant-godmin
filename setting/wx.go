package setting

import "log"

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
	AdminUserIds   string
}

var AdminAppid = "qywx"

var (
	WxAppConfig   = map[string]WxApp{}
	QyWxAppConfig = map[string]QyWxApp{}
)

func confQyWxAdmin() {
	app, err := Cfg.GetSection(AdminAppid)
	if err == nil {
		QyWxAppConfig[AdminAppid] = QyWxApp{
			Corpid:         app.Key("QYWX_CORPID").MustString(""),
			AgentId:        app.Key("QYWX_AGENT_ID").MustInt(0),
			Secret:         app.Key("QYWX_SECRET").MustString(""),
			Token:          app.Key("QYWX_TOKEN").MustString(""),
			EncodingAESKey: app.Key("QYWX_ENCODING_AES_KEY").MustString(""),
			AdminUserIds:   app.Key("QYWX_ADMIN_USERIDS").MustString(""),
		}
		log.Println("[INFO] QyWx Config", AdminAppid, "OK")
	} else {
		log.Println("[ERROR] QyWx Config", AdminAppid, "ERROR", err)
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
		log.Println("[INFO] Wx Config", section, "OK")
	} else {
		log.Println("[ERROR] Wx Config", section, "ERROR", err)
	}
}
