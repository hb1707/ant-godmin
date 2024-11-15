package auth

import (
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/work/addresslist"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	workJS "github.com/silenceper/wechat/v2/work/js"
	"github.com/silenceper/wechat/v2/work/oauth"
	"strconv"
	"time"
)

func GetQyOpenID(appid, code string) (oauth.ResUserInfo, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxAuth := miniapp.GetOauth()
	res, err := wxAuth.Code2Session(code)
	return res, err
}
func GetQyWxUserID(appid, code string) (oauth.ResUserInfo, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxAuth := miniapp.GetOauth()
	res, err := wxAuth.UserFromCode(code)
	return res, err
}

func GetQyWxConfig(appid, url string) (conf *workJS.Config) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
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

var cacheGetQyWxAgentConfigTime = make(map[string]time.Time)
var cacheQyWxAgentConfig = make(map[string]*workJS.Config)

func GetQyWxAgentConfig(appid, url string) (conf *workJS.Config) {
	var cacheKey = appid + url
	if time.Since(cacheGetQyWxAgentConfigTime[cacheKey]) < time.Hour && cacheQyWxAgentConfig[cacheKey] != nil {
		return cacheQyWxAgentConfig[cacheKey]
	}
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxJs := miniapp.GetJs()
	conf, err := wxJs.GetConfig(url)
	if err != nil {
		log.Error(err)
		return
	}
	cacheGetQyWxAgentConfigTime[cacheKey] = time.Now()
	cacheQyWxAgentConfig[cacheKey] = conf
	return
}
func GetQyUser(appid, userID string) (*addresslist.UserGetResponse, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxUser := miniapp.GetAddressList()
	userInfo, err := wxUser.UserGet(userID)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

type ReqLaunchCode struct {
	UserId string `json:"userId" form:"userId"`
}

func GetQyLaunchCode(appid, userID, other string) (*oauth.RespLaunchCode, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxUser := miniapp.GetOauth()
	res, err := wxUser.GetLaunchCode(userID, other)
	return res, err
}

type ReqUser struct {
	Code  string `json:"code" form:"code"`
	GetQy bool   `json:"get_qy" form:"get_qy"`
	From  string `json:"from" form:"from"`
	Appid string `json:"appid" form:"appid"`
}
