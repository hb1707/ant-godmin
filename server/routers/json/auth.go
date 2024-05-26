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
	"strings"
	"time"
)

type ReqRegister struct {
	Username  string `json:"username" form:"username"`
	Password1 string `json:"password1" form:"password1"`
	Password2 string `json:"password2" form:"password2"`
}

func RegisterWithPassword(c *gin.Context) {
	var req ReqRegister
	err := c.BindJSON(&req)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	if req.Password1 != req.Password2 {
		jsonErr(c, http.StatusBadRequest, consts.ErrInconsistentPassword)
		return
	}
	var reg = new(auth.UserReg)
	reg.Username = req.Username
	reg.Password1 = req.Password1
	reg.Password2 = req.Password2
	reg, err = auth.RegisterHandler(setting.AdminAppid, reg)
	if err != nil {
		jsonErr(c, http.StatusBadRequest, err)
		return
	}
	if reg.Password1 != reg.Password2 {
		jsonErr(c, http.StatusBadRequest, consts.ErrInconsistentPassword)
		return
	}
	u := model.NewSysUser()
	u.UUID = uuid.New()
	u.QywxUserid = ""
	u.HeaderImg = reg.Avatar
	u.NickName = reg.Username
	u.Username = reg.Username
	u.RealName = reg.RealName
	u.AuthorityId = consts.AuthorityIdStaff
	salt := fun.SubStr(fun.MD5(strconv.Itoa(int(time.Now().UnixNano()))), 0, 4)
	u.Salt = salt
	u.Password = auth.Cryptosystem(reg.Password1, salt)
	u.Edit()
	oldUser := u
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
				userQyWx, err := auth.GetQyUser(setting.AdminAppid, res.UserID)
				if err != nil {
					jsonErr(c, http.StatusBadRequest, err)
					return
				}
				/*if userQyWx.Mobile == "" {
					jsonErr(c, http.StatusBadRequest, consts.ErrEmptyMoblie)
					return
				}*/
				var reg = new(auth.UserReg)
				reg.Userid = userQyWx.UserID
				reg.Username = userQyWx.Alias
				reg.RealName = userQyWx.Name
				reg.Mobile = userQyWx.Mobile
				reg.Avatar = userQyWx.Avatar
				reg, err = auth.RegisterHandler(setting.AdminAppid, reg)
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
				u.NickName = reg.Username
				u.Username = reg.Username
				u.RealName = reg.RealName
				u.AuthorityId = consts.AuthorityIdStaff
				salt := fun.SubStr(fun.MD5(strconv.Itoa(int(time.Now().UnixNano()))), 0, 4)
				u.Salt = salt
				u.Password = auth.Cryptosystem(reg.Password1, salt)
				u.Edit()
				oldUser = u
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

func RefreshToken(c *gin.Context) {
	ver := c.GetHeader("version")
	verArr := strings.Split(ver, ".")
	var isNew = false
	var versionOld = []int{1, 1, 18}
	if len(verArr) > 0 {
		if len(verArr) == 3 {
			ver1, _ := strconv.Atoi(verArr[0])
			ver2, _ := strconv.Atoi(verArr[1])
			ver3, _ := strconv.Atoi(verArr[2])
			if ver1 > versionOld[0] || (ver1 == versionOld[0] && ver2 > versionOld[1]) || (ver1 == versionOld[0] && ver2 == versionOld[1] && ver3 >= versionOld[2]) {
				isNew = true
			}
		}
	}
	md := auth.Middleware()
	if !isNew {
		md.MaxRefresh = md.MaxRefresh * 4
	}
	md.RefreshHandler(c)
	return
}
