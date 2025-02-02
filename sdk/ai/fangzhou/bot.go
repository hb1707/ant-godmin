package fangzhou

import (
	"context"
	"fmt"
	"github.com/hb1707/ant-godmin/pkg/log"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"io"
)

func (c *Client) BotChatStream(endpointId string, messages []*model.ChatCompletionMessage) error {
	client := arkruntime.NewClientWithAkSk(
		c.ApiAk, c.ApiSk,
		//arkruntime.WithBaseUrl("https://api-knowledgebase.mlp.cn-beijing.volces.com/api/knowledge"),
	)
	// 创建一个上下文，通常用于传递请求的上下文信息，如超时、取消等
	ctx := context.Background()
	// 构建聊天完成请求，设置请求的模型和消息内容
	req := model.BotChatCompletionRequest{
		Model:    endpointId,
		Messages: messages,
	}
	stream, err := client.CreateBotChatCompletionStream(ctx, req) //arkruntime.WithCustomHeader("V-Account-Id", "2103628750"),

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
		case c.StreamChan <- recv.ChatCompletionStreamResponse:
		case <-ctx.Done():
			return nil
			//case <-time.After(5 * time.Second):
			//	log.Error("standard chat error: %v", err)
			//	return err
		}
	}
}
