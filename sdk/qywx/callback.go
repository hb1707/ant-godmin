package qywx

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/kf"
	"github.com/silenceper/wechat/v2/work/message"
)

func WxKfDecrypt(appid string, req kf.SignatureOptions) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:         setting.QyWxAppConfig[appid].Corpid,
		AgentID:        strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret:     setting.QyWxAppConfig[appid].KfSecret,
		Cache:          auth.Memory(appid),
		Token:          setting.QyWxAppConfig[appid].Token,
		EncodingAESKey: setting.QyWxAppConfig[appid].EncodingAESKey,
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
	appid := setting.AdminAppid
	cfg := &workConfig.Config{
		CorpID:         setting.QyWxAppConfig[appid].Corpid,
		Cache:          auth.Memory(appid),
		Token:          setting.QyWxAppConfig[appid].Token,
		EncodingAESKey: setting.QyWxAppConfig[appid].EncodingAESKey,
	}
	work := wc.GetWork(cfg)
	wxMsg := work.GetServer(req, writer)
	res, err := wxMsg.VerifyURL()
	if err != nil {
		log.Println("[ERROR]", err)
	}
	return res
}

func WxServerDecryptPost(req *http.Request, writer http.ResponseWriter, messageHandler func(msg *message.MixMessage) *message.Reply) {
	wc := wechat.NewWechat()
	appid := setting.AdminAppid
	cfg := &workConfig.Config{
		CorpID:         setting.QyWxAppConfig[appid].Corpid,
		Cache:          auth.Memory(appid),
		Token:          setting.QyWxAppConfig[appid].Token,
		EncodingAESKey: setting.QyWxAppConfig[appid].EncodingAESKey,
	}
	miniapp := wc.GetWork(cfg)
	wxServ := miniapp.GetServer(req, writer)
	var bodyBytes []byte // 我们需要的body内容
	// 从原有Request.Body读取
	bodyBytes, _ = io.ReadAll(req.Body)
	// 新建缓冲区并替换原有Request.body
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	// 当前函数可以使用body内容
	//WxPushMsgToStaff([]string{"HuangBin"}, string(bodyBytes))
	wxServ.SetMessageHandler(messageHandler)
	err := wxServ.Serve()
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}
	err = wxServ.Send()
	if err != nil {
		log.Println("[ERROR]", err)
		return
	}
}
