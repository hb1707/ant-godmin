# Redis 发布订阅功能使用文档

## 概述

本 Redis 客户端提供了完整的发布订阅（Pub/Sub）功能，支持：

- 简单消息发布与订阅
- JSON 格式消息发布与订阅
- 模式匹配订阅
- 自定义消息处理器
- 链式调用

## 初始化

首先需要初始化 Redis 客户端：

```go
import "your-project/cache/rdb"

// 简单初始化
err := rdb.InitRedis("localhost:6379", "", 0)

// 或使用配置初始化
config := rdb.DefaultConfig("localhost:6379", "", 0)
config.PoolSize = 20
err := rdb.InitRedisWithConfig(config)
```

## 基础用法

### 1. 发布消息

#### 发布简单字符串消息

```go
// 全局函数
err := rdb.Publish("my-channel", "Hello, Redis!")

// 链式调用
err := rdb.New().Publish("my-channel", "Hello, Redis!")

// 带自定义超时
err := rdb.New().WithTimeout(3 * time.Second).Publish("my-channel", "Hello!")
```

#### 发布 JSON 消息

```go
type Message struct {
Type    string    `json:"type"`
Content string    `json:"content"`
Time    time.Time `json:"time"`
}

msg := Message{
Type:    "notification",
Content: "这是一条通知消息",
Time:    time.Now(),
}

// 自动序列化为 JSON
err := rdb.PublishJSON("notification-channel", msg)
```

### 2. 订阅消息

#### 基础订阅（手动处理）

```go
// 订阅一个或多个频道
pubsub := rdb.Subscribe("channel1", "channel2")
defer pubsub.Close()

// 接收消息
ch := pubsub.Channel()
for msg := range ch {
fmt.Printf("频道: %s, 消息: %s\n", msg.Channel, msg.Payload)
}
```

#### 使用处理器订阅（推荐）

```go
ctx := context.Background()

// 定义消息处理器
handler := func(channel string, message string) {
fmt.Printf("收到消息 - 频道: %s, 内容: %s\n", channel, message)

// 处理业务逻辑...
}

// 订阅并自动处理消息
err := rdb.SubscribeWithHandler(ctx, handler, "channel1", "channel2")
```

#### 带超时的订阅

```go
// 订阅 30 秒后自动停止
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

handler := func (channel string, message string) {
fmt.Printf("消息: %s\n", message)
}

err := rdb.SubscribeWithHandler(ctx, handler, "my-channel")
```

### 3. 模式订阅

订阅匹配特定模式的所有频道：

```go
ctx := context.Background()

handler := func(channel string, message string) {
fmt.Printf("频道 [%s]: %s\n", channel, message)
}

// 订阅所有以 "user:" 开头的频道
err := rdb.PSubscribeWithHandler(ctx, handler, "user:*")

// 订阅多个模式
err := rdb.PSubscribeWithHandler(ctx, handler, "user:*", "order:*", "log:*")
```

## 实际应用场景

### 场景 1：用户登录通知

```go
// 发布端
type LoginEvent struct {
UserID    string    `json:"user_id"`
IP        string    `json:"ip"`
Timestamp time.Time `json:"timestamp"`
}

func NotifyUserLogin(userID, ip string) error {
event := LoginEvent{
UserID:    userID,
IP:        ip,
Timestamp: time.Now(),
}
return rdb.PublishJSON("events:login", event)
}

// 订阅端
func StartLoginMonitor() {
ctx := context.Background()

handler := func (channel string, message string) {
var event LoginEvent
if err := json.Unmarshal([]byte(message), &event); err != nil {
log.Printf("解析登录事件失败: %v", err)
return
}

log.Printf("用户 %s 从 %s 登录", event.UserID, event.IP)
// 执行安全检查、记录日志等...
}

go func () {
if err := rdb.SubscribeWithHandler(ctx, handler, "events:login"); err != nil {
log.Printf("订阅登录事件失败: %v", err)
}
}()
}
```

### 场景 2：缓存失效通知

```go
// 发布端 - 当数据更新时通知缓存失效
func InvalidateCache(key string) error {
return rdb.Publish("cache:invalidate", key)
}

// 订阅端 - 监听缓存失效并清理本地缓存
func StartCacheInvalidationListener() {
ctx := context.Background()

handler := func (channel string, message string) {
cacheKey := message
log.Printf("清理缓存: %s", cacheKey)

// 从本地缓存中删除
localCache.Delete(cacheKey)
}

go func () {
if err := rdb.SubscribeWithHandler(ctx, handler, "cache:invalidate"); err != nil {
log.Printf("订阅缓存失效失败: %v", err)
}
}()
}
```

### 场景 3：实时聊天系统

```go
// 发布端 - 发送聊天消息
type ChatMessage struct {
From    string    `json:"from"`
To      string    `json:"to"`
Content string    `json:"content"`
Time    time.Time `json:"time"`
}

func SendChatMessage(from, to, content string) error {
msg := ChatMessage{
From:    from,
To:      to,
Content: content,
Time:    time.Now(),
}

// 发送到接收者的频道
channel := fmt.Sprintf("chat:%s", to)
return rdb.PublishJSON(channel, msg)
}

// 订阅端 - 接收聊天消息
func ListenUserMessages(userID string) {
ctx := context.Background()

handler := func (channel string, message string) {
var msg ChatMessage
if err := json.Unmarshal([]byte(message), &msg); err != nil {
log.Printf("解析消息失败: %v", err)
return
}

log.Printf("收到来自 %s 的消息: %s", msg.From, msg.Content)
// 推送给客户端...
}

channel := fmt.Sprintf("chat:%s", userID)
go func () {
if err := rdb.SubscribeWithHandler(ctx, handler, channel); err != nil {
log.Printf("订阅聊天消息失败: %v", err)
}
}()
}
```

### 场景 4：系统日志收集

```go
// 使用模式订阅收集所有日志
func StartLogCollector() {
ctx := context.Background()

handler := func (channel string, message string) {
// channel 格式: log:service_name:level
parts := strings.Split(channel, ":")
if len(parts) >= 3 {
service := parts[1]
level := parts[2]

log.Printf("[%s][%s] %s", service, level, message)
// 写入日志文件或发送到日志服务...
}
}

// 订阅所有日志频道
go func () {
if err := rdb.PSubscribeWithHandler(ctx, handler, "log:*:*"); err != nil {
log.Printf("订阅日志失败: %v", err)
}
}()
}

// 各服务发布日志
func LogError(service, message string) {
channel := fmt.Sprintf("log:%s:error", service)
_ = rdb.Publish(channel, message)
}

func LogInfo(service, message string) {
channel := fmt.Sprintf("log:%s:info", service)
_ = rdb.Publish(channel, message)
}
```

## 后台服务示例

创建一个完整的后台订阅服务：

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"your-project/cache/rdb"
)

func main() {
	// 初始化 Redis
	if err := rdb.InitRedis("localhost:6379", "", 0); err != nil {
		log.Fatalf("初始化 Redis 失败: %v", err)
	}
	defer rdb.CloseRedis()

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup 用于等待所有订阅结束
	var wg sync.WaitGroup

	// 启动登录事件监听
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler := func(channel string, message string) {
			log.Printf("[LOGIN] %s", message)
		}
		if err := rdb.SubscribeWithHandler(ctx, handler, "events:login"); err != nil {
			log.Printf("登录事件订阅失败: %v", err)
		}
	}()

	// 启动缓存失效监听
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler := func(channel string, message string) {
			log.Printf("[CACHE] Invalidate: %s", message)
		}
		if err := rdb.SubscribeWithHandler(ctx, handler, "cache:invalidate"); err != nil {
			log.Printf("缓存失效订阅失败: %v", err)
		}
	}()

	// 启动聊天消息监听（模式匹配）
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler := func(channel string, message string) {
			log.Printf("[CHAT][%s] %s", channel, message)
		}
		if err := rdb.PSubscribeWithHandler(ctx, handler, "chat:*"); err != nil {
			log.Printf("聊天消息订阅失败: %v", err)
		}
	}()

	log.Println("订阅服务已启动，按 Ctrl+C 停止...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("正在停止订阅服务...")
	cancel()
	wg.Wait()
	log.Println("订阅服务已停止")
}
```

## API 参考

### 全局函数

| 函数                                                 | 说明                     |
|----------------------------------------------------|------------------------|
| `Publish(channel, message)`                        | 发布消息到指定频道              |
| `PublishJSON(channel, message)`                    | 发布 JSON 消息到指定频道        |
| `Subscribe(channels...)`                           | 订阅一个或多个频道，返回 PubSub 对象 |
| `PSubscribe(patterns...)`                          | 订阅匹配模式的频道，返回 PubSub 对象 |
| `SubscribeWithHandler(ctx, handler, channels...)`  | 订阅并使用处理器自动处理消息         |
| `PSubscribeWithHandler(ctx, handler, patterns...)` | 模式订阅并使用处理器自动处理消息       |

### 链式调用方法

```go
client := rdb.New()
client.WithContext(ctx).Publish(channel, message)
client.WithTimeout(3*time.Second).PublishJSON(channel, message)
```

## 注意事项

1. **Context 管理**：订阅是阻塞操作，使用 context 来控制订阅的生命周期
2. **错误处理**：订阅可能因网络问题中断，需要适当的重连机制
3. **消息可靠性**：Redis Pub/Sub 不保证消息持久化，订阅者不在线时消息会丢失
4. **性能考虑**：大量订阅可能影响 Redis 性能，合理规划频道设计
5. **Goroutine 管理**：订阅通常在 goroutine 中运行，注意避免 goroutine 泄漏

## 测试

运行测试：

```bash
go test -v ./cache/rdb -run TestPublish
go test -v ./cache/rdb -run TestSubscribe
```

运行基准测试：

```bash
go test -v ./cache/rdb -bench=BenchmarkPublish -benchmem
```

