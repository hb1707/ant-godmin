package auth

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/exfun/fun"
	"net/http"
	"strconv"
	"time"
)

var (
	identityKey = "ID"
	maxRefresh  = time.Hour * 24
	tokenMaxAge = time.Hour * 24 * 7 //Second
	Realm       string
	Key         string
)

type LoginPost struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}
type Auth struct {
	Uid      int    `json:"uid"`
	UserName string `json:"username"`
	Token    Token  `json:"-"`
}

type Token struct {
	Exp time.Time `json:"exp"`
	Str string    `json:"str"`
}

var CacheAuth = make(map[int]*Auth)

func Middleware() *jwt.GinJWTMiddleware {
	m, err := NewMiddleware()
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	return m
}

func NewMiddleware() (*jwt.GinJWTMiddleware, error) {
	// the jwt middleware
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:             Realm,              //HTTP Basic Auth 账号密码作用域
		Key:               []byte(Key),        //服务端密钥
		Timeout:           time.Hour * 24 * 7, //token 过期时间
		MaxRefresh:        maxRefresh,         //token 允许更新时间
		SendAuthorization: true,
		SendCookie:        false,
		IdentityKey:       identityKey,
		//1. 在初次登录的接口中使用的验证方法，并返回验证成功后的用户对象。
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals LoginPost
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", consts.ErrMissingLoginValues
			}

			if user := LoginCheckPw(&loginVals); user != nil {

				return &TokenUser{
					UUID:        user.UUID, //PayloadFunc中将此字段值赋给sub
					ID:          user.Id,
					Appid:       c.GetHeader("Appid"),
					AuthorityId: user.AuthorityId,
					AdmLv:       fun.If2Int(user.AuthorityId == consts.AuthorityIdAdmin, 1, 0),
				}, nil
			}

			return nil, consts.ErrFailedAuthentication
		},

		//2. 生成Token时，添加额外业务相关的信息

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*TokenUser); ok {
				m := fun.Struct2Map(v, "")
				//b, _ := json.Marshal(v)
				return m
			}
			return jwt.MapClaims{}
		},

		//2.1 登录成功后返回
		LoginResponse: func(c *gin.Context, code int, token string, t time.Time) {
			c.JSON(http.StatusOK, gin.H{
				"code":    code,
				"token":   token,
				"expire":  t.Format(time.RFC3339),
				"max":     maxRefresh / time.Second,
				"message": "login successfully",
			})
		},
		//2.2 刷新成功后返回
		RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {
			//user, _ := CheckTokenUser(c)
			//CacheAuth[user.Uid].SetToken(token)
			c.JSON(http.StatusOK, gin.H{
				"code":    code,
				"token":   token,
				"expire":  t.Format(time.RFC3339),
				"max":     maxRefresh / time.Second,
				"message": "refresh successfully",
			})
		},

		//3. 已登录，接收请求时，其他接口验证token通过之后，提取必要的身份信息，并放入Context
		IdentityHandler: func(c *gin.Context) interface{} {
			//var u TokenUser
			claims := jwt.ExtractClaims(c)
			//u = claims.(TokenUser) //对应PayloadFunc -> jwt.MapClaims
			//_ = json.Unmarshal([]byte(sub), &u)
			if claims["AuthorityId"].(string) == "" {
				c.Set("is_user", true)
			} else {
				c.Set("authority_id", claims["AuthorityId"].(string))
			}
			return int(claims[identityKey].(float64)) //对应下文Authorizator 接收的参数
		},
		//4. 已登录，接收请求时，其他接口将提取的身份信息做最后一步的验证
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if id, ok := data.(int); ok && id > 0 {
				return true
			}
			return false //403
		},
		//5. 验证失败后设置错误信息
		Unauthorized: func(c *gin.Context, code int, message string) {
			errData := make(gin.H)
			errData["success"] = false
			errData["errorCode"] = strconv.Itoa(code)
			errData["errorMessage"] = message
			errData["status"] = "error"
			c.JSON(code, errData)
		},
		LogoutResponse: func(c *gin.Context, code int) {

			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
			})
			//user, _ := CheckTokenUser(c)
			//CacheAuth[user.Uid].SetToken("")
		},
		//设置token获取位置，一般默认在头部的Authorization中，或者query的token字段，cookie中的jwt字段。
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		//Header中token的头部字段，默认常用名称Bearer。
		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		//设置时间函数
		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})
}
func CheckTokenUser(c *gin.Context) {
	sub, exists := c.Get(identityKey)
	if exists {
		uid := sub.(int)
		if uid > 0 {
			if c.GetBool("is_user") {
				c.Set("user_uid", uid)
			} else {
				c.Set("adm_uid", uid)
			}
			c.Next()
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"message": "权限无效！",
	})
	return
}
func GetUID(c *gin.Context) int {
	sub, exists := c.Get("adm_uid")
	if exists {
		return sub.(int)
	}
	return 0
}
func GetAuthID(c *gin.Context) string {
	sub, exists := c.Get("authority_id")
	if exists {
		return sub.(string)
	}
	return ""
}
func Identity(c *gin.Context) (int, string) {
	uid := GetUID(c)
	authId := GetAuthID(c)
	return uid, authId
}
func (a *Auth) SetToken(token string) Token {
	if a == nil {
		return Token{}
	}
	a.Token.Str = token
	if token != "" {
		a.Token.Exp = time.Now().Add(tokenMaxAge)
	} else {
		a.Token.Exp = time.Time{}
	}

	return a.Token
}
