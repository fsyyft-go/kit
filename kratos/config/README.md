# Config 包

`config` 包是对 Kratos 配置系统的扩展，主要提供自定义解码器功能，支持对特定后缀的配置值进行解码处理（如 base64）。

## 特性

- 支持对带特定后缀的配置值进行解码
- 目前支持 base64 解码（使用 .b64 后缀）
- 与 Kratos 配置系统无缝集成
- 保持原有配置功能

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/file"
    kit_kratos_config "github.com/fsyyft-go/kit/kratos/config"
)

// 定义配置结构
type Config struct {
    App struct {
        Name     string `json:"name"`
        Password string `json:"password"` // 将自动解码 base64 值
    } `json:"app"`
}

func main() {
    // 创建配置实例
    c := config.New(
        config.WithSource(
            file.NewSource("config.yaml"),
        ),
        // 设置自定义解码器
        config.WithDecoder(kit_kratos_config.NewDecoder().Decode),
    )
    
    // 加载配置
    if err := c.Load(); err != nil {
        panic(err)
    }

    // 解析配置到结构体
    var cfg Config
    if err := c.Scan(&cfg); err != nil {
        panic(err)
    }

    // 使用配置
    fmt.Printf("App Name: %s\n", cfg.App.Name)
    fmt.Printf("Password (decoded): %s\n", cfg.App.Password)
}
```

### 配置示例

```yaml
# config.yaml
app:
  name: example-kratos-config
  password.b64: 5oiR5pivIGJhc2U2NCDov4fnmoQgdGVzdA==  # 将被自动解码
```

## 最佳实践

1. 配置命名
   - 使用 .b64 后缀表示需要进行 base64 解码的值
   - 保持命名规范一致
   - 避免在不需要解码的值上使用后缀

2. 与 Kratos 集成
   - 使用 config.WithDecoder 设置自定义解码器
   - 遵循 Kratos 配置最佳实践
   - 合理组织配置结构

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 