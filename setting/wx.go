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
	app, _ := Cfg.GetSection(AdminAppid)
	QyWxAppConfig[AdminAppid] = QyWxApp{
		Corpid:         getString(app, AdminAppid, "QYWX_CORPID", ""),
		AgentId:        getInt(app, AdminAppid, "QYWX_AGENT_ID", 0),
		Secret:         getString(app, AdminAppid, "QYWX_SECRET", ""),
		Token:          getString(app, AdminAppid, "QYWX_TOKEN", ""),
		EncodingAESKey: getString(app, AdminAppid, "QYWX_ENCODING_AES_KEY", ""),
		AdminUserIds:   getString(app, AdminAppid, "QYWX_ADMIN_USERIDS", ""),
	}
	log.Println("[INFO] QyWx Config", AdminAppid, "OK")
}

func ConfWxApp(section string, appid string) {
	app, _ := Cfg.GetSection(section)
	WxAppConfig[appid] = WxApp{
		AppSecret:      getString(app, section, "WX_SECRET", ""),
		Token:          getString(app, section, "WX_TOKEN", ""),
		EncodingAESKey: getString(app, section, "WX_ENCODING_AES_KEY", ""),
	}
	log.Println("[INFO] Wx Config", section, "OK")
}
