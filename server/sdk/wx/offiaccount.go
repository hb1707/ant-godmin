package wx

import (
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/setting"
	offConfig "github.com/hb1707/wechat/v2/officialaccount/config"
)

func WxOaUploadImg(appId string, image string) (string, error) {
	wc := wechat.NewWechat()
	cfg := &offConfig.Config{
		AppID:     appId,
		AppSecret: setting.WxAppConfig[appId].AppSecret,
		Cache:     auth.Memory(appId),
	}
	official := wc.GetOfficialAccount(cfg)
	m := official.GetMaterial()
	upload, err := m.ImageUpload(image)
	if err != nil {
		return "", err
	}
	return upload, nil
}
