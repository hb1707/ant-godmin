package aliyun

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
)

type TokenResult struct {
	ErrMsg string
	Token  struct {
		UserId     string
		Id         string
		ExpireTime int64
	}
}

func (c *Client) CreateToken(domain string) (*TokenResult, error) {
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
		return &tr, nil
	}
}
