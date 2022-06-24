package json

import (
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/auth"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/setting"
	"net/http"
)

type ReqLogin struct {
	Username  string `json:"username" form:"username"`
	Password  string `json:"password" form:"password"`
	AutoLogin bool   `json:"autoLogin" form:"autoLogin"`
	Code      string `json:"code" form:"code"`
	Type      string `json:"type" form:"type" binding:"required"`
}

func LoginWithPasswordOrQywxCode(c *gin.Context) {
	var req ReqLogin
	err := c.BindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	if req.Type == "qywx" && req.Code != "" {
		res, err := auth.GetQyWxUserID(setting.AdminAppid, req.Code)
		if err != nil {
			jsonErr(c, http.StatusBadRequest, err)
			return
		}

		if res.UserID != "" {
			oldUser := model.NewSysUser("qywx_userid = ?", res.UserID).GetOne("id desc")
			if oldUser.Id == 0 {
				user, err := auth.GetQyUser(setting.AdminAppid, res.UserID)
				if err != nil {
					jsonErr(c, http.StatusBadRequest, err)
					return
				}
				/*if user.Mobile == "" {
					jsonErr(c, http.StatusBadRequest, consts.ErrEmptyMoblie)
					return
				}*/
				oldUser, err = auth.RegisterHandler(setting.AdminAppid, user)
				if err != nil {
					jsonErr(c, http.StatusBadRequest, err)
					return
				}
			} else if oldUser.Username != "" {
				user, err := auth.GetQyUser(setting.AdminAppid, res.UserID)
				if err != nil {
					jsonErr(c, http.StatusBadRequest, err)
					return
				}
				var up auth.UpdatePost
				up.Avatar = user.Avatar
				up.NickName = user.Alias
				auth.Update(oldUser.Id, &up)
			} else {
				user, err := auth.GetQyUser(setting.AdminAppid, res.UserID)
				if err != nil {
					jsonErr(c, http.StatusBadRequest, err)
					return
				}
				if user.Mobile == "" {
					jsonErr(c, http.StatusBadRequest, consts.ErrEmptyMoblie)
					return
				}
				var up auth.UpdatePost
				up.Avatar = user.Avatar
				up.Username = user.Mobile
				up.NickName = user.Alias
				auth.Update(oldUser.Id, &up)
			}
			data, err := auth.TokenGenerator(&auth.TokenUser{
				UUID:        oldUser.UUID,
				ID:          oldUser.Id,
				Appid:       "",
				AuthorityId: oldUser.AuthorityId,
				AdmLv:       0,
			})
			if err != nil {
				jsonErr(c, http.StatusBadRequest, err)
				return
			}
			jsonResult(c, http.StatusOK, data, gin.H{"currentAuthority": oldUser.AuthorityId})
			return
		}
	} else {
		oldUser := model.NewSysUser("username = ?", req.Username).GetOne("")
		if oldUser.Id > 0 {
			if req.Password != "" && req.Username != "" {
				checkUser := auth.LoginCheckPw(&auth.LoginPost{
					Username: req.Username,
					Password: req.Password,
				})
				if checkUser != nil {
					data, err := auth.TokenGenerator(&auth.TokenUser{
						UUID:        oldUser.UUID,
						ID:          oldUser.Id,
						Appid:       "",
						AuthorityId: oldUser.AuthorityId,
						AdmLv:       0,
					})
					if err != nil {
						jsonErr(c, http.StatusBadRequest, err)
						return
					}
					jsonResult(c, http.StatusOK, data, gin.H{"currentAuthority": checkUser.AuthorityId})
					return
				}
			}
		}
	}
	jsonErr(c, http.StatusOK, consts.ErrFailedAuthentication, gin.H{"type": "account"})
	return
}

func GetUser(c *gin.Context) {
	uid := auth.GetAdmUID(c)
	if uid > 0 {
		exist, user := auth.GetSelf(uid)
		if exist {
			jsonResult(c, http.StatusOK, user)
			return
		} else {
			jsonErr(c, http.StatusOK, consts.ErrUnauthorized, gin.H{"data": map[string]interface{}{"isLogin": false}})
			return
		}
	}
}
