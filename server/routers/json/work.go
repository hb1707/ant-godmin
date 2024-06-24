package json

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/setting"
	"github.com/hb1707/exfun/fun"
	"net/http"
	"strconv"
	"time"
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
		oldUser := model.NewSysUser("qywx_userid = ?", res.UserID).GetOne("")
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
		oldUser := model.NewSysUser("id = ?", userID).GetOne("")
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
	userQyWx, err := auth.GetQyUser(appid, res.UserID)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	if res.UserID == "" && res.OpenID == "" {
		jsonErr(c, http.StatusBadRequest, consts.ErrEmptyUnionID)
		return
	}
	oldUser := model.NewSysUser("username = ?", userQyWx.Mobile).GetOne("")
	var u *model.SysUsers
	userQyWx.UserID = res.UserID
	if oldUser.Id > 0 {
		auth.UpdateQywxUserid(oldUser.Id, res.UserID)
		u = oldUser
	} else {
		var reg = new(auth.UserReg)
		reg.Userid = userQyWx.UserID
		reg.Username = userQyWx.Mobile
		reg.RealName = userQyWx.Name
		reg.Mobile = userQyWx.Mobile
		reg.Avatar = userQyWx.Avatar
		reg.QyWxUser = userQyWx
		if reg.Password1 != reg.Password2 {
			jsonErr(c, http.StatusBadRequest, consts.ErrInconsistentPassword)
			return
		}
		u := model.NewSysUser()
		u.UUID = uuid.New()
		reg.Password1 = "e053cc3b86e072868bfaf1be0be78331" //123456edu_we
		reg.Password2 = reg.Password1

		u.QywxUserid = userQyWx.UserID
		u.HeaderImg = reg.Avatar
		u.NickName = userQyWx.Alias
		u.Username = reg.Username
		u.RealName = reg.RealName
		u.AuthorityId = consts.AuthorityIdStaff
		salt := fun.SubStr(fun.MD5(strconv.Itoa(int(time.Now().UnixNano()))), 0, 4)
		u.Salt = salt
		u.Password = auth.Cryptosystem(reg.Password1, salt)
		u.Edit()
		reg, err = auth.RegisterHandler(appid, reg)
		if err != nil {
			jsonErr(c, http.StatusBadRequest, err)
			return
		}
		oldUser = u
	}
	data, err := auth.TokenGenerator(&auth.TokenUser{
		UUID:        oldUser.UUID,
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
