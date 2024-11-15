package wx

import (
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram/config"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/miniprogram/message"
	"github.com/silenceper/wechat/v2/miniprogram/qrcode"
	"github.com/silenceper/wechat/v2/miniprogram/subscribe"
	"net/http"
	"strings"
)

// MAppDecryptPost 解密微信小程序消息
func MAppDecryptPost(appid string, r *http.Request) (message.MsgType, message.EventType, message.PushData, error) {
	wc := wechat.NewWechat()
	memory := auth.Memory(appid)
	sessionKey := setting.WxAppConfig[appid].EncodingAESKey
	cfg := &config.Config{
		AppID:          appid,
		Cache:          memory,
		Token:          setting.WxAppConfig[appid].Token,
		EncodingAESKey: sessionKey,
	}
	miniapp := wc.GetMiniProgram(cfg)
	wxServ := miniapp.GetMessageReceiver()
	msgType, eventType, data, err := wxServ.GetMsgData(r)
	if err != nil {
		log.Error("GetMsgData", msgType, eventType, err)
		return "", "", nil, err
	}
	return msgType, eventType, data, nil
}

// MAppDecryptGet 解密微信小程序消息
func MAppDecryptGet(appid string, r *http.Request) (string, message.PushData, error) {
	wc := wechat.NewWechat()
	memory := auth.Memory(appid)
	sessionKey := setting.WxAppConfig[appid].EncodingAESKey
	cfg := &config.Config{
		AppID:          appid,
		Cache:          memory,
		Token:          setting.WxAppConfig[appid].Token,
		EncodingAESKey: sessionKey,
	}
	miniapp := wc.GetMiniProgram(cfg)
	wxServ := miniapp.GetMessageReceiver()
	msgType, data, err := wxServ.GetMsg(r)
	if err != nil {
		log.Error("GetMsgData", msgType, err)
		return "", nil, err
	}
	return msgType, data, nil
}

// SubscribeSend 发送订阅消息
func SubscribeSend(appid string, toUser string, templateID string, data map[string]*subscribe.DataItem, page string) error {
	wc := wechat.NewWechat()
	memory := auth.Memory(appid)
	sessionKey := setting.WxAppConfig[appid].EncodingAESKey
	cfg := &config.Config{
		AppID:          appid,
		Cache:          memory,
		Token:          setting.WxAppConfig[appid].Token,
		EncodingAESKey: sessionKey,
	}
	miniapp := wc.GetMiniProgram(cfg)
	subri := miniapp.GetSubscribe()
	msg := subscribe.Message{}
	msg.ToUser = toUser
	msg.TemplateID = templateID
	msg.Page = page
	msg.Data = data
	msg.Lang = "zh_CN"
	err := subri.Send(&msg)
	if err != nil {
		log.Error("Send", err)
		return err
	}
	return nil
}

func GetMiniAppQrcode(appid string, path string, isTest bool) []byte {
	var scene = ""
	pathArr := strings.Split(path, "?")
	if len(pathArr) > 1 {
		scene = pathArr[1]
		path = pathArr[0]
	}
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &miniConfig.Config{
		AppID:     appid,
		AppSecret: setting.WxAppConfig[appid].AppSecret,
		Cache:     memory,
	}
	miniApp := wc.GetMiniProgram(cfg)
	qr := miniApp.GetQRCode()
	imgByte, err := qr.GetWXACodeUnlimit(qrcode.QRCoder{
		Page:       path,
		Scene:      scene,
		EnvVersion: fun.If2String(setting.IsTest || isTest, "trial", "release"),
	})
	if err != nil {
		log.Error(err)
		return nil
	}
	return imgByte
}
func GetMiniAppQrcodeStatic(appid string, path string) []byte {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &miniConfig.Config{
		AppID:     appid,
		AppSecret: setting.WxAppConfig[appid].AppSecret,
		Cache:     memory,
	}
	miniApp := wc.GetMiniProgram(cfg)
	qr := miniApp.GetQRCode()
	imgByte, err := qr.GetWXACode(qrcode.QRCoder{
		Path:       path,
		EnvVersion: fun.If2String(setting.IsTest, "develop", "release"),
	})
	if err != nil {
		log.Error(err)
		return nil
	}
	return imgByte
}
