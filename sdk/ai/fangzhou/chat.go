package fangzhou

import (
	"context"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"io"
)

type Config struct {
	ApiKey string
}

type Client struct {
	Config
	StreamChan chan model.ChatCompletionStreamResponse
}

func NewChat(config Config) *Client {
	return &Client{
		Config:     config,
		StreamChan: make(chan model.ChatCompletionStreamResponse),
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

func (c *Client) ChatStream(endpointId string, messages []*model.ChatCompletionMessage) error {
	client := arkruntime.NewClientWithApiKey(
		c.ApiKey,
	)
	// 创建一个上下文，通常用于传递请求的上下文信息，如超时、取消等
	ctx := context.Background()
	// 构建聊天完成请求，设置请求的模型和消息内容
	req := model.ChatCompletionRequest{
		Model:    endpointId,
		Messages: messages,
		Stream:   true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Error("standard chat error: %v", err)
		return err
	}
	defer stream.Close()
	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Error("standard chat error: %v", err)
			return err
		}
		c.StreamChan <- recv
	}
}
