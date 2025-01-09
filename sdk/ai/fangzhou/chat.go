package fangzhou

import (
	"context"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type Config struct {
	ApiKey string
}

type Client struct {
	Config
}

func NewChat(config Config) *Client {
	return &Client{
		Config: config,
	}
}

func (c *Client) Chat(endpointId string, messages []*model.ChatCompletionMessage) ([]*model.ChatCompletionChoice, error) {
	client := arkruntime.NewClientWithApiKey(
		c.ApiKey,
	)
	// 创建一个上下文，通常用于传递请求的上下文信息，如超时、取消等
	ctx := context.Background()
	// 构建聊天完成请求，设置请求的模型和消息内容
	req := model.ChatCompletionRequest{
		Model:    endpointId,
		Messages: messages,
	}
	// 发送聊天完成请求，并将结果存储在 resp 中，将可能出现的错误存储在 err 中
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Error("standard chat error: %v", err)
		return nil, err
	}
	return resp.Choices, nil
}
