package rdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hb1707/ant-godmin/setting"
	"github.com/redis/go-redis/v9"
)

var (
	// ErrClientNotInitialized Redis 客户端未初始化
	ErrClientNotInitialized = errors.New("redis client not initialized")
)

// Client 全局 Redis 客户端实例
var Client *redis.Client

// ClientRedis Redis 客户端封装，提供链式调用
type ClientRedis struct {
	client *redis.Client
	ctx    context.Context
}

// defaultContext 默认上下文（5秒超时）
func defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// Config Redis 配置选项
type Config struct {
	Addr         string
	Username     string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultConfig 默认配置
func DefaultConfig(addr, username, password string, db int) *Config {
	return &Config{
		Addr:         addr,
		Username:     username,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// InitRedis 初始化 Redis 客户端（简化版）
func InitRedis() error {
	addr := fmt.Sprintf("%s:%d", setting.Redis.Host, setting.Redis.Port)
	username := setting.Redis.Username
	password := setting.Redis.Password
	db := setting.Redis.DB
	return InitRedisWithConfig(DefaultConfig(addr, username, password, db))
}

// InitRedisWithConfig 使用配置初始化 Redis 客户端
func InitRedisWithConfig(cfg *Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  4 * time.Second,
	})

	// 验证连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	return nil
}

// CloseRedis 关闭 Redis 客户端连接
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// GetRedisClient 获取 Redis 客户端实例
func GetRedisClient() *redis.Client {
	return Client
}

// New 创建一个新的 ClientRedis 实例，用于链式调用
func New() *ClientRedis {
	ctx, _ := defaultContext()
	return &ClientRedis{
		client: Client,
		ctx:    ctx,
	}
}

// WithContext 设置自定义上下文
func (r *ClientRedis) WithContext(ctx context.Context) *ClientRedis {
	r.ctx = ctx
	return r
}

// WithTimeout 设置超时时间
func (r *ClientRedis) WithTimeout(timeout time.Duration) *ClientRedis {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	r.ctx = ctx
	return r
}

// ========== 链式调用方法 ==========

// Set 设置键值对
func (r *ClientRedis) Set(key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取键值对
func (r *ClientRedis) Get(key string) (string, error) {
	if r.client == nil {
		return "", ErrClientNotInitialized
	}
	return r.client.Get(r.ctx, key).Result()
}

// Del 删除键
func (r *ClientRedis) Del(keys ...string) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.Del(r.ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *ClientRedis) Exists(keys ...string) (int64, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.Exists(r.ctx, keys...).Result()
}

// Expire 设置过期时间
func (r *ClientRedis) Expire(key string, expiration time.Duration) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// TTL 获取键的剩余过期时间
func (r *ClientRedis) TTL(key string) (time.Duration, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.TTL(r.ctx, key).Result()
}

// Incr 自增
func (r *ClientRedis) Incr(key string) (int64, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.Incr(r.ctx, key).Result()
}

// IncrBy 按指定值自增
func (r *ClientRedis) IncrBy(key string, value int64) (int64, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.IncrBy(r.ctx, key, value).Result()
}

// Decr 自减
func (r *ClientRedis) Decr(key string) (int64, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.Decr(r.ctx, key).Result()
}

// DecrBy 按指定值自减
func (r *ClientRedis) DecrBy(key string, value int64) (int64, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.DecrBy(r.ctx, key, value).Result()
}

// SetJSON 设置 JSON 对象
func (r *ClientRedis) SetJSON(key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	return r.client.Set(r.ctx, key, data, expiration).Err()
}

// GetJSON 获取 JSON 对象
func (r *ClientRedis) GetJSON(key string, dest interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// MGet 批量获取
func (r *ClientRedis) MGet(keys ...string) ([]interface{}, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.MGet(r.ctx, keys...).Result()
}

// MSet 批量设置（键值对交替出现）
func (r *ClientRedis) MSet(pairs ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.MSet(r.ctx, pairs...).Err()
}

// Keys 查找匹配的键
func (r *ClientRedis) Keys(pattern string) ([]string, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.Keys(r.ctx, pattern).Result()
}

// ========== Hash 操作 ==========

// HSet 设置 Hash 字段
func (r *ClientRedis) HSet(key string, values ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.HSet(r.ctx, key, values...).Err()
}

// HGet 获取 Hash 字段
func (r *ClientRedis) HGet(key, field string) (string, error) {
	if r.client == nil {
		return "", ErrClientNotInitialized
	}
	return r.client.HGet(r.ctx, key, field).Result()
}

// HGetAll 获取 Hash 所有字段
func (r *ClientRedis) HGetAll(key string) (map[string]string, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.HGetAll(r.ctx, key).Result()
}

// HDel 删除 Hash 字段
func (r *ClientRedis) HDel(key string, fields ...string) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.HDel(r.ctx, key, fields...).Err()
}

// HExists 检查 Hash 字段是否存在
func (r *ClientRedis) HExists(key, field string) (bool, error) {
	if r.client == nil {
		return false, ErrClientNotInitialized
	}
	return r.client.HExists(r.ctx, key, field).Result()
}

// ========== List 操作 ==========

// LPush 从左侧推入列表
func (r *ClientRedis) LPush(key string, values ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.LPush(r.ctx, key, values...).Err()
}

// RPush 从右侧推入列表
func (r *ClientRedis) RPush(key string, values ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.RPush(r.ctx, key, values...).Err()
}

// LPop 从左侧弹出列表元素
func (r *ClientRedis) LPop(key string) (string, error) {
	if r.client == nil {
		return "", ErrClientNotInitialized
	}
	return r.client.LPop(r.ctx, key).Result()
}

// RPop 从右侧弹出列表元素
func (r *ClientRedis) RPop(key string) (string, error) {
	if r.client == nil {
		return "", ErrClientNotInitialized
	}
	return r.client.RPop(r.ctx, key).Result()
}

// LRange 获取列表范围内的元素
func (r *ClientRedis) LRange(key string, start, stop int64) ([]string, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.LRange(r.ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (r *ClientRedis) LLen(key string) (int64, error) {
	if r.client == nil {
		return 0, ErrClientNotInitialized
	}
	return r.client.LLen(r.ctx, key).Result()
}

// ========== Set 操作 ==========

// SAdd 添加集合成员
func (r *ClientRedis) SAdd(key string, members ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.SAdd(r.ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *ClientRedis) SMembers(key string) ([]string, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.SMembers(r.ctx, key).Result()
}

// SIsMember 检查是否是集合成员
func (r *ClientRedis) SIsMember(key string, member interface{}) (bool, error) {
	if r.client == nil {
		return false, ErrClientNotInitialized
	}
	return r.client.SIsMember(r.ctx, key, member).Result()
}

// SRem 移除集合成员
func (r *ClientRedis) SRem(key string, members ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.SRem(r.ctx, key, members...).Err()
}

// ========== ZSet 操作 ==========

// ZAdd 添加有序集合成员
func (r *ClientRedis) ZAdd(key string, members ...redis.Z) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (r *ClientRedis) ZRange(key string, start, stop int64) ([]string, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.ZRange(r.ctx, key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合范围内的成员（带分数）
func (r *ClientRedis) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	if r.client == nil {
		return nil, ErrClientNotInitialized
	}
	return r.client.ZRangeWithScores(r.ctx, key, start, stop).Result()
}

// ZRem 移除有序集合成员
func (r *ClientRedis) ZRem(key string, members ...interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.ZRem(r.ctx, key, members...).Err()
}

// ========== 发布订阅操作 ==========

// Publish 发布消息到指定频道
func (r *ClientRedis) Publish(channel string, message interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	return r.client.Publish(r.ctx, channel, message).Err()
}

// PublishJSON 发布 JSON 格式消息到指定频道
func (r *ClientRedis) PublishJSON(channel string, message interface{}) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	return r.client.Publish(r.ctx, channel, data).Err()
}

// Subscribe 订阅一个或多个频道
func (r *ClientRedis) Subscribe(channels ...string) *redis.PubSub {
	if r.client == nil {
		return nil
	}
	return r.client.Subscribe(r.ctx, channels...)
}

// PSubscribe 订阅一个或多个模式
func (r *ClientRedis) PSubscribe(patterns ...string) *redis.PubSub {
	if r.client == nil {
		return nil
	}
	return r.client.PSubscribe(r.ctx, patterns...)
}

// SubscribeHandler 订阅频道的消息处理器类型
type SubscribeHandler func(channel string, message string)

// SubscribeWithHandler 订阅频道并使用处理器处理消息
func (r *ClientRedis) SubscribeWithHandler(ctx context.Context, handler SubscribeHandler, channels ...string) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}

	pubsub := r.client.Subscribe(ctx, channels...)
	defer func() {
		_ = pubsub.Close()
	}()

	// 等待订阅确认
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive subscription confirmation: %w", err)
	}

	// 接收消息
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg != nil {
				handler(msg.Channel, msg.Payload)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// PSubscribeWithHandler 订阅模式并使用处理器处理消息
func (r *ClientRedis) PSubscribeWithHandler(ctx context.Context, handler SubscribeHandler, patterns ...string) error {
	if r.client == nil {
		return ErrClientNotInitialized
	}

	pubsub := r.client.PSubscribe(ctx, patterns...)
	defer func() {
		_ = pubsub.Close()
	}()

	// 等待订阅确认
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive subscription confirmation: %w", err)
	}

	// 接收消息
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg != nil {
				handler(msg.Channel, msg.Payload)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ========== 全局便捷函数 ==========

// Set 设置键值对（永不过期）
func Set(key string, value interface{}) error {
	return New().Set(key, value, 0)
}

// SetWithTTL 设置键值对（带过期时间）
func SetWithTTL(key string, value interface{}, expiration time.Duration) error {
	return New().Set(key, value, expiration)
}

// Get 获取键值对
func Get(key string) (string, error) {
	return New().Get(key)
}

// Del 删除键
func Del(keys ...string) error {
	return New().Del(keys...)
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	count, err := New().Exists(key)
	return count > 0, err
}

// Incr 自增键值
func Incr(key string) (int64, error) {
	return New().Incr(key)
}

// Decr 自减键值
func Decr(key string) (int64, error) {
	return New().Decr(key)
}

// IsConnected 检查 Redis 是否连接成功
func IsConnected() bool {
	if Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Client.Ping(ctx).Result()
	return err == nil
}

// SetJSON 设置 JSON 对象
func SetJSON(key string, value interface{}, expiration time.Duration) error {
	return New().SetJSON(key, value, expiration)
}

// GetJSON 获取 JSON 对象
func GetJSON(key string, dest interface{}) error {
	return New().GetJSON(key, dest)
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return New().Expire(key, expiration)
}

// TTL 获取键的剩余过期时间
func TTL(key string) (time.Duration, error) {
	return New().TTL(key)
}
