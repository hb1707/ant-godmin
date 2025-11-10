package rdb_test

import (
	"fmt"
	"time"

	"github.com/hb1707/ant-godmin/cache/rdb" // 请替换为实际的包路径
)

// ExampleSimpleUsage 展示最简单的使用方式
func ExampleSimpleUsage() {
	// 初始化
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	// 基本操作
	rdb.Set("name", "张三")
	rdb.SetWithTTL("token", "abc123", 10*time.Minute)

	name, _ := rdb.Get("name")
	fmt.Println(name) // 输出: 张三

	// 计数器
	rdb.Incr("views")
	count, _ := rdb.Incr("views")
	fmt.Println(count) // 输出: 2
}

// ExampleChainedCalls 展示链式调用
func ExampleChainedCalls() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	// 使用自定义超时
	client := rdb.New().WithTimeout(10 * time.Second)
	client.Set("key1", "value1", 0)
	client.Set("key2", "value2", time.Hour)

	value, _ := client.Get("key1")
	fmt.Println(value)
}

// ExampleJSONOperations 展示 JSON 操作
func ExampleJSONOperations() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	// 保存 JSON 对象
	user := User{
		ID:       1001,
		Username: "zhangsan",
		Email:    "zhangsan@example.com",
	}
	rdb.SetJSON("user:1001", user, time.Hour)

	// 读取 JSON 对象
	var loadedUser User
	if err := rdb.GetJSON("user:1001", &loadedUser); err == nil {
		fmt.Printf("User: %s (%s)\n", loadedUser.Username, loadedUser.Email)
	}
}

// ExampleHashOperations 展示 Hash 操作
func ExampleHashOperations() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	client := rdb.New()

	// 设置用户信息
	client.HSet("user:1001",
		"username", "zhangsan",
		"email", "zhangsan@example.com",
		"age", "25",
	)

	// 获取单个字段
	username, _ := client.HGet("user:1001", "username")
	fmt.Println(username) // 输出: zhangsan

	// 获取所有字段
	fields, _ := client.HGetAll("user:1001")
	for k, v := range fields {
		fmt.Printf("%s: %s\n", k, v)
	}
}

// ExampleListOperations 展示 List 操作（队列）
func ExampleListOperations() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	client := rdb.New()

	// 作为队列使用
	client.RPush("tasks", "task1", "task2", "task3")

	// 处理任务
	for {
		task, err := client.LPop("tasks")
		if err != nil {
			break // 队列为空
		}
		fmt.Printf("处理任务: %s\n", task)
	}
}

// ExampleSetOperations 展示 Set 操作（标签）
func ExampleSetOperations() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	client := rdb.New()

	// 添加文章标签
	client.SAdd("article:100:tags", "golang", "redis", "docker")

	// 检查标签是否存在
	exists, _ := client.SIsMember("article:100:tags", "golang")
	fmt.Println(exists) // 输出: true

	// 获取所有标签
	tags, _ := client.SMembers("article:100:tags")
	fmt.Println(tags) // 输出: [golang redis docker]
}

// ExampleCaching 展示缓存模式
func ExampleCaching() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	type Article struct {
		ID      int
		Title   string
		Content string
	}

	// 模拟数据库查询函数
	getArticleFromDB := func(id int) (*Article, error) {
		// 实际从数据库查询
		return &Article{ID: id, Title: "文章标题", Content: "文章内容"}, nil
	}

	// 缓存读取函数
	getArticle := func(id int) (*Article, error) {
		cacheKey := fmt.Sprintf("article:%d", id)

		// 先查缓存
		var article Article
		err := rdb.GetJSON(cacheKey, &article)
		if err == nil {
			fmt.Println("从缓存获取")
			return &article, nil
		}

		// 缓存未命中，查询数据库
		fmt.Println("从数据库获取")
		article2, err := getArticleFromDB(id)
		if err != nil {
			return nil, err
		}

		// 写入缓存，1小时过期
		rdb.SetJSON(cacheKey, article2, time.Hour)
		return article2, nil
	}

	// 使用
	article, _ := getArticle(100)
	fmt.Println(article.Title)
}

// ExampleRateLimiter 展示限流器实现
func ExampleRateLimiter() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	// 检查用户是否超过访问限制
	checkRateLimit := func(userID string, maxRequests int64, window time.Duration) bool {
		key := fmt.Sprintf("ratelimit:%s", userID)
		client := rdb.New()

		count, err := client.Incr(key)
		if err != nil {
			return false
		}

		// 第一次访问，设置过期时间
		if count == 1 {
			client.Expire(key, window)
		}

		return count <= maxRequests
	}

	// 模拟用户请求
	userID := "user123"
	for i := 0; i < 5; i++ {
		// 每分钟最多3次请求
		if checkRateLimit(userID, 3, time.Minute) {
			fmt.Printf("请求 %d: 允许\n", i+1)
		} else {
			fmt.Printf("请求 %d: 限流\n", i+1)
		}
	}
}

// ExampleDistributedLock 展示分布式锁
func ExampleDistributedLock() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	acquireLock := func(key string, ttl time.Duration) bool {
		err := rdb.New().Set(key, "locked", ttl)
		return err == nil
	}

	releaseLock := func(key string) {
		rdb.Del(key)
	}

	// 使用分布式锁
	lockKey := "order:12345:lock"
	if acquireLock(lockKey, 10*time.Second) {
		defer releaseLock(lockKey)

		fmt.Println("获取锁成功，处理订单...")
		// 执行业务逻辑
		time.Sleep(1 * time.Second)
		fmt.Println("订单处理完成")
	} else {
		fmt.Println("获取锁失败，订单正在被其他进程处理")
	}
}

// ExampleBatchOperations 展示批量操作
func ExampleBatchOperations() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	client := rdb.New()

	// 批量设置
	client.MSet(
		"user:1:name", "张三",
		"user:2:name", "李四",
		"user:3:name", "王五",
	)

	// 批量获取
	values, _ := client.MGet("user:1:name", "user:2:name", "user:3:name")
	for i, v := range values {
		if v != nil {
			fmt.Printf("用户 %d: %v\n", i+1, v)
		}
	}

	// 查找匹配的键
	keys, _ := client.Keys("user:*:name")
	fmt.Printf("找到 %d 个用户\n", len(keys))
}

// ExampleSessionManagement 展示会话管理
func ExampleSessionManagement() {
	rdb.InitRedis("localhost:6379", "", 0)
	defer rdb.CloseRedis()

	type UserSession struct {
		UserID    int       `json:"user_id"`
		Username  string    `json:"username"`
		LoginTime time.Time `json:"login_time"`
		IP        string    `json:"ip"`
	}

	// 创建会话
	createSession := func(sessionID string, userID int, username, ip string) error {
		session := UserSession{
			UserID:    userID,
			Username:  username,
			LoginTime: time.Now(),
			IP:        ip,
		}
		// 会话30分钟过期
		return rdb.SetJSON("session:"+sessionID, session, 30*time.Minute)
	}

	// 获取会话
	getSession := func(sessionID string) (*UserSession, error) {
		var session UserSession
		err := rdb.GetJSON("session:"+sessionID, &session)
		if err != nil {
			return nil, err
		}
		return &session, nil
	}

	// 删除会话（登出）
	deleteSession := func(sessionID string) error {
		return rdb.Del("session:" + sessionID)
	}

	// 使用
	sessionID := "sess_abc123"
	createSession(sessionID, 1001, "zhangsan", "192.168.1.1")

	session, _ := getSession(sessionID)
	fmt.Printf("用户 %s 已登录\n", session.Username)

	deleteSession(sessionID)
	fmt.Println("用户已登出")
}
