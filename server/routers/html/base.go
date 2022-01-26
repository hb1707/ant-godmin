package html

import (
	"github.com/flosch/pongo2/v4"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/setting"
)

func DataHtml(c *gin.Context, data map[string]interface{}) pongo2.Context {

	m := c.GetInt("uid")
	if _, exist := data["full_screen"]; !exist {
		data["full_screen"] = false
	}
	data["me"] = m
	//data["is_admin"] = m.Uid == module.AdminUid
	data["web_url"] = setting.App.WEBURL
	data["api_url"] = setting.App.APIURL
	return data
}
