package auth

import (
	"strconv"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/exfun/fun"
	"github.com/silenceper/wechat/v2/work/addresslist"
)

var TypeMap = map[string]string{
	"v": "_1",
}
var PasswordSalt string

type TokenUser struct {
	UUID        uuid.UUID
	UidHash     string
	Bid         uint
	Sub         string
	Role        string
	ID          uint
	Appid       string
	AuthorityId string
	StaffId     uint
	AdmLv       int
	Tester      uint
	jwt.MapClaims
}

type UserReg struct {
	Userid    string                       `json:"userid" form:"userid"`
	Username  string                       `json:"username" form:"username"`
	RealName  string                       `json:"realName" form:"realName"`
	Avatar    string                       `json:"avatar" form:"avatar"`
	Mobile    string                       `json:"mobile" form:"mobile"`
	Password1 string                       `json:"password1" form:"password1"`
	Password2 string                       `json:"password2" form:"password2"`
	QyWxUser  *addresslist.UserGetResponse `json:"qywxUser" form:"qywxUser"`
}

var RegisterHandler func(appid string, user *UserReg) (*UserReg, error)
var GetSelf func(uid int) (exist bool, user interface{})

type UpdatePost struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Avatar   string `json:"avatar" form:"avatar"`
	NickName string `json:"nick_name" json:"nick_name"`
}

type BindPost struct {
	WxOpenid string
}

func TokenGenerator(user *TokenUser) (map[string]interface{}, error) {
	userToken, expire, err := Middleware().TokenGenerator(user)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		//"uid":         user.ID,
		//"nickName":    user.RealName,
		"bid":         user.Bid,
		"sub":         user.Sub,
		"role":        user.Role,
		"uidHash":     user.UidHash,
		"token":       userToken,
		"tokenExpire": expire,
	}, nil
}
func TokenClear(user *TokenUser) (map[string]interface{}, error) {
	userToken, expire, err := Middleware().TokenGenerator(user)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		//"uid":         user.ID,
		//"nickName":    user.RealName,
		"token":       userToken,
		"tokenExpire": expire,
	}, nil
}
func LoginCheckPw(p *LoginPost) *model.SysUsers {
	oldUser := model.NewSysUser("username = ?", p.Username).GetOne("")
	postPW := Cryptosystem(p.Password, oldUser.Salt)
	if p.Password != "" && postPW == oldUser.Password {
		return oldUser
	}
	return nil
}
func HelloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	id, _ := c.Get(identityKey)
	c.JSON(200, gin.H{
		"claims":    claims,
		identityKey: id,
		"text":      "Hello World.",
	})
}
func UpdateQywxUserid(uid uint, userid string) {
	u := model.NewSysUser()
	u.Id = uid
	u.QywxUserid = userid
	u.Edit()
}
func UpdateWxUnionId(uid uint, unionId string) {
	u := model.NewSysUser()
	u.Id = uid
	u.WxUnionId = unionId
	u.Edit()
}

func Update(uid uint, up *UpdatePost) {
	u := model.NewSysUser()
	u.Id = uid
	if up.Password != "" {
		salt := fun.SubStr(fun.MD5(strconv.Itoa(int(time.Now().UnixNano()))), 0, 4)
		up.Password = Cryptosystem(up.Password, salt)
		u.Password = up.Password
	}
	u.HeaderImg = up.Avatar
	u.Username = up.Username
	u.NickName = up.NickName
	u.Edit()
}

func (u *TokenUser) IsAdmin() bool {
	if u.AdmLv > 0 {
		return true
	} else {
		return false
	}
}

func Cryptosystem(password string, salt string) string {
	return fun.MD5(password + PasswordSalt + salt)
}
