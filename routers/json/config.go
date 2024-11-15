package json

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"net/http"
)

type RespConfig struct {
	Title         string `json:"title"`
	SuperAdminUid uint   `json:"super_admin_uid"`
}

func Config(c *gin.Context) {
	var config RespConfig
	var existAdmin = model.NewSysUser("authority_id = ?", consts.AuthorityIdSuperAdmin).GetOne("id desc")
	if existAdmin.Id > 0 {
		config.SuperAdminUid = existAdmin.Id
	}
	jsonResult(c, http.StatusOK, config)
	return
}
