package openai

import (
	"context"
	"fmt"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/sashabaranov/go-openai"
	"io"
)

type Client struct {
	*openai.Client
	StreamChan chan openai.ChatCompletionStreamResponse
}

func NewClient(baseUrl string, apiKey string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseUrl
	client := openai.NewClientWithConfig(config)
	return &Client{
		Client:     client,
		StreamChan: make(chan openai.ChatCompletionStreamResponse),
	}
}
func (c *Client) Chat(model string, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionResponse, error) {
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ChatStream(model string, messages []openai.ChatCompletionMessage) error {
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Error(err)
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
