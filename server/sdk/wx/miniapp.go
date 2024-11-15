package wx

import (
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/wechat/v2/miniprogram/config"
	"github.com/hb1707/wechat/v2/miniprogram/message"
	"github.com/hb1707/wechat/v2/miniprogram/subscribe"
	"net/http"
)

// MAppDecryptPost 解密微信小程序消息
func MAppDecryptPost(appid string, r *http.Request) (message.MsgType, message.EventType, message.PushData, error) {
	wc := wechat.NewWechat()
	memory := auth.Memory(appid)
	sessionKey := setting.WxMAppEncodingAESKey
	cfg := &config.Config{
		AppID:          appid,
		Cache:          memory,
		Token:          setting.WxMAppToken,
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
	sessionKey := setting.WxMAppEncodingAESKey
	cfg := &config.Config{
		AppID:          appid,
		Cache:          memory,
		Token:          setting.WxMAppToken,
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
	sessionKey := setting.WxMAppEncodingAESKey
	cfg := &config.Config{
		AppID:          appid,
		Cache:          memory,
		Token:          setting.WxMAppToken,
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
