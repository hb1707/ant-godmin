package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/sdk/aliyun"
	qywxAdmin "github.com/hb1707/ant-godmin/sdk/qywx"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
)

var appInfo setting.AppInfo
var ReadinessProbe bool

func InitApp() {
	appInfo = setting.ReadAppInfo("./app.json")
}

func ReadinessNotice() {
	var sae = fun.If2String(appInfo.Name != "", appInfo.Name, "管理中心")
	if !ReadinessProbe {
		ReadinessProbe = true
		card := new(qywxAdmin.TemplateCardButton)
		card.CardType = "button_interaction"
		card.Source = new(qywxAdmin.Source)
		card.Source.IconUrl = fmt.Sprintf("%s%s", setting.App.STATICURL, "/assets/logo.png")
		card.Source.Desc = fmt.Sprintf("服务已更新 %s", time.Now().Format("2006-01-02 15:04:05"))
		card.Source.DescColor = 1
		card.MainTitle = new(qywxAdmin.MainTitle)
		card.MainTitle.Title = fmt.Sprintf("%s [%s] 服务已启动", strings.ToUpper(sae), fun.Domain(setting.App.APIURL, ""))
		card.MainTitle.Desc = fmt.Sprintf("版本：%s\n更新通道：%s\n更新内容：%s", appInfo.Version, appInfo.Channel, appInfo.Desc)
		card.TaskId = fmt.Sprintf("ver_%s_%s_%s", sae, time.Now().Format("20060102150405"), fun.MD5(setting.App.APIURL))
		card.ButtonList = []qywxAdmin.Button{
			{Text: "忽略", Style: 2, Key: "ver_" + sae + "_ignore"},
			{Text: "部署/回滚", Style: 1, Key: "ver_" + sae + "_rollback"},
		}
		appid := setting.AdminAppid
		if setting.QyWxAppConfig[setting.AdminAppid].AdminUserIds == "" {
			log.Error("AdminUserIds为空")
			return
		}
		userIds := strings.Split(setting.QyWxAppConfig[setting.AdminAppid].AdminUserIds, "|")
		_, _, err := qywxAdmin.WxPushMsgCard(appid, userIds, card)
		if err != nil {
			aliyun.SendEmail(setting.Email.SystemMail, setting.Email.AdminEmail, "QYWX API报错", fmt.Sprintf("%s服务启动推送失败：%s", sae, err))
			log.Error(err)
		}
	}
}
