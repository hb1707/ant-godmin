package json

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/setting"
	"net/http"
)

func WxGetWorkUser(c *gin.Context) {
	var req auth.ReqUser
	err := c.BindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	appid := c.GetHeader("Appid")
	res, err := auth.GetQyOpenID(appid, req.Code)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	if res.UserID == "" && res.OpenID == "" {
		jsonErr(c, http.StatusBadRequest, consts.ErrEmptyUnionID)
		return
	}
	if res.UserID != "" {
		oldUser := model.NewSysUser("qywx_userid = ?", res.UserID).One("")
		data, err := auth.TokenGenerator(&auth.TokenUser{
			UUID:        oldUser.UUID,
			ID:          oldUser.Id,
			Appid:       appid,
			AuthorityId: oldUser.AuthorityId,
			AdmLv:       0,
		})
		if err != nil {
			jsonErr(c, http.StatusBadRequest, err)
			return
		}
		data["deviceID"] = res.DeviceID
		data["qywxUserinfo"] = nil
		if oldUser.Id == 0 || req.GetQy {
			user, err := auth.GetQyUser(appid, res.UserID)
			if err != nil {
				jsonErr(c, http.StatusBadRequest, err)
				return
			}
			data["userinfo"] = user
		}
		jsonResult(c, http.StatusOK, data)
		return
	}
}
func WxGetLaunchCode(c *gin.Context) {
	var req auth.ReqLaunchCode
	err := c.BindQuery(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	appid := c.GetHeader("Appid")
	if appid == "" {
		appid = setting.AdminAppid
	}
	userID := auth.GetAdmUID(c)
	if userID > 0 {
		oldUser := model.NewSysUser("id = ?", userID).One("")
		if oldUser.Id > 0 {
			user, err := auth.GetQyLaunchCode(appid, oldUser.QywxUserid, req.UserId)
			if err != nil {
				jsonErr(c, http.StatusBadRequest, err)
				return
			}
			jsonResult(c, http.StatusOK, user)
			return
		}
	}
	jsonErr(c, http.StatusBadRequest, consts.ErrUnauthorized)
	return
}

type ReqQyReg struct {
	Username string `json:"username" `
	WxCode   string `json:"wxCode" binding:"required"`
}

func WorkRegister(c *gin.Context) {
	var req ReqQyReg
	err := c.BindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	appid := c.GetHeader("Appid")
	res, err := auth.GetQyOpenID(appid, req.WxCode)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	user, err := auth.GetQyUser(appid, res.UserID)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	if res.UserID == "" && res.OpenID == "" {
		jsonErr(c, http.StatusBadRequest, consts.ErrEmptyUnionID)
		return
	}
	oldUser := model.NewSysUser("username = ?", user.Mobile).One("")
	var u *model.SysUsers
	user.Userid = res.UserID
	if oldUser.Id > 0 {
		auth.UpdateQywxUserid(oldUser.Id, res.UserID)
		u = oldUser
	} else {
		u, err = auth.RegisterHandler(appid, user)
		if err != nil {
			jsonErr(c, http.StatusBadRequest, err)
			return
		}
	}
	data, err := auth.TokenGenerator(&auth.TokenUser{
		UUID:        u.UUID,
		ID:          u.Id,
		Appid:       appid,
		AuthorityId: u.AuthorityId,
		AdmLv:       0,
	})
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	jsonResult(c, http.StatusOK, data)
	return
}
