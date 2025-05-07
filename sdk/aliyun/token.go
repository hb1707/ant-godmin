package aliyun

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"time"
)

type TokenResult struct {
	ErrMsg string
	Token  struct {
		UserId     string
		Id         string
		ExpireTime int64
	}
}

var tokenCache = make(map[string]TokenResult)

func (c *Client) CreateToken(domain string) (*TokenResult, error) {
	if token, ok := tokenCache[domain]; ok {
		if token.Token.ExpireTime-30 > time.Now().Unix() {
			return &token, nil
		}
	}
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Domain = domain //"nls-meta.cn-shanghai.aliyuncs.com"
	request.ApiName = "CreateToken"
	request.Version = "2019-02-28"
	response, err := c.ProcessCommonRequest(request)
	if err != nil {
		panic(err)
	}
	fmt.Print(response.GetHttpStatus())
	fmt.Print(response.GetHttpContentString())
	var tr TokenResult
	err = json.Unmarshal([]byte(response.GetHttpContentString()), &tr)
	if err != nil {
		return nil, err
	} else {
		tokenCache[domain] = tr
		return &tr, nil
	}
}
