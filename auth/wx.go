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
	"time"
)

var WxMemory map[string]cache.Cache
var WxMemoryAt map[string]time.Time

func Memory(appid string) cache.Cache {
	if WxMemory == nil {
		WxMemory = make(map[string]cache.Cache)
	}
	if WxMemoryAt == nil {
		WxMemoryAt = make(map[string]time.Time)
	}
	if WxMemory[appid] == nil || time.Since(WxMemoryAt[appid]) > time.Hour {
		WxMemory[appid] = cache.NewMemory()
		WxMemoryAt[appid] = time.Now()
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
