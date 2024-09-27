package wx

import (
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/miniprogram/message"
	"net/http"
)

// MAppDecryptPost 解密微信小程序消息
func MAppDecryptPost(appid string, r *http.Request) (message.PushData, error) {
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
		return nil, err
	}
	return data, nil
}

// MAppDecryptGet 解密微信小程序消息
func MAppDecryptGet(appid string, r *http.Request) (message.PushData, error) {
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
		return nil, err
	}
	return data, nil
}
