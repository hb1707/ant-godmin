package qywx

import (
	"strconv"
	"strings"
	"time"

	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/work/appchat"
	workConfig "github.com/silenceper/wechat/v2/work/config"
)

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Picurl      string `json:"picurl"`
}

type News struct {
	Articles []Article `json:"articles"`
}
type SendNewsRequest struct {
	*appchat.SendRequestCommon
	ChatID  string `json:"chatid"`
	MsgType string `json:"msgtype"`
	News    News   `json:"news"`
}

func WxPushMsgToGroup(appid string, chatId []string, msg string, imgs []string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetAppChat()
	var reqMsg appchat.SendTextRequest
	reqMsg.SendRequestCommon = new(appchat.SendRequestCommon)
	reqMsg.ChatID = strings.Join(chatId, "|")
	reqMsg.MsgType = "text"
	reqMsg.Text = appchat.TextField{
		Content: msg,
	}
	res, err := wxCon.SendText(reqMsg)
	if err != nil {
		log.Error("[ERROR]", err)
		return ""
	}
	time.Sleep(time.Second)
	if len(imgs) > 0 {
		var reqMsg SendNewsRequest
		reqMsg.SendRequestCommon = new(appchat.SendRequestCommon)
		reqMsg.ChatID = strings.Join(chatId, "|")
		reqMsg.MsgType = "markdown"
		var articles []Article
		for i, i2 := range imgs {
			articles = append(articles, Article{
				Title:       "图片" + strconv.Itoa(i+1),
				Description: "图片" + strconv.Itoa(i+1),
				Url:         i2,
				Picurl:      i2,
			})
		}
		reqMsg.News = News{
			Articles: articles,
		}
		res2, err := wxCon.Send("news", reqMsg)
		if err != nil {
			log.Error("[ERROR]", err)
			return ""
		}
		return res2.ErrMsg
	}
	return res.ErrMsg
}

type SendMarkdownRequest struct {
	*appchat.SendRequestCommon
	ChatID   string            `json:"chatid"`
	MsgType  string            `json:"msgtype"`
	Markdown appchat.TextField `json:"markdown"`
}

func WxPushMarkdownToGroup(appid string, chatId []string, markdown string) string {
	wc := wechat.NewWechat()
	cfg := &workConfig.Config{
		CorpID:     setting.QyWxAppConfig[appid].Corpid,
		AgentID:    strconv.Itoa(setting.QyWxAppConfig[appid].AgentId),
		CorpSecret: setting.QyWxAppConfig[appid].Secret,
		Cache:      Memory(appid),
	}
	miniapp := wc.GetWork(cfg)
	wxCon := miniapp.GetAppChat()
	var reqMsg SendMarkdownRequest
	reqMsg.SendRequestCommon = new(appchat.SendRequestCommon)
	reqMsg.ChatID = strings.Join(chatId, "|")
	reqMsg.MsgType = "markdown"
	reqMsg.Markdown = appchat.TextField{
		Content: markdown,
	}
	res, err := wxCon.Send("markdown", reqMsg)
	if err != nil {
		log.Error("[ERROR]", err)
		return ""
	}
	return res.ErrMsg
}
