package fangzhou

import (
	"context"
	"fmt"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"io"
)

type Config struct {
	ApiKey string
	ApiAk  string
	ApiSk  string
}

type Client struct {
	Config
	BotStreamChan chan model.BotChatCompletionStreamResponse
	StreamChan    chan model.ChatCompletionStreamResponse
}

func NewChat(config Config) *Client {
	return &Client{
		Config:        config,
		BotStreamChan: make(chan model.BotChatCompletionStreamResponse),
		StreamChan:    make(chan model.ChatCompletionStreamResponse),
	}
}

func (c *Client) Chat(endpointId string, messages []*model.ChatCompletionMessage) ([]*model.ChatCompletionChoice, error) {
	client := arkruntime.NewClientWithAkSk(
		c.ApiAk, c.ApiSk,
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
	client := arkruntime.NewClientWithAkSk(
		c.ApiAk, c.ApiSk,
		//arkruntime.WithBaseUrl("https://api-knowledgebase.mlp.cn-beijing.volces.com/api/knowledge"),
	)
	// 创建一个上下文，通常用于传递请求的上下文信息，如超时、取消等
	ctx := context.Background()
	// 构建聊天完成请求，设置请求的模型和消息内容
	req := model.ChatCompletionRequest{
		Model:    endpointId,
		Messages: messages,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req) //arkruntime.WithCustomHeader("V-Account-Id", "2103628750"),

	if err != nil {
		log.Error("standard chat error: %v", err)
		return err
	}
	defer stream.Close()
	for {
		recv, err := stream.Recv()
		if err != nil && err != io.EOF {
			log.Error(fmt.Sprintf("standard chat error: %v", err))
			return err
		}
		// 超时处理
		select {
		case c.StreamChan <- recv:
		case <-ctx.Done():
			return nil
			//case <-time.After(5 * time.Second):
			//	log.Error("standard chat error: %v", err)
			//	return err
		}
	}
}

func (c *Client) ChatStreamAdd(text string) {
	var msg model.ChatCompletionStreamResponse
	msg.Choices = append(msg.Choices, &model.ChatCompletionStreamChoice{
		Delta: model.ChatCompletionStreamChoiceDelta{
			Content: text,
		},
	})
	c.StreamChan <- msg
}
