package auth

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/hb1707/ant-godmin/consts"
	"github.com/hb1707/ant-godmin/model"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/exfun/fun"
	"net/http"
	"strconv"
	"time"
)

var (
	anonymousKey = "AnonymousID"
	identityKey  = "ID"
	maxRefresh   = time.Hour * 24 * 90
	tokenMaxAge  = time.Hour * 24 * 30 //Second
	Realm        string
	Key          string
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
	if Realm == "" {
		log.Fatal("Realm未配置！")
	}
	m, err := NewMiddleware()
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	return m
}

func NewMiddleware() (*jwt.GinJWTMiddleware, error) {
	// the jwt middleware
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:             Realm,       //HTTP Basic Auth 账号密码作用域
		Key:               []byte(Key), //服务端密钥
		Timeout:           tokenMaxAge, //token 过期时间
		MaxRefresh:        maxRefresh,  //token 允许更新时间
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
				if staffId, ok := claims["StaffId"].(float64); staffId > 0 && ok {
					c.Set("staff_id", staffId)
				}
				c.Set("is_user", true)
			} else {
				c.Set("is_user", false)
				c.Set("authority_id", claims["AuthorityId"].(string))
			}
			if tester, ok := claims["Tester"].(float64); tester > 0 && ok {
				c.Set("tester_lev", uint(tester))
			} else {
				c.Set("tester_lev", uint(0))
			}
			if uuidStr, ok := claims["UidHash"].(string); uuidStr != "" && ok {
				c.Set("uid_hash", uuidStr)
			} else {
				c.Set("uid_hash", "")
			}
			if bid, ok := claims["Bid"].(float64); bid > 0 && ok {
				c.Set("bid", uint(bid))
			} else {
				c.Set("bid", 0)
			}
			if sub, ok := claims["Sub"].(string); sub != "" && ok {
				c.Set("sub", sub)
			} else {
				c.Set("sub", "")
			}
			if role, ok := claims["Role"].(string); role != "" && ok {
				c.Set("role", role)
			} else {
				c.Set("role", "")
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
			path := c.Request.URL.Path

			errData := make(gin.H)
			errData["success"] = false
			errData["errorCode"] = strconv.Itoa(code)
			errData["errorMessage"] = message
			if message == jwt.ErrExpiredToken.Error() && path == "/api/nft/user/fresh" {
				errData["message"] = "refresh successfully"
			}
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

type GinJWTMiddleware struct {
	*jwt.GinJWTMiddleware
}

func MiddlewareAnonymous() *GinJWTMiddleware {
	if Realm == "" {
		log.Fatal("Realm未配置！")
	}
	m, err := NewMiddleware()
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	return &GinJWTMiddleware{
		m,
	}
}

func (mw *GinJWTMiddleware) MiddlewareFuncAnonymous(c *gin.Context) {
	claims, err := mw.GetClaimsFromJWT(c)
	if err == nil {
		if claims["exp"] == nil {
			c.Next()
			return
		}

		if _, ok := claims["exp"].(float64); !ok {
			c.Next()
			return
		}

		if int64(claims["exp"].(float64)) < mw.TimeFunc().Unix() {
			c.Next()
			return
		}
		c.Set("JWT_PAYLOAD", claims)
		identity := mw.IdentityHandler(c)
		if identity != nil {
			c.Set(anonymousKey, identity)
		}
	}
	c.Next()
}
func CheckAdminPublic(c *gin.Context) {
	sub, exists := c.Get(anonymousKey)
	if exists {
		uid := sub.(int)
		if uid > 0 {
			c.Set("user_uid", uid)
		}
	}
	staffUid := GetStaffID(c)
	if staffUid > 0 {
		staff := model.NewSysUser("id = ?", staffUid).GetOne("")
		if staff.Id == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "权限无效！",
			})
			return
		} else {
			c.Set("authority_id", staff.AuthorityId)
		}
	}
	c.Next()
	return
}

func GetUserUID(c *gin.Context) int {
	sub, exists := c.Get("user_uid")
	if exists {
		return sub.(int)
	}
	return 0
}
func GetUidHash(c *gin.Context) string {
	sub, exists := c.Get("uid_hash")
	if exists {
		parse := sub.(string)
		return parse
	}
	return ""
}

func GetBid(c *gin.Context) uint {
	sub, exists := c.Get("bid")
	if exists {
		return sub.(uint)
	}
	return 0
}
func GetBidSub(c *gin.Context) string {
	sub, exists := c.Get("sub")
	if exists {
		return sub.(string)
	}
	return ""
}

func GetAdmUID(c *gin.Context) int {
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
func GetStaffID(c *gin.Context) uint {
	sub, exists := c.Get("staff_id")
	if exists {
		return uint(sub.(float64))
	}
	return 0
}
func GetTesterLev(c *gin.Context) uint {
	sub, exists := c.Get("tester_lev")
	if exists {
		return sub.(uint)
	}
	return 0
}
func Identity(c *gin.Context) (int, string) {
	uid := GetAdmUID(c)
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
