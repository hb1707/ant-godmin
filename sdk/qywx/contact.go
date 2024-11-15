package qywx

import (
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/externalcontact"
	"strconv"
)

func WxGetUser(appid string, qyUserid string) externalcontact.ExternalUserDetailResponse {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		Cache:      Memory(appid),
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetExternalContact()
	res, err := wxCon.GetExternalUserDetail(qyUserid)
	if err != nil {
		log.Error(err)
	}
	return *res
}
func WxEditUserTag(appid string, qyUserid string, externalUserid string, tags []string) error {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		Cache:      Memory(appid),
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetExternalContact()
	var reqMsg = externalcontact.MarkTagRequest{}
	reqMsg.UserID = qyUserid
	reqMsg.ExternalUserID = externalUserid
	reqMsg.AddTag = tags
	err := wxCon.MarkTag(reqMsg)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func WxGetMyUsers(qyUserid string) []string {
	wc := wechat.NewWechat()
	appid := setting.AdminAppid
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		Cache:      Memory(appid),
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetExternalContact()
	res, err := wxCon.GetExternalUserList(qyUserid)
	if err != nil {
		log.Error(err)
	}
	return res

}
func WxPushMsgToUser(externalUserid []string, msg string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		CorpSecret: setting.SecretExternalContact,
		Cache:      Memory("appid"),
	}
	work := wc.GetWork(cfg)
	wxCon := work.GetExternalContact()
	var reqMsg = new(externalcontact.AddMsgTemplateRequest)
	reqMsg.ChatType = "single"
	reqMsg.ExternalUserID = externalUserid
	reqMsg.Text.Content = msg
	res, err := wxCon.AddMsgTemplate(reqMsg)
	if err != nil {
		log.Error(err)
	}
	return res.MsgID
}
