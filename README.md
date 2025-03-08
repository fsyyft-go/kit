[![GitHub language count](https://img.shields.io/github/languages/count/fsyyft-go/kit)](https://github.com/fsyyft-go/kit)
[![GitHub top language](https://img.shields.io/github/languages/top/fsyyft-go/kit)](https://github.com/fsyyft-go/kit)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/blob/main/go.mod)
[![Go Doc](https://pkg.go.dev/badge/github.com/fsyyft-go/kit)](https://pkg.go.dev/github.com/fsyyft-go/kit)
[![Go Report Card](https://goreportcard.com/badge/github.com/fsyyft-go/kit)](https://goreportcard.com/report/github.com/fsyyft-go/kit)
[![GitHub stars](https://img.shields.io/github/stars/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/network)
[![GitHub issues](https://img.shields.io/github/issues/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/pulls)
[![GitHub contributors](https://img.shields.io/github/contributors/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/graphs/contributors)
[![GitHub license](https://img.shields.io/github/license/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/blob/main/LICENSE)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/releases)
[![GitHub last commit](https://img.shields.io/github/last-commit/fsyyft-go/kit)](https://github.com/fsyyft-go/kit/commits/main)
[![GitHub repo size](https://img.shields.io/github/repo-size/fsyyft-go/kit)](https://github.com/fsyyft-go/kit)
[![GitHub workflow status](https://img.shields.io/github/actions/workflow/status/fsyyft-go/kit/go.yml)](https://github.com/fsyyft-go/kit/actions)
[![Go Mod Updates](https://img.shields.io/github/go-mod/updates-available/fsyyft-go/kit)](https://github.com/fsyyft-go/kit)
[![Sourcegraph](https://sourcegraph.com/github.com/fsyyft-go/kit/-/badge.svg)](https://sourcegraph.com/github.com/fsyyft-go/kit)

# Kit - Go 工具包集合

Kit 是一个功能丰富的 Go 语言工具包集合，旨在提供常用的工具函数和组件，帮助开发者更快速地构建高质量的 Go 应用程序。

## 模块列表

### crypto

#### [crypto/des](crypto/des/README.md)

DES 加密工具：提供 DES-CBC 加密/解密功能，支持 PKCS7 填充和多种输入格式（字节数组、字符串、16 进制字符串）。[详细说明 →](crypto/des/README.md)

### kratos

#### [kratos/config](kratos/config/README.md)

配置解码器：对 Kratos 配置系统的扩展，支持对特定后缀（如 .b64）的配置值进行解码。[详细说明 →](kratos/config/README.md)

#### [kratos/middleware](kratos/middleware/README.md)

中间件集合：提供了验证（validate）和基本认证（basicauth）两个中间件，支持请求验证和 HTTP Basic Authentication。[详细说明 →](kratos/middleware/README.md)

#### [kratos/transport/http](kratos/transport/http/README.md)

HTTP 适配器：提供 Kratos HTTP 服务器到 Gin 引擎的转换功能，支持路由和参数转换。[详细说明 →](kratos/transport/http/README.md)

### [log](log/README.md)

日志抽象接口，提供统一的日志记录标准，支持多种底层实现。[详细说明 →](log/README.md)

### [runtime](runtime/README.md)

运行时管理：提供应用程序运行时组件的生命周期管理。[详细说明 →](runtime/README.md)

#### [runtime/goroutine](runtime/goroutine/README.md)

⚠️ 低级工具：用于获取 goroutine ID，仅用于特殊调试场景。[详细说明 →](runtime/goroutine/README.md)

### [testing](testing/README.md)

测试日志工具：提供带有统一前缀的测试日志输出功能，使测试输出更加清晰易读。[详细说明 →](testing/README.md)

### [time](time/README.md)

基于 [carbon](https://github.com/dromara/carbon) 库的时间处理工具包，提供简单的相对时间获取功能和可配置的时间格式化选项。支持编译时配置时区、格式、语言等参数。[详细说明 →](time/README.md)

更多模块正在开发中，敬请期待...

## 如何贡献

我们欢迎任何形式的贡献，包括但不限于：

- 提交问题和建议
- 改进文档
- 提交代码改进
- 分享使用经验

贡献前请阅读我们的 [贡献指南](CONTRIBUTING.md)。

### 开发流程

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 更新日志

详见 [CHANGELOG.md](CHANGELOG.md)

## 常见问题

常见问题解答请查看 [FAQ.md](FAQ.md)

## 版权声明

Copyright © 2025 fsyyft-go

本项目采用 [MIT 许可证](LICENSE)。详见 [LICENSE](LICENSE) 文件。

## 联系我们

- 提交 Issue: [GitHub Issues](https://github.com/fsyyft-go/kit/issues)
- 邮件联系: [fsyyft@gmail.com](mailto:fsyyft@gmail.com)

## 致谢

感谢所有贡献者对本项目的支持！

## 相关项目

---

如果觉得这个项目对你有帮助，欢迎 star ⭐️


