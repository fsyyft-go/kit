# bloom

## 简介

bloom 包实现了高效的布隆过滤器（Bloom Filter），用于大规模集合的成员判定，具备极高的空间效率和可配置的误判率。支持分组、可插拔存储（内存/自定义）、多哈希函数、并发安全，适用于缓存预判、黑名单、唯一性校验等场景。

### 主要特性

- 支持标准布隆过滤器接口（添加、判定、分组操作）
- 支持自定义存储后端（内存、可扩展为 Redis 等）
- 支持分组隔离，适合多租户/多集合场景
- 支持误判率、元素数量等参数灵活配置
- 并发安全，适合高并发环境
- 采用 murmur3 哈希算法，分布均匀
- 完善的错误处理与日志能力
- 完整单元测试覆盖

### 设计理念

bloom 包遵循"高效、灵活、易用"的设计理念，核心接口与实现分离，支持多种存储后端。通过 Option 模式灵活配置，支持分组、日志、参数自定义。默认实现为高性能内存存储，便于本地高并发场景。

## 安装

### 前置条件

- Go 版本要求：Go 1.18+
- 依赖要求：
  - github.com/spaolacci/murmur3
  - github.com/spf13/cast
  - github.com/stretchr/testify（仅测试）
  - github.com/golang/mock/gomock（仅测试）

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/container/bloom
```

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "github.com/fsyyft-go/kit/container/bloom"
)

func main() {
    // 创建布隆过滤器，预计 10000 元素，误判率 0.01
    bf, _, err := bloom.NewBloom(
        bloom.WithName("my-bloom"),
        bloom.WithExpectedElements(10000),
        bloom.WithFalsePositiveRate(0.01),
    )
    if err != nil {
        panic(err)
    }
    ctx := context.Background()
    // 添加元素
    _ = bf.Put(ctx, "foo")
    // 判定元素
    exists, _ := bf.Contain(ctx, "foo")
    fmt.Println("foo 是否可能存在：", exists)
}
```

### 分组用法

```go
// 支持分组隔离
_ = bf.GroupPut(ctx, "group1", "bar")
exists, _ := bf.GroupContain(ctx, "group1", "bar")
```

### 自定义存储

```go
// 使用内存存储，指定内存块大小（字节）
store := bloom.NewMemoryStore(1024*1024) // 1MB
bf, _, _ := bloom.NewBloom(
    bloom.WithStore(store),
    bloom.WithExpectedElements(1000),
)
```

## 详细指南

### 核心概念

- **布隆过滤器**：空间效率极高的概率型集合判定结构，适合大数据量场景。
- **误判率**：可配置，越低则空间需求越大。
- **分组**：同一过滤器可隔离多个逻辑集合。
- **存储后端**：可插拔，默认内存实现。
- **多哈希函数**：自动根据参数计算最优数量。

### 常见用例

- 缓存穿透预判
- 黑名单/白名单判定
- 唯一性校验
- 分布式唯一性预过滤（可扩展 Redis 存储）

### 最佳实践

- 合理设置元素数量和误判率，避免空间浪费或误判过高
- 对于分布式场景可自定义 Store 实现
- 并发场景下直接使用，无需额外加锁
- 始终检查返回的 error

## API 文档

### 主要类型

```go
// Bloom 布隆过滤器接口
 type Bloom interface {
    Contain(ctx context.Context, value string) (bool, error)
    Put(ctx context.Context, value string) error
    GroupContain(ctx context.Context, group, value string) (bool, error)
    GroupPut(ctx context.Context, group, value string) error
}

// Store 存储后端接口
 type Store interface {
    Exist(ctx context.Context, key string, hash []uint64) (bool, error)
    Add(ctx context.Context, key string, hash []uint64) error
}

// Option 配置项类型
 type Option func(*bloom)

// NewBloom 创建布隆过滤器
func NewBloom(opts ...Option) (Bloom, func(), error)

// NewMemoryStore 创建内存存储
func NewMemoryStore(size int) Store
```

### 关键函数

- `NewBloom`：创建布隆过滤器，支持 Option 配置
- `Put/Contain`：添加/判定元素
- `GroupPut/GroupContain`：分组添加/判定
- `NewMemoryStore`：创建内存存储

### 配置选项

- `WithName(name string)`：设置过滤器名称
- `WithStore(store Store)`：自定义存储后端
- `WithExpectedElements(n uint64)`：预计元素数量
- `WithFalsePositiveRate(p float64)`：误判率
- `WithLogger(logger log.Logger)`：自定义日志

## 错误处理

- 名称为空、重复、误判率参数非法等会返回详细错误
- 存储层错误会原样返回
- 建议所有操作均检查 error

## 性能指标

- 内存存储百万级元素判定/添加操作均为 O(k)，k 为哈希函数数，单次操作微秒级
- 误判率与空间、哈希函数数量成反比

## 测试覆盖率

- 单元测试覆盖所有接口、边界、异常场景
- 使用 gomock+testify，覆盖率 100%

## 调试指南

- 检查参数设置是否合理（元素量、误判率）
- 内存存储建议根据实际场景调整 size
- 分组键名冲突可通过 WithName 区分

## 相关文档

- [Bloom Filter 维基百科](https://en.wikipedia.org/wiki/Bloom_filter)
- [murmur3 哈希算法](https://github.com/spaolacci/murmur3)

## 贡献指南

欢迎提交 Issue、PR 或建议，详见 [贡献指南](../../CONTRIBUTING.md)。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE)。 