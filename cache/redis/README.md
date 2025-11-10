# Redis 客户端使用指南

这是一个优雅的 Redis 客户端封装，支持多种调用方式，满足不同场景需求。

## 特性

- ✅ 支持链式调用
- ✅ 自动管理 context
- ✅ JSON 序列化支持
- ✅ 完整的 Redis 数据类型支持（String、Hash、List、Set、ZSet）
- ✅ 批量操作
- ✅ 连接池管理
- ✅ 向后兼容

## 初始化

### 简单初始化

```go
import "your_project/cache/redis"

// 简单初始化
err := redis.InitRedis("localhost:6379", "", 0)
if err != nil {
log.Fatal(err)
}
defer redis.CloseRedis()
```

### 使用配置初始化

```go
cfg := redis.DefaultConfig("localhost:6379", "", 0)
cfg.PoolSize = 20
cfg.MinIdleConns = 10

err := redis.InitRedisWithConfig(cfg)
if err != nil {
log.Fatal(err)
}
```

## 使用方式

### 方式一：简洁调用（推荐）

最简单的使用方式，无需传递 context：

```go
// 字符串操作
redis.Set("key", "value") // 永不过期
redis.SetWithTTL("key", "value", 10*time.Minute) // 10分钟过期
value, err := redis.Get("key")
redis.Del("key1", "key2")

// 计数器
count, _ := redis.Incr("counter")
redis.Decr("counter")

// JSON 对象
type User struct {
Name string
Age  int
}
user := User{Name: "张三", Age: 25}
redis.SetJSON("user:1", user, time.Hour)

var loadedUser User
redis.GetJSON("user:1", &loadedUser)

// 过期时间管理
redis.Expire("key", 5*time.Minute)
ttl, _ := redis.TTL("key")
```

### 方式二：链式调用（灵活）

支持自定义 context 和超时：

```go
// 基本使用
err := redis.New().Set("key", "value", time.Minute)
value, err := redis.New().Get("key")

// 自定义超时
redis.New().WithTimeout(10*time.Second).Set("key", "value", 0)

// 使用自定义 context
ctx := context.WithValue(context.Background(), "trace_id", "123")
redis.New().WithContext(ctx).Set("key", "value", 0)

// 链式调用多个操作
client := redis.New().WithTimeout(5*time.Second)
client.Set("key1", "value1", 0)
client.Set("key2", "value2", 0)
```

## 完整功能示例

### Hash 操作

```go
client := redis.New()

// 设置 Hash 字段
client.HSet("user:1", "name", "张三", "age", "25")

// 获取单个字段
name, _ := client.HGet("user:1", "name")

// 获取所有字段
fields, _ := client.HGetAll("user:1")

// 检查字段是否存在
exists, _ := client.HExists("user:1", "name")

// 删除字段
client.HDel("user:1", "age")
```

### List 操作

```go
client := redis.New()

// 推入元素
client.RPush("queue", "task1", "task2", "task3")
client.LPush("stack", "item1", "item2")

// 弹出元素
task, _ := client.LPop("queue")
item, _ := client.RPop("stack")

// 获取列表范围
items, _ := client.LRange("queue", 0, -1) // 获取所有

// 获取列表长度
length, _ := client.LLen("queue")
```

### Set 操作

```go
client := redis.New()

// 添加成员
client.SAdd("tags", "golang", "redis", "docker")

// 获取所有成员
members, _ := client.SMembers("tags")

// 检查成员是否存在
exists, _ := client.SIsMember("tags", "golang")

// 移除成员
client.SRem("tags", "docker")
```

### ZSet（有序集合）操作

```go
client := redis.New()

// 添加成员（带分数）
client.ZAdd("leaderboard",
goredis.Z{Score: 100, Member: "player1"},
goredis.Z{Score: 200, Member: "player2"},
goredis.Z{Score: 150, Member: "player3"},
)

// 获取排名（从低到高）
players, _ := client.ZRange("leaderboard", 0, -1)

// 获取排名（带分数）
playersWithScores, _ := client.ZRangeWithScores("leaderboard", 0, -1)

// 移除成员
client.ZRem("leaderboard", "player1")
```

### 批量操作

```go
client := redis.New()

// 批量设置
client.MSet("key1", "value1", "key2", "value2", "key3", "value3")

// 批量获取
values, _ := client.MGet("key1", "key2", "key3")

// 查找匹配的键
keys, _ := client.Keys("user:*")

// 批量删除
client.Del("key1", "key2", "key3")
```

### 高级用法

```go
// 检查连接状态
if redis.IsConnected() {
fmt.Println("Redis 已连接")
}

// 检查键是否存在
exists, _ := redis.Exists("key")

// 自增指定值
newValue, _ := redis.New().IncrBy("counter", 10)
newValue, _ := redis.New().DecrBy("counter", 5)

// 获取原始客户端（用于更底层的操作）
rawClient := redis.GetRedisClient()
```

## 实际应用场景

### 场景1：用户会话管理

```go
// 保存用户会话
type UserSession struct {
UserID   int
Username string
LoginAt  time.Time
}

session := UserSession{
UserID:   1001,
Username: "张三",
LoginAt:  time.Now(),
}

// 30分钟过期
redis.SetJSON("session:"+sessionID, session, 30*time.Minute)

// 读取会话
var loadedSession UserSession
if err := redis.GetJSON("session:"+sessionID, &loadedSession); err != nil {
// 会话不存在或已过期
}
```

### 场景2：分布式锁

```go
func AcquireLock(key string, ttl time.Duration) bool {
client := redis.New()
// NX 表示键不存在时才设置
err := client.Set(key, "locked", ttl)
return err == nil
}

func ReleaseLock(key string) {
redis.Del(key)
}

// 使用
if AcquireLock("order:123:lock", 10*time.Second) {
defer ReleaseLock("order:123:lock")
// 执行业务逻辑
}
```

### 场景3：限流器

```go
func CheckRateLimit(userID string, limit int64, window time.Duration) bool {
key := fmt.Sprintf("ratelimit:%s", userID)
client := redis.New()

count, err := client.Incr(key)
if err != nil {
return false
}

// 第一次访问，设置过期时间
if count == 1 {
client.Expire(key, window)
}

return count <= limit
}

// 使用：每分钟最多100次请求
if !CheckRateLimit(userID, 100, time.Minute) {
// 返回限流错误
}
```

### 场景4：排行榜

```go
// 更新玩家分数
func UpdateScore(playerID string, score float64) {
redis.New().ZAdd("game:leaderboard", goredis.Z{
Score:  score,
Member: playerID,
})
}

// 获取前10名
func GetTopPlayers(n int64) ([]string, error) {
// ZRevRange 从高到低排序
return redis.New().ZRange("game:leaderboard", 0, n-1)
}
```

### 场景5：缓存查询结果

```go
func GetUser(userID int) (*User, error) {
cacheKey := fmt.Sprintf("user:%d", userID)

// 先查缓存
var user User
err := redis.GetJSON(cacheKey, &user)
if err == nil {
return &user, nil
}

// 缓存未命中，查询数据库
user, err = db.QueryUser(userID)
if err != nil {
return nil, err
}

// 写入缓存，1小时过期
redis.SetJSON(cacheKey, user, time.Hour)
return &user, nil
}
```

## 错误处理

```go
import goredis "github.com/redis/go-redis/v9"

value, err := redis.Get("key")
if err != nil {
if errors.Is(err, goredis.Nil) {
// 键不存在
} else if errors.Is(err, redis.ErrClientNotInitialized) {
// 客户端未初始化
} else {
// 其他错误
}
}
```

## 性能优化建议

1. **使用批量操作**：优先使用 `MGet`/`MSet` 而不是多次调用 `Get`/`Set`
2. **合理设置过期时间**：避免 Redis 内存溢出
3. **使用连接池**：通过 `Config` 配置合适的连接池大小
4. **避免大键值**：单个键值建议不超过 10MB
5. **使用 Pipeline**：对于大量操作，可通过 `GetRedisClient()` 获取原始客户端使用 Pipeline

## 线程安全

所有方法都是线程安全的，可以在多个 goroutine 中并发使用。

