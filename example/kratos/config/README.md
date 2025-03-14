 # Kratos 配置扩展示例

本示例展示了如何使用 Kit 的 Kratos 配置扩展功能，实现对特殊格式配置值（如 base64 编码和 DES 加密）的自动解码处理。

## 功能特性

- 支持对带特定后缀的配置值进行自动解码
- 支持 base64 解码（使用 .b64 后缀）
- 支持 DES 解密（使用 .des 后缀）
- 与 Kratos 配置系统无缝集成
- 提供命令行工具进行 DES 加密解密操作

## 设计原理

Kit 的 Kratos 配置扩展模块采用了以下设计：

- 通过自定义解码器扩展 Kratos 的配置系统
- 使用后缀标识需要特殊处理的配置项（如 .b64、.des）
- 在解码过程中自动识别并处理特殊格式的配置值
- 提供命令行工具便于配置值的加密和解密

这种设计使得应用程序可以安全地存储敏感配置信息，同时保持配置文件的可读性和易用性。

## 使用方法

### 1. 编译和运行

在 Unix/Linux/macOS 系统上：

```bash
# 添加执行权限
chmod +x build.sh

# 构建和运行
./build.sh
```

### 2. 配置文件示例

```yaml
# config.yaml
app:
  name: example-kratos-config
  password.des: A16121360A42757F6B6307A8AD8C37163647D18BE7921339  # DES 加密的密码
  addr.b63: 5oiR5pivIGJhc2U2NCDov4fnmoQgdGVzdA==  # base64 编码的地址
```

### 3. 代码示例

#### 基本配置加载

```go
// 创建配置管理器实例
c := config.New(
    // 设置配置源为文件源，指定配置文件路径
    config.WithSource(
        file.NewSource(configPath),
    ),
    // 设置自定义解码器，支持特殊格式处理
    config.WithDecoder(kit_kratos_config.NewDecoder().Decode),
)

// 加载配置
if err := c.Load(); err != nil {
    panic(err)
}

// 声明配置结构体
var cfg Config
// 将加载的配置扫描到结构体中
if err := c.Scan(&cfg); err != nil {
    panic(err)
}

// 使用配置（已自动解码）
fmt.Printf("%+v\n", cfg.App.Name)
fmt.Printf("%+v\n", cfg.App.Password)  // 已自动解密
fmt.Printf("%+v\n", cfg.App.Addr)      // 已自动解码
```

#### 使用 DES 加密/解密工具

```bash
# 使用默认密钥加密数据
./config des --data "需要加密的数据" --encrypt

# 使用自定义密钥加密数据
./config des --key "自定义密钥" --data "需要加密的数据" --encrypt

# 使用自定义密钥解密数据
./config des --key "自定义密钥" --data "已加密的数据" --encrypt=false
```

### 4. 输出示例

```
pwd /path/to/example/kratos/config
example-kratos-config
password123
我是 base64 编码的 test
```

### 5. 在其他项目中使用

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
        Password string `json:"password"`  // 对应 password.des
        ApiKey   string `json:"apiKey"`    // 对应 apiKey.b64
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

    // 使用配置（已自动解码）
    fmt.Printf("App Name: %s\n", cfg.App.Name)
    fmt.Printf("Password (decoded): %s\n", cfg.App.Password)
    fmt.Printf("API Key (decoded): %s\n", cfg.App.ApiKey)
}
```

## 注意事项

- 配置键名需要使用特定后缀（.b64、.des）才能触发自动解码
- 解码后的值会存储在去除后缀的键名下
- DES 加密默认使用内置密钥，可以通过命令行工具指定自定义密钥
- 确保 base64 编码和 DES 加密的值格式正确，否则会导致解码错误
- 在生产环境中，建议使用环境变量或其他安全方式管理加密密钥

## 相关文档

- [Kratos 配置文档](https://go-kratos.dev/docs/component/config/)
- [Kit Kratos 配置扩展文档](../../kratos/config/README.md)
- [Kit DES 加密模块文档](../../crypto/des/README.md)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../../LICENSE) 文件。