package qywx

import (
	"encoding/json"
	"fmt"
	log2 "log"
	"strconv"
	"strings"

	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/util"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/message"
)

func WxPushMsgToStaff(appid string, userid []string, msg string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()
	var reqMsg message.SendTextRequest
	reqMsg.SendRequestCommon = new(message.SendRequestCommon)
	reqMsg.AgentID = strconv.Itoa(setting.QyWxAppConfig[appid].AgentId)
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

func WxPushMsgCard(appid string, userid []string, msg *TemplateCardButton) (string, string, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()
	var reqMsg TemplateCardRequest
	reqMsg.SendRequestCommon = new(message.SendRequestCommon)
	reqMsg.ToUser = strings.Join(userid, "|")
	reqMsg.AgentID = strconv.Itoa(setting.QyWxAppConfig[appid].AgentId)
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
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetMessage()

	var reqMsg = new(TemplateUpdate)
	if userid != nil && len(userid) > 0 {
		reqMsg.Userids = userid
	} else {
		reqMsg.Atall = 1
	}
	reqMsg.Agentid = agentId
	reqMsg.ResponseCode = responseCode
	reqMsg.UpdateButton = NewUpdateButton(replaceName)
	res, err := UpdateTemplate(wxCon, reqMsg)
	if err != nil {
		log.Error("[ERROR]", err)
		return ""
	}
	return res
}

// NewUpdateButton 更新点击用户的按钮文案
func NewUpdateButton(replaceName string) *UpdateButton {
	btn := new(UpdateButton)
	btn.Button.ReplaceName = replaceName
	return btn
}

const messageUpdateTemplateCardURL = "https://api.weixin.qq.com/cgi-bin/message/update_template_card"

type resTemplateSend struct {
	util.CommonError
	Invaliduser  string `json:"invaliduser"`   //不合法的userid，不区分大小写，统一转为小写
	Invalidparty string `json:"invalidparty"`  //不合法的partyid
	Invalidtag   string `json:"invalidtag"`    //不合法的标签id
	MsgID        string `json:"msgid"`         //消息id，用于撤回应用消息
	ResponseCode string `json:"response_code"` //仅消息类型为“按钮交互型”，“投票选择型”和“多项选择型”的模板卡片消息返回，应用可使用response_code调用更新模版卡片消息接口，24小时内有效，且只能使用一次
}

// UpdateTemplate 更新模版卡片消息
func UpdateTemplate(wxCon *message.Client, msg *TemplateUpdate) (msgID string, err error) {
	var accessToken string
	accessToken, err = wxCon.GetAccessToken()
	if err != nil {
		return
	}
	uri := fmt.Sprintf("%s?access_token=%s", messageUpdateTemplateCardURL, accessToken)
	var response []byte
	response, err = util.PostJSON(uri, msg)
	if err != nil {
		return
	}
	var result resTemplateSend
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg send error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	msgID = result.MsgID
	return
}

const messageDelURL = "https://api.weixin.qq.com/cgi-bin/message/recall"

type ReqRecall struct {
	MsgID int64 `json:"msgid"`
}

func MessageRecall(wxCon *message.Client, msgID int64) (err error) {
	var accessToken string
	accessToken, err = wxCon.GetAccessToken()
	if err != nil {
		return
	}
	uri := fmt.Sprintf("%s?access_token=%s", messageDelURL, accessToken)
	var response []byte
	response, err = util.PostJSON(uri, &ReqRecall{
		MsgID: msgID,
	})
	if err != nil {
		return
	}
	var result util.CommonError
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg send error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}
