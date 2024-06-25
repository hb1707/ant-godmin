package qywx

import (
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	"github.com/silenceper/wechat/v2/work/externalcontact"
	"strconv"
)

func WxGetUser(appid string, qyUserid string) externalcontact.ExternalUserDetailResponse {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		Cache:      memory,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		Cache:      memory,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
	memory := cache.NewMemory()
	appid := setting.AdminAppid
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		Cache:      memory,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		CorpSecret: setting.SecretExternalContact,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetExternalContact()
	var reqMsg = new(externalcontact.ReqMessage)
	reqMsg.ChatType = externalcontact.ChatTypeSingle
	reqMsg.ExternalUserid = externalUserid
	reqMsg.Text.Content = msg
	res, err := wxCon.Send(reqMsg)
	if err != nil {
		log.Error(err)
	}
	return strconv.FormatInt(res, 10)
}
