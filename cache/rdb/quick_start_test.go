package rdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"
)

// Message 示例消息结构
type Message struct {
	ID      int       `json:"id"`
	Type    string    `json:"type"`
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

func QuickTest(t *testing.T) {
	// 1. 初始化 Redis 客户端
	if err := InitRedis("localhost:6379", "", 0); err != nil {
		log.Fatalf("初始化 Redis 失败: %v", err)
	}
	defer func() {
		_ = CloseRedis()
	}()

	fmt.Println("Redis 客户端初始化成功!")

	// 创建一个 context，用于控制订阅生命周期
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建 Redis 客户端实例
	rdb := New()

	// 2. 启动订阅者（在 goroutine 中运行）
	go startSubscriber(ctx)

	// 等待订阅者准备好
	time.Sleep(500 * time.Millisecond)

	// 3. 发布简单字符串消息
	fmt.Println("\n=== 发布简单消息 ===")
	if err := rdb.Publish("test-channel", "Hello, Redis Pub/Sub!"); err != nil {
		log.Printf("发布消息失败: %v", err)
	} else {
		fmt.Println("✓ 消息发布成功")
	}

	time.Sleep(500 * time.Millisecond)

	// 4. 发布 JSON 消息
	fmt.Println("\n=== 发布 JSON 消息 ===")
	msg := Message{
		ID:      1,
		Type:    "notification",
		Content: "这是一条 JSON 格式的消息",
		Time:    time.Now(),
	}

	if err := rdb.PublishJSON("json-channel", msg); err != nil {
		log.Printf("发布 JSON 消息失败: %v", err)
	} else {
		fmt.Println("✓ JSON 消息发布成功")
	}

	time.Sleep(500 * time.Millisecond)

	// 5. 发布到模式匹配的频道
	fmt.Println("\n=== 发布到模式频道 ===")
	channels := []string{"user:1001:login", "user:1002:logout", "order:2001:created"}
	for _, ch := range channels {
		if err := rdb.Publish(ch, fmt.Sprintf("事件来自频道: %s", ch)); err != nil {
			log.Printf("发布到 %s 失败: %v", ch, err)
		} else {
			fmt.Printf("✓ 已发布到: %s\n", ch)
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 6. 发布聊天消息示例
	fmt.Println("\n=== 发布聊天消息 ===")
	chatMsg := map[string]interface{}{
		"from":    "Alice",
		"to":      "Bob",
		"message": "你好，Bob！",
		"time":    time.Now(),
	}

	if err := rdb.PublishJSON("chat:bob", chatMsg); err != nil {
		log.Printf("发布聊天消息失败: %v", err)
	} else {
		fmt.Println("✓ 聊天消息发布成功")
	}

	// 等待所有消息被处理
	time.Sleep(2 * time.Second)
	fmt.Println("\n程序将在 30 秒后退出（或按 Ctrl+C 提前退出）...")
	time.Sleep(25 * time.Second)
}

// startSubscriber 启动订阅者，监听多个频道
func startSubscriber(ctx context.Context) {
	fmt.Println("启动订阅者...")

	// 订阅 1: 监听简单消息频道
	go func() {
		handler := func(channel string, message string) {
			fmt.Printf("[订阅1] 收到消息 - 频道: %s, 内容: %s\n", channel, message)
		}
		if err := New().SubscribeWithHandler(ctx, handler, "test-channel"); err != nil {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("订阅 test-channel 失败: %v", err)
			}
		}
	}()

	// 订阅 2: 监听 JSON 消息频道
	go func() {
		handler := func(channel string, message string) {
			var msg Message
			if err := json.Unmarshal([]byte(message), &msg); err != nil {
				log.Printf("解析 JSON 失败: %v", err)
				return
			}
			fmt.Printf("[订阅2] JSON 消息 - 频道: %s, ID: %d, 类型: %s, 内容: %s\n",
				channel, msg.ID, msg.Type, msg.Content)
		}
		if err := New().SubscribeWithHandler(ctx, handler, "json-channel"); err != nil {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("订阅 json-channel 失败: %v", err)
			}
		}
	}()

	// 订阅 3: 使用模式匹配监听用户事件
	go func() {
		handler := func(channel string, message string) {
			fmt.Printf("[订阅3] 用户事件 - 频道: %s, 消息: %s\n", channel, message)
		}
		if err := New().PSubscribeWithHandler(ctx, handler, "user:*"); err != nil {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("订阅 user:* 模式失败: %v", err)
			}
		}
	}()

	// 订阅 4: 使用模式匹配监听订单事件
	go func() {
		handler := func(channel string, message string) {
			fmt.Printf("[订阅4] 订单事件 - 频道: %s, 消息: %s\n", channel, message)
		}
		if err := New().PSubscribeWithHandler(ctx, handler, "order:*"); err != nil {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("订阅 order:* 模式失败: %v", err)
			}
		}
	}()

	// 订阅 5: 监听聊天消息
	go func() {
		handler := func(channel string, message string) {
			var chatMsg map[string]interface{}
			if err := json.Unmarshal([]byte(message), &chatMsg); err != nil {
				log.Printf("解析聊天消息失败: %v", err)
				return
			}
			fmt.Printf("[订阅5] 聊天消息 - 频道: %s, 从: %v, 到: %v, 内容: %v\n",
				channel, chatMsg["from"], chatMsg["to"], chatMsg["message"])
		}
		if err := New().PSubscribeWithHandler(ctx, handler, "chat:*"); err != nil {
			if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
				log.Printf("订阅 chat:* 模式失败: %v", err)
			}
		}
	}()

	fmt.Println("所有订阅者已启动")
}
