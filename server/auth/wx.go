package auth

import (
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram/auth"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	offJS "github.com/silenceper/wechat/v2/officialaccount/js"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	workJS "github.com/silenceper/wechat/v2/work/js"
	"github.com/silenceper/wechat/v2/work/oauth"
	"github.com/silenceper/wechat/v2/work/user"
)

func GetMpAccessToken(appId string) (string, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: setting.WxAppConfig[appId].AppSecret,
		Cache:     memory,
	}
	official := wc.GetOfficialAccount(cfg)
	res, err := official.GetAccessToken()
	return res, err
}

func GetMpJsConfig(appId, url string) (*offJS.Config, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: setting.WxAppConfig[appId].AppSecret,
		Cache:     memory,
	}
	official := wc.GetOfficialAccount(cfg)
	js := official.GetJs()
	config, err := js.GetConfig(url)
	return config, err
}

func GetOpenID(appid, code string) (auth.ResCode2Session, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &miniConfig.Config{
		AppID:     appid,
		AppSecret: setting.WxAppConfig[appid].AppSecret,
		Cache:     memory,
	}
	miniapp := wc.GetMiniProgram(cfg)
	wxAuth := miniapp.GetAuth()
	res, err := wxAuth.Code2Session(code)
	return res, err
}
func GetQyOpenID(appid, code string) (oauth.ResUserInfo, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxAuth := miniapp.GetOauth()
	res, err := wxAuth.Code2Session(code)
	return res, err
}
func GetQyWxUserID(appid, code string) (oauth.ResUserInfo, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxAuth := miniapp.GetOauth()
	res, err := wxAuth.UserFromCode(code)
	return res, err
}
func GetQyWxConfig(appid, url string) (conf *workJS.Config) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxJs := miniapp.GetJs()
	conf, err := wxJs.GetConfig(url)
	if err != nil {
		log.Error(err)
		return
	}
	return
}
func GetQyWxAgentConfig(appid, url string) (conf *workJS.Config) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxJs := miniapp.GetJs()
	conf, err := wxJs.GetConfig(url)
	if err != nil {
		log.Error(err)
		return
	}
	return
}
func GetQyUser(appid, userID string) (*user.Info, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxUser := miniapp.GetUser()
	res, err := wxUser.GetUserInfo(userID)
	return res, err
}

type ReqLaunchCode struct {
	UserId string `json:"userId" form:"userId"`
}

func GetQyLaunchCode(appid, userID, other string) (*user.RespLaunchCode, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      memory,
	}
	miniapp := wc.GetWork(cfg)
	wxUser := miniapp.GetUser()
	res, err := wxUser.GetLaunchCode(userID, other)
	return res, err
}

type ReqUser struct {
	Code  string `json:"code" form:"code"`
	GetQy bool   `json:"get_qy" form:"get_qy"`
	From  string `json:"from" form:"from"`
	Appid string `json:"appid" form:"appid"`
}
