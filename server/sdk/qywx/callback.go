package qywx

import (
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/setting"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/kf"
	"log"
	"net/http"
)

func WxKfDecrypt(appid string, req kf.SignatureOptions) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:         setting.Corpid,
		AgentID:        setting.QyWxAppConfig[appid].AgentId,
		CorpSecret:     setting.QyWxAppConfig[appid].KfSecret,
		Cache:          auth.Memory(appid),
		Token:          setting.WxWorkToken,
		EncodingAESKey: setting.WxWorkEncodingAESKey,
	}
	miniapp := wc.GetWork(cfg)
	wxKf, err := miniapp.GetKF()
	if err != nil {
		return err.Error()
	}
	res, err := wxKf.VerifyURL(req)
	return res
}
func WxServerDecryptGet(req *http.Request, writer http.ResponseWriter) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:         setting.Corpid,
		Cache:          auth.Memory("appid"),
		Token:          setting.WxWorkToken,
		EncodingAESKey: setting.WxWorkEncodingAESKey,
	}
	work := wc.GetWork(cfg)
	wxMsg := work.GetServer(req, writer)
	res, err := wxMsg.VerifyURL()
	if err != nil {
		log.Println("[ERROR]", err)
	}
	return res
}
