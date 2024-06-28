package wx

import (
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
)

func WxOaUploadImg(appId string, image string) (string, error) {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: setting.WxAppConfig[appId].AppSecret,
		Cache:     memory,
	}
	official := wc.GetOfficialAccount(cfg)
	m := official.GetMaterial()
	upload, err := m.ImageUpload(image)
	if err != nil {
		return "", err
	}
	return upload, nil
}
