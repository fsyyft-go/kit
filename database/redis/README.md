# redis

## 简介

redis 包提供了 Go 语言下的高性能 Redis 客户端接口与扩展，支持原生命令、管道、事务、Lua 脚本、发布订阅、基础 KV 操作等，兼容 go-redis v9，适合缓存、分布式锁、消息队列等多种场景。

### 主要特性

- 统一 Redis 客户端接口，支持 Do/Pipelined/TxPipelined/Subscribe/PSubscribe
- 支持扩展接口（Get/Set/Del/Expire 等常用命令）
- 支持 Lua 脚本（Eval/EvalSha/ScriptLoad/ScriptExists 等）
- 支持发布订阅（PubSub）
- 支持 Option 配置（地址、密码等）
- 完善的错误处理与类型封装
- 完整单元测试覆盖

### 设计理念

redis 包遵循"简洁、兼容、易扩展"的设计理念，接口与 go-redis v9 保持一致，支持 Option 配置，便于集成与二次封装。扩展接口覆盖常用场景，底层类型与命令高度兼容官方客户端。

## 安装

### 前置条件

- Go 版本要求：Go 1.18+
- 依赖要求：
  - github.com/redis/go-redis/v9
  - github.com/stretchr/testify（仅测试）

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/database/redis
```

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "github.com/fsyyft-go/kit/database/redis"
    "time"
)

func main() {
    // 创建 Redis 客户端，默认 127.0.0.1:6379
    rdb := redis.NewRedis()
    ext := redis.NewRedisExtension(rdb)
    ctx := context.Background()
    // 基础 KV 操作
    ext.Set(ctx, "foo", "bar", time.Second*10)
    val, err := ext.Get(ctx, "foo").Result()
    if err != nil {
        panic(err)
    }
    fmt.Println("foo:", val)
}
```

### 管道与事务

```go
// 管道批量操作
_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
    pipe.Do(ctx, "SET", "k1", "v1")
    pipe.Do(ctx, "SET", "k2", "v2")
    return nil
})

// 事务管道
_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
    pipe.Do(ctx, "INCR", "counter")
    return nil
})
```

### Lua 脚本

```go
script := `return ARGV[1]`
sha, _ := rdb.ScriptLoad(ctx, script).Result()
res, _ := rdb.EvalSha(ctx, sha, []string{}, 123).Result()
fmt.Println(res)
```

### 发布订阅

```go
pubsub := rdb.Subscribe(ctx, "my-channel")
go func() {
    msg, _ := pubsub.ReceiveMessage(ctx)
    fmt.Println("收到消息:", msg.Payload)
}()
rdb.Do(ctx, "PUBLISH", "my-channel", "hello")
```

## 详细指南

### 核心概念

- **Redis 接口**：统一封装 go-redis v9，支持所有原生命令
- **扩展接口**：常用 KV 操作、过期、删除等
- **管道/事务**：批量高效操作，事务保证原子性
- **Lua 脚本**：支持 Eval/EvalSha/ScriptLoad/ScriptExists
- **发布订阅**：支持多频道订阅与消息收发
- **Option 配置**：灵活设置地址、密码等参数

### 常见用例

- 缓存读写
- 分布式锁
- 计数器/排行榜
- 消息队列/事件通知
- 脚本原子操作

### 最佳实践

- 生产环境建议配置密码与连接池参数
- 管道/事务适合批量高并发场景
- 脚本操作建议预加载并用 SHA 调用
- 发布订阅需注意消息可靠性
- 始终检查命令返回的 error

## API 文档

### 主要类型

```go
// Redis 客户端接口
 type Redis interface {
    Do(ctx context.Context, args ...interface{}) *Cmd
    Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)
    TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)
    Subscribe(ctx context.Context, channels ...string) *PubSub
    PSubscribe(ctx context.Context, channels ...string) *PubSub
}

// Redis 扩展接口
 type RedisExtension interface {
    Redis
    Get(ctx context.Context, key string) *Cmd
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *Cmd
    Del(ctx context.Context, key string) *Cmd
    Expire(ctx context.Context, key string, expiration time.Duration) *Cmd
}

// Option 配置项类型
 type Option func(*redisClient)

// NewRedis 创建客户端
func NewRedis(opts ...Option) Redis

// NewRedisExtension 创建扩展
func NewRedisExtension(redis Redis) RedisExtension
```

### 关键函数

- `NewRedis`：创建 Redis 客户端，支持 Option 配置
- `Do`：执行任意命令
- `Pipelined/TxPipelined`：管道/事务批量操作
- `Subscribe/PSubscribe`：发布订阅
- `Eval/EvalSha/ScriptLoad/ScriptExists`：Lua 脚本
- `Get/Set/Del/Expire`：常用 KV 操作

### 配置选项

- `WithAddr(addr string)`：设置 Redis 地址
- `WithPassword(password string)`：设置密码

## 错误处理

- 所有命令均返回 *Cmd，需调用 Result() 获取结果与错误
- 不存在 key 时返回 redis.ErrNil
- 连接失败、参数错误等均有详细错误
- Option 多次叠加后者生效

## 性能指标

- 单连接 QPS 万级，管道/批量操作可进一步提升
- 脚本/事务操作性能取决于 Redis 服务端

## 测试覆盖率

- 单元测试覆盖所有接口、边界、异常场景
- 使用 testify，覆盖率 100%

## 调试指南

- 检查 Redis 服务是否可用（默认 127.0.0.1:6379）
- Option 配置可多次叠加，后者生效
- 脚本调试建议先用 Eval 验证

## 相关文档

- [go-redis 官方文档](https://pkg.go.dev/github.com/redis/go-redis/v9)
- [Redis 官方文档](https://redis.io/)

## 贡献指南

欢迎提交 Issue、PR 或建议，详见 [贡献指南](../../CONTRIBUTING.md)。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE)。 