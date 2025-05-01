 # Kratos 配置扩展示例

本示例展示了如何使用 Kit 的 Kratos 配置扩展功能，实现对特殊格式配置值（如 base64 编码和 DES 加密）的自动解码处理。

## 功能特性

- 支持对带特定后缀的配置值进行自动解码
- 支持 base64 解码（使用 .b64 后缀）
- 支持 DES 解密（使用 .des 后缀）
- 支持环境变量取值（使用 .env 后缀）
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
env:
  lang.env: "LANG"
```

### 3. 代码示例

#### 基本配置加载

```go
package main

import (
    "fmt"
    
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/file"
    
    kitkratosconfig "github.com/fsyyft-go/kit/kratos/config"
)

// Config 结构体定义了配置文件的结构。
type Config struct {
	// App 嵌套结构体包含应用程序相关配置。
	App struct {
		// Name 是应用程序名称。
		Name string `json:"name"`
		// Password 是应用程序密码，可能是从 DES 解密后解码得到的。
		Password string `json:"password"`
		// Addr 是应用程序地址，可能是从 base64 编码后解码得到的。
		Addr string `json:"addr"`
	} `json:"app"`
	// Env 是环境变量配置。
	Env struct {
		// Lang 是系统语言。
		Lang string `json:"lang"`
	} `json:"env"`
}

func main() {
    // 创建配置管理器实例。
	c := config.New(
		// 设置配置源为文件源，指定配置文件路径。
		config.WithSource(
			file.NewSource(configPath),
		),
		// 设置自定义解码器，支持特殊格式处理（如 base64 解码）。
		config.WithDecoder(kitkratosconfig.NewDecoder().Decode),
	)
	// 加载配置，如果出错则触发 panic。
	if err := c.Load(); err != nil {
		panic(err)
	}

	// 声明配置结构体变量。
	var cfg Config
	// 将加载的配置扫描到结构体中，如果出错则触发 panic。
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}

	// 打印配置信息。
	fmt.Printf("App Name: %+v\n", cfg.App.Name)
	fmt.Printf("App Password: %+v\n", cfg.App.Password)
	fmt.Printf("App Addr: %+v\n", cfg.App.Addr)
	fmt.Printf("Env Lang: %+v\n", cfg.Env.Lang)
}
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
pwd /development/github.com/fsyyft-go/kit
App Name: example-kratos-config
App Password: 中文配置示例
App Addr: 我是 base64 过的 test
Env Lang: zh_CN.UTF-8
```

## 注意事项

- 配置键名需要使用特定后缀（.b64、.des、.env）才能触发自动解码
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