# Redis 客户端 - 发布订阅功能

## 概述

本 Redis 客户端已完成优化，提供了完整的发布订阅（Pub/Sub）功能，代码更加优雅且易于使用。

## 主要特性

✅ **发布订阅功能**

- 简单字符串消息发布/订阅
- JSON 格式消息自动序列化/反序列化
- 模式匹配订阅（Pattern Subscribe）
- 自定义消息处理器
- 支持多频道同时订阅

✅ **优雅的 API 设计**

- 链式调用支持
- 全局便捷函数
- Context 生命周期管理
- 清晰的错误处理

✅ **高性能**

- 连接池管理
- 可配置的超时和重试
- 异步消息处理

## 快速开始

### 1. 安装依赖

```bash
go get github.com/redis/go-redis/v9
```

### 2. 初始化客户端

```go
import "your-project/cache/rdb"

// 简单初始化
err := rdb.InitRedis("localhost:6379", "", 0)
if err != nil {
    log.Fatal(err)
}
defer rdb.CloseRedis()
```

### 3. 发布消息

```go
// 发布简单消息
err := rdb.Publish("my-channel", "Hello, Redis!")

// 发布 JSON 消息
type Message struct {
    Content string    `json:"content"`
    Time    time.Time `json:"time"`
}
msg := Message{Content: "Hello", Time: time.Now()}
err := rdb.PublishJSON("json-channel", msg)
```

### 4. 订阅消息

```go
ctx := context.Background()

// 定义消息处理器
handler := func(channel string, message string) {
    fmt.Printf("收到消息: %s\n", message)
}

// 订阅频道
err := rdb.SubscribeWithHandler(ctx, handler, "my-channel")
```

### 5. 模式订阅

```go
// 订阅所有以 "user:" 开头的频道
handler := func(channel string, message string) {
    fmt.Printf("[%s] %s\n", channel, message)
}

err := rdb.PSubscribeWithHandler(ctx, handler, "user:*", "order:*")
```

## 文件说明

| 文件                  | 说明                        |
|---------------------|---------------------------|
| `redis.go`          | 核心实现，包含所有 Redis 操作和发布订阅功能 |
| `pubsub_example.go` | 详细的使用示例，展示各种应用场景          |
| `pubsub_test.go`    | 单元测试和基准测试                 |
| `quick_start.go`    | 快速开始示例，可直接运行              |
| `PUBSUB_USAGE.md`   | 完整的使用文档和 API 参考           |

## 运行示例

```bash
# 确保 Redis 服务已启动
redis-server

# 运行快速开始示例（需要修改 import 路径）
go run cache/rdb/quick_start.go

# 运行测试
go test -v ./cache/rdb -run TestPublish

# 运行基准测试
go test -v ./cache/rdb -bench=. -benchmem
```

## API 示例

### 基础操作

```go
// 发布消息
rdb.Publish("channel", "message")
rdb.PublishJSON("channel", struct{...})

// 订阅消息
rdb.Subscribe("channel1", "channel2")
rdb.PSubscribe("pattern:*")

// 带处理器的订阅（推荐）
rdb.SubscribeWithHandler(ctx, handler, "channel")
rdb.PSubscribeWithHandler(ctx, handler, "pattern:*")
```

### 链式调用

```go
// 自定义超时
rdb.New().WithTimeout(3*time.Second).Publish("channel", "msg")

// 自定义 Context
ctx := context.WithValue(context.Background(), "key", "value")
rdb.New().WithContext(ctx).PublishJSON("channel", data)
```

### 实际应用场景

#### 用户登录通知

```go
type LoginEvent struct {
    UserID string    `json:"user_id"`
    IP     string    `json:"ip"`
    Time   time.Time `json:"time"`
}

// 发布
event := LoginEvent{UserID: "123", IP: "1.2.3.4", Time: time.Now()}
rdb.PublishJSON("events:login", event)

// 订阅
handler := func(channel, message string) {
    var event LoginEvent
    json.Unmarshal([]byte(message), &event)
    // 处理登录事件...
}
rdb.SubscribeWithHandler(ctx, handler, "events:login")
```

#### 缓存失效通知

```go
// 数据更新时通知缓存失效
rdb.Publish("cache:invalidate", "user:123:profile")

// 监听并清理缓存
handler := func(channel, message string) {
    localCache.Delete(message)
}
rdb.SubscribeWithHandler(ctx, handler, "cache:invalidate")
```

#### 实时聊天系统

```go
// 发送消息给用户
type ChatMsg struct {
    From string `json:"from"`
    To   string `json:"to"`
    Msg  string `json:"msg"`
}

msg := ChatMsg{From: "alice", To: "bob", Msg: "Hi!"}
rdb.PublishJSON(fmt.Sprintf("chat:%s", msg.To), msg)

// 接收消息
handler := func(channel, message string) {
    var msg ChatMsg
    json.Unmarshal([]byte(message), &msg)
    // 推送给客户端...
}
rdb.PSubscribeWithHandler(ctx, handler, "chat:*")
```

## 优化说明

相比原始代码，本次优化包括：

1. **代码清理**
    - 移除了冗余代码和注释
    - 统一了错误处理方式
    - 优化了 Context 管理

2. **功能完善**
    - 完整实现了发布订阅功能
    - 添加了 JSON 消息支持
    - 支持模式匹配订阅
    - 提供了消息处理器模式

3. **更优雅的 API**
    - 链式调用支持
    - 全局便捷函数
    - 清晰的函数命名
    - 完整的类型安全

4. **文档完善**
    - 详细的使用示例
    - 实际应用场景
    - 单元测试和基准测试
    - 完整的 API 文档

## 注意事项

1. **订阅是阻塞操作**：通常需要在 goroutine 中运行
2. **Context 管理**：使用 Context 控制订阅的生命周期
3. **消息可靠性**：Redis Pub/Sub 不保证消息持久化
4. **错误处理**：订阅可能因网络问题中断，需要重连机制
5. **性能考虑**：合理设计频道结构，避免过多订阅

## 下一步

- 根据实际项目需求调整配置参数
- 实现重连机制（如需要）
- 添加消息确认机制（如需要）
- 集成到现有项目中

## 相关资源

- [Redis 官方文档 - Pub/Sub](https://redis.io/docs/manual/pubsub/)
- [go-redis 文档](https://redis.uptrace.dev/)
- [PUBSUB_USAGE.md](./PUBSUB_USAGE.md) - 完整使用文档

## 支持

如有问题或建议，请参考：

- `PUBSUB_USAGE.md` - 详细的使用文档
- `pubsub_example.go` - 各种场景示例
- `pubsub_test.go` - 单元测试示例

