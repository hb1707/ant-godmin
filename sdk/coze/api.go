package coze

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/hb1707/exfun/fun/curl"
	"os"
	"strconv"
	"time"
)

const (
	tokenUri     = "https://api.coze.cn/api/permission/oauth2/token"
	agentListUri = "https://api.coze.cn/v1/space/published_bots_list"
)

type Config struct {
	Host     string
	Appid    string
	CertPath string
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

func (c *Client) GetJWT(key string) string {
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
func (c *Client) GetToken(key string) (string, error) {
	var token = c.GetJWT(key)
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
		return "", err
	}
	var result struct {
		AccessToken  string `json:"access_token"`
		Error        string `json:"error"`
		ErrorMessage string `json:"error_message"`
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return "", err
	}
	if result.Error != "" {
		log.Error(result.ErrorMessage)
		return "", errors.New(result.ErrorMessage)
	}
	c.Token = result.AccessToken
	return result.AccessToken, nil

}

type SpaceBot struct {
	BotId       string `json:"bot_id"`
	BotName     string `json:"bot_name"`
	Description string `json:"description"`
	IconUrl     string `json:"icon_url"`
	PublishTime string `json:"publish_time"`
}

type SpaceBots struct {
	SpaceBots []SpaceBot `json:"space_bots"`
	Total     int        `json:"total"`
}

type AgentList struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data SpaceBots `json:"data"`
}

func (c *Client) GetAgentList(spaceID string) (*AgentList, error) {
	client := curl.Config{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + c.Token,
		},
	}
	params := map[string]string{
		"space_id":   spaceID,
		"page_size":  strconv.Itoa(20),
		"page_index": strconv.Itoa(1),
	}
	resp, _, err := client.GET(agentListUri, params)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var result AgentList
	err = json.Unmarshal(resp, &result)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &result, nil
}
