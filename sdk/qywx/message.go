package qywx

import (
	"encoding/json"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/message"
	log2 "log"
	"strconv"
	"strings"
)

func WxPushMsgToStaff(appid string, userid []string, msg string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()
	var reqMsg message.SendTextRequest
	reqMsg.SendRequestCommon = new(message.SendRequestCommon)
	reqMsg.AgentID = setting.QyWxAppConfig[appid].AgentId
	reqMsg.ToUser = strings.Join(userid, "|")
	reqMsg.MsgType = "text"
	reqMsg.Text = message.TextField{
		Content: msg,
	}
	res, err := wxCon.SendText(reqMsg)
	if err != nil {
		log2.Println("[ERROR]", err) //防止循环引用
		return ""
	}
	return res.MsgID
}
func WxPushMsgToGroup(appid string, userid []string, msg string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()
	var reqMsg message.SendTextRequest
	reqMsg.SendRequestCommon = new(message.SendRequestCommon)
	reqMsg.ToUser = strings.Join(userid, "|")
	reqMsg.AgentID = setting.QyWxAppConfig[appid].AgentId
	reqMsg.MsgType = "text"
	reqMsg.Text = message.TextField{
		Content: msg,
	}
	res, err := wxCon.SendText(reqMsg)
	if err != nil {
		log.Error("[ERROR]", err)
		return ""
	}
	return res.MsgID
}
func WxPushMsgCard(appid string, userid []string, msg *TemplateCardButton) (string, string, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()
	var reqMsg TemplateCardRequest
	reqMsg.SendRequestCommon = new(message.SendRequestCommon)
	reqMsg.ToUser = strings.Join(userid, "|")
	reqMsg.AgentID = setting.QyWxAppConfig[appid].AgentId
	reqMsg.MsgType = "template_card"
	reqMsg.TemplateCard = msg
	res, err := wxCon.Send("MessageSendTemplateCard", reqMsg)
	if err != nil {
		data, _ := json.Marshal(msg)
		log.Error("[ERROR]", err, string(data))
		return "", "", err
	}
	return res.MsgID, res.ResponseCode, nil
}
func WxPushMsgCardUpdate(appid string, userid []string, responseCode string, agentId int, replaceName string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()
	var reqMsg = new(message.TemplateUpdate)
	if userid != nil && len(userid) > 0 {
		reqMsg.Userids = userid
	} else {
		reqMsg.Atall = 1
	}
	reqMsg.Agentid = agentId
	reqMsg.ResponseCode = responseCode
	reqMsg.UpdateButton = message.NewUpdateButton(replaceName)
	res, err := wxCon.UpdateTemplate(reqMsg)
	if err != nil {
		log.Error("[ERROR]", err)
		return ""
	}
	return res
}
