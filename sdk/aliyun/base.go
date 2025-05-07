package aliyun

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"net/http"
	"time"
)

var AccessKeyId = ""
var AccessKeySecret = ""

type Client struct {
	*sdk.Client
}

func NewClient(regionId string) *Client {
	c := sdk.NewConfig()
	c.HttpTransport = &http.Transport{
		IdleConnTimeout: 10 * time.Second,
	}
	c.EnableAsync = true
	c.GoRoutinePoolSize = 1
	c.MaxTaskQueueSize = 1
	c.Timeout = 10 * time.Second
	credential := credentials.NewAccessKeyCredential(AccessKeyId, AccessKeySecret)
	client, err := sdk.NewClientWithOptions(regionId, c, credential)
	if err != nil {
		panic(err)
	}
	return &Client{
		Client: client,
	}
}
