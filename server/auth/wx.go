package auth

import (
	"fmt"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram/auth"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	offJS "github.com/silenceper/wechat/v2/officialaccount/js"
	"github.com/silenceper/wechat/v2/util"
	"github.com/silenceper/wechat/v2/work/addresslist"
	workConfig "github.com/silenceper/wechat/v2/work/config"
	workJS "github.com/silenceper/wechat/v2/work/js"
	"github.com/silenceper/wechat/v2/work/oauth"
	"time"
)

var WxMemory map[string]cache.Cache

func Memory(appid string) cache.Cache {
	if WxMemory == nil {
		WxMemory = make(map[string]cache.Cache)
	}
	if WxMemory[appid] == nil {
		WxMemory[appid] = cache.NewMemory()
	}
	return WxMemory[appid]
}

var cacheGetMpAccessTokenTime = make(map[string]time.Time)
var cacheMpAccessToken = make(map[string]string)

func GetMpAccessToken(appId string, isRefresh bool) (string, error) {
	if !isRefresh && time.Since(cacheGetMpAccessTokenTime[appId]) < time.Minute*3 && cacheMpAccessToken[appId] != "" {
		return cacheMpAccessToken[appId], nil
	}
	wc := wechat.NewWechat()
	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: setting.WxAppConfig[appId].AppSecret,
		Cache:     Memory(appId),
	}
	official := wc.GetOfficialAccount(cfg)
	res, err := official.GetAccessToken()
	if err != nil {
		log.Error(err)
		return "", err
	}
	cacheGetMpAccessTokenTime[appId] = time.Now()
	cacheMpAccessToken[appId] = res
	return res, err
}

var cacheGetMpJsConfigTime = make(map[string]time.Time)
var cacheMpJsConfig = make(map[string]*offJS.Config)

func GetMpJsConfig(appId, url string) (*offJS.Config, error) {
	var cacheKey = appId + url
	if time.Since(cacheGetMpJsConfigTime[cacheKey]) < time.Hour && cacheMpJsConfig[cacheKey] != nil {
		return cacheMpJsConfig[cacheKey], nil
	}
	wc := wechat.NewWechat()
	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: setting.WxAppConfig[appId].AppSecret,
		Cache:     Memory(appId),
	}
	official := wc.GetOfficialAccount(cfg)
	accessToken, err := GetMpAccessToken(appId, false)
	if err != nil {
		return nil, err
	}

	js := official.GetJs()

	/* todo 有问题，跳过了缓存
	config, err := js.GetConfig(url)
	if err != nil {
		log.Error(err)
		return nil, err
	}*/

	var ticketStr string
	ticketStr, err = js.GetTicket(accessToken)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	nonceStr := util.RandomStr(16)
	timestamp := util.GetCurrTS()
	str := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", ticketStr, nonceStr, timestamp, url)
	sigStr := util.Signature(str)
	var config = new(offJS.Config)
	config.AppID = js.AppID
	config.NonceStr = nonceStr
	config.Timestamp = timestamp
	config.Signature = sigStr
	cacheGetMpJsConfigTime[cacheKey] = time.Now()
	cacheMpJsConfig[cacheKey] = config
	return config, err
}

func GetOpenID(appid, code string) (auth.ResCode2Session, error) {
	wc := wechat.NewWechat()
	cfg := &miniConfig.Config{
		AppID:     appid,
		AppSecret: setting.WxAppConfig[appid].AppSecret,
		Cache:     Memory(appid),
	}
	miniapp := wc.GetMiniProgram(cfg)
	wxAuth := miniapp.GetAuth()
	res, err := wxAuth.Code2Session(code)
	return res, err
}
func GetPhone(appid, code string) (*auth.GetPhoneNumberResponse, error) {
	wc := wechat.NewWechat()
	cfg := &miniConfig.Config{
		AppID:     appid,
		AppSecret: setting.WxAppConfig[appid].AppSecret,
		Cache:     Memory(appid),
	}
	miniapp := wc.GetMiniProgram(cfg)
	wxAuth := miniapp.GetAuth()
	res, err := wxAuth.GetPhoneNumber(code)
	return res, err
}
func GetQyOpenID(appid, code string) (oauth.ResUserInfo, error) {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
		CorpID:     setting.Corpid,
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
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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
		CorpID:     setting.Corpid,
		AgentID:    setting.QyWxAppConfig[appid].AgentId,
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

type ReqUserPhone struct {
	Code    string `json:"code" form:"code"`
	UnionID string `json:"unionId" form:"unionId"`
	From    string `json:"from" form:"from"`
	Appid   string `json:"appid" form:"appid"`
}
