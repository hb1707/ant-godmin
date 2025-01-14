package coze

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/exfun/fun/curl"
	"os"
	"time"
)

const (
	tokenUri                = "https://api.coze.cn/api/permission/oauth2/token"
	oauthCodeUri            = "https://www.coze.cn/api/permission/oauth2/authorize"
	oauthCodeUriCollaborate = "https://www.coze.cn/api/permission/oauth2/workspace_id"
	agentListUri            = "https://api.coze.cn/v1/space/published_bots_list"
	botPublish              = "https://api.coze.cn/v1/bot/publish"
)

type Config struct {
	Host         string
	Appid        string
	CertPath     string
	ClientId     string
	ClientSecret string
	WorkspaceId  string
	RedirectUri  string
}

type Client struct {
	Token string
	Config
}

func NewClient(c Config) *Client {
	return &Client{
		Config: c,
	}
}

func (c *Client) OAuthCodeUri(key string) string {
	params := map[string]string{
		"response_type": "code",
		"client_id":     c.ClientId,
		"redirect_uri":  c.RedirectUri,
		"state":         key,
	}
	uri := curl.Web(oauthCodeUri, params)
	return uri
	//client := &http.Client{}
	//req, err := http.NewRequest("GET", uri, nil)
	//if err != nil {
	//	log.Error(err)
	//	return ""
	//}
	//client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
	//	return http.ErrUseLastResponse
	//}
	//resp, err := client.Do(req)
	//if err != nil {
	//	log.Error(err)
	//	return ""
	//}
	//defer resp.Body.Close()
	//
	//return resp.Header.Get("Location")
}
func (c *Client) OAuthCodeUriCollaborate(key string) string {
	params := map[string]string{
		"response_type": "code",
		"client_id":     c.ClientId,
		"redirect_uri":  c.RedirectUri,
		"state":         key,
	}
	uri := curl.Web(oauthCodeUriCollaborate+"/"+c.WorkspaceId+"/authorize", params)
	return uri
}
func (c *Client) OAuthToken(code string) (string, string, error) {
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.ClientSecret,
		},
	}
	params := map[string]any{
		"grant_type":   "authorization_code",
		"code":         code,
		"client_id":    c.ClientId,
		"redirect_uri": c.RedirectUri,
	}
	b, _ := json.Marshal(params)
	resp, _, err := client.POST(tokenUri, b)
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	var result Result
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	if result.ErrorMessage != "" {
		log.Error(string(resp))
		return "", "", errors.New(result.ErrorMessage)
	}
	c.Token = result.AccessToken
	return result.AccessToken, result.RefreshToken, nil
}
func (c *Client) OAuthTokenRefresh(token string) (string, string, error) {
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.ClientSecret,
		},
	}
	params := map[string]any{
		"client_id":     c.ClientId,
		"grant_type":    "refresh_token",
		"refresh_token": token,
	}
	b, _ := json.Marshal(params)
	resp, _, err := client.POST(tokenUri, b)
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	var result Result
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	if result.ErrorMessage != "" {
		log.Error(string(resp))
		return "", "", errors.New(result.ErrorMessage)
	}
	c.Token = result.AccessToken
	return result.AccessToken, result.RefreshToken, nil
}
func (c *Client) GenJWT(key string) string {
	// 生成jwt token
	// 准备 JWT 的 Header 和 Payload

	payload := jwt.MapClaims{
		"iss": c.Appid,                              // OAuth 应用的 ID
		"aud": "api.coze.cn",                        // 扣子 API 的 Endpoint
		"iat": time.Now().Unix(),                    // JWT 开始生效的时间
		"exp": time.Now().Add(1 * time.Hour).Unix(), // JWT 过期时间
		"jti": fmt.Sprint(time.Now().UnixMilli()),   // 随机字符串，防止重放攻击
	}
	//p, _ := json.Marshal(payload)
	//log.Debug(string(p))
	pem, err := os.ReadFile(c.CertPath)
	if err != nil {
		log.Error(err)
		return ""
	}
	//log.Debug(string(pem))
	// 使用私钥进行签名
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		log.Error(err)
		return ""
	}

	// 创建 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)
	token.Header["kid"] = key // 设置公钥指纹
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		log.Error(err)
		return ""
	}
	return tokenString
}
func (c *Client) TokenByJWT(key string) (string, string, error) {
	var token = c.GenJWT(key)
	//log.Debug(token)
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + token,
		},
	}
	params := map[string]any{
		"duration_seconds": 86399,
		"grant_type":       "urn:ietf:params:oauth:grant-type:jwt-bearer",
	}
	b, _ := json.Marshal(params)
	resp, _, err := client.POST(tokenUri, b)
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	var result Result
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return "", "", err
	}
	if result.Error != "" {
		log.Error(result.ErrorMessage)
		return "", "", errors.New(result.ErrorMessage)
	}
	c.Token = result.AccessToken
	return result.AccessToken, result.RefreshToken, nil

}
