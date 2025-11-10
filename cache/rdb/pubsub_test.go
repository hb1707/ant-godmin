package rdb

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestPublishSubscribe 测试发布订阅基本功能
func TestPublishSubscribe(t *testing.T) {
	// 初始化 Redis（根据实际情况修改连接信息）
	if err := InitRedis("localhost:6379", "", 0); err != nil {
		t.Skipf("无法连接到 Redis: %v", err)
		return
	}
	defer func() {
		_ = CloseRedis()
	}()
	rds := New()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建一个通道来接收消息
	received := make(chan string, 1)

	// 启动订阅 goroutine
	go func() {
		handler := func(channel string, message string) {
			received <- message
		}
		_ = rds.SubscribeWithHandler(ctx, handler, "test-channel")
	}()

	// 等待订阅准备好
	time.Sleep(500 * time.Millisecond)

	// 发布消息
	testMsg := "Hello, Test!"
	if err := rds.Publish("test-channel", testMsg); err != nil {
		t.Fatalf("发布消息失败: %v", err)
	}

	// 接收消息
	select {
	case msg := <-received:
		if msg != testMsg {
			t.Errorf("期望收到 %s, 实际收到 %s", testMsg, msg)
		}
	case <-time.After(2 * time.Second):
		t.Error("超时未收到消息")
	}
}

// TestPublishSubscribeJSON 测试 JSON 消息发布订阅
func TestPublishSubscribeJSON(t *testing.T) {
	if err := InitRedis("localhost:6379", "", 0); err != nil {
		t.Skipf("无法连接到 Redis: %v", err)
		return
	}
	defer func() {
		_ = CloseRedis()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rds := New()

	type TestMessage struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
	}

	testMsg := TestMessage{ID: 1, Content: "测试消息"}
	received := make(chan string, 1)

	go func() {
		handler := func(channel string, message string) {
			received <- message
		}
		_ = rds.SubscribeWithHandler(ctx, handler, "json-channel")
	}()

	time.Sleep(500 * time.Millisecond)

	if err := rds.PublishJSON("json-channel", testMsg); err != nil {
		t.Fatalf("发布 JSON 消息失败: %v", err)
	}

	select {
	case msg := <-received:
		t.Logf("收到 JSON 消息: %s", msg)
	case <-time.After(2 * time.Second):
		t.Error("超时未收到消息")
	}
}

// TestPatternSubscribe 测试模式订阅
func TestPatternSubscribe(t *testing.T) {
	if err := InitRedis("localhost:6379", "", 0); err != nil {
		t.Skipf("无法连接到 Redis: %v", err)
		return
	}
	rds := New()
	defer func() {
		_ = CloseRedis()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	received := make(chan string, 3)

	go func() {
		handler := func(channel string, message string) {
			received <- fmt.Sprintf("%s:%s", channel, message)
		}
		_ = rds.PSubscribeWithHandler(ctx, handler, "test:*")
	}()

	time.Sleep(500 * time.Millisecond)

	// 发布到不同的匹配频道
	channels := []string{"test:1", "test:2", "test:abc"}
	for _, ch := range channels {
		if err := rds.Publish(ch, "message"); err != nil {
			t.Errorf("发布到 %s 失败: %v", ch, err)
		}
	}

	// 接收消息
	count := 0
	timeout := time.After(2 * time.Second)
	for count < len(channels) {
		select {
		case msg := <-received:
			t.Logf("收到消息: %s", msg)
			count++
		case <-timeout:
			t.Errorf("仅收到 %d/%d 条消息", count, len(channels))
			return
		}
	}
}

// BenchmarkPublish 基准测试：发布消息性能
func BenchmarkPublish(b *testing.B) {
	if err := InitRedis("localhost:6379", "", 0); err != nil {
		b.Skipf("无法连接到 Redis: %v", err)
		return
	}
	rds := New()
	defer func() {
		_ = CloseRedis()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rds.Publish("bench-channel", "benchmark message")
	}
}

// BenchmarkPublishJSON 基准测试：发布 JSON 消息性能
func BenchmarkPublishJSON(b *testing.B) {
	if err := InitRedis("localhost:6379", "", 0); err != nil {
		b.Skipf("无法连接到 Redis: %v", err)
		return
	}
	rds := New()
	defer func() {
		_ = CloseRedis()
	}()

	type BenchMessage struct {
		ID   int    `json:"id"`
		Data string `json:"data"`
	}

	msg := BenchMessage{ID: 1, Data: "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rds.PublishJSON("bench-channel", msg)
	}
}

// ExampleSubscribeWithHandler_userLogin 示例：用户登录事件订阅
func ExampleSubscribeWithHandler_userLogin() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rds := New()
	handler := func(channel string, message string) {
		fmt.Printf("用户登录: %s\n", message)
	}

	_ = rds.SubscribeWithHandler(ctx, handler, "events:login")
}

// ExamplePSubscribeWithHandler_chatMessages 示例：订阅所有聊天消息
func ExamplePSubscribeWithHandler_chatMessages() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	handler := func(channel string, message string) {
		fmt.Printf("聊天消息 [%s]: %s\n", channel, message)
	}
	rds := New()
	_ = rds.PSubscribeWithHandler(ctx, handler, "chat:*")
}
