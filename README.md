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

### [algorithms](algorithms/)

#### [algorithms/snowflake](algorithms/snowflake/)

分布式唯一 ID 生成器：实现经典 Snowflake 算法，支持多节点高并发、趋势递增 64 位 ID、多种编码格式，适用于数据库主键、分布式事务、消息队列等场景。[详细说明 →](algorithms/snowflake/README.md)

### [bytes](bytes/)

字节操作工具：提供安全的随机字节生成功能，基于加密安全的随机数生成器，适用于生成nonce、salt、会话令牌等安全场景。[详细说明 →](bytes/README.md)

### [cache](cache/)

高性能进程内缓存：基于 ristretto 的缓存实现，支持过期时间设置、泛型接口和自动内存管理。[详细说明 →](cache/README.md)

### [convert](convert/)

通用类型转换工具：支持任意类型与基础类型、切片、Map、结构体之间的安全转换，兼容 gconv，提供带错误和无错误两套 API，适用于数据解析、配置加载、接口适配等场景。[详细说明 →](convert/README.md)

### [container](container/)

#### [container/bloom](container/bloom/)

高效布隆过滤器：支持分组、可插拔存储、误判率灵活配置，适合缓存预判、唯一性校验等大规模集合判定场景。[详细说明 →](container/bloom/README.md)

### [crypto](crypto/)

#### [crypto/aes](crypto/aes/)

AES 加密工具：提供 AES-GCM 加密/解密功能，支持多种输入格式（字节数组、字符串、Base64、Hex）和自动随机 nonce 生成。[详细说明 →](crypto/aes/README.md)

#### [crypto/des](crypto/des/)

DES 加密工具：提供 DES-CBC 加密/解密功能，支持 PKCS7 填充和多种输入格式（字节数组、字符串、16 进制字符串）。[详细说明 →](crypto/des/README.md)

#### [crypto/md5](crypto/md5/)

MD5 哈希工具：提供便捷的字符串 MD5 哈希计算功能，支持带错误处理和忽略错误的版本，适用于数据校验和缓存键生成。[详细说明 →](crypto/md5/README.md)

#### [crypto/otp](crypto/otp/)

一次性密码工具：提供基于时间的一次性密码（TOTP）算法实现，支持多种哈希算法、自定义密码长度和生成兼容的验证器 URL。[详细说明 →](crypto/otp/README.md)

#### [crypto/rsa](crypto/rsa/)

RSA 加密工具：提供 RSA 加密/解密功能，支持公钥加密/私钥解密和私钥加密/公钥解密（数字签名）操作，以及 PEM 格式密钥处理。[详细说明 →](crypto/rsa/README.md)

#### [crypto/sha](crypto/sha/)

SHA256 哈希工具：提供便捷的字符串 SHA256 哈希计算功能，支持带错误处理和忽略错误的版本，适用于数据完整性校验、签名、区块链等安全场景。[详细说明 →](crypto/sha/README.md)

### [database](database/)

#### [database/redis](database/redis/)

高性能 Redis 客户端：支持原生命令、管道、事务、Lua 脚本、发布订阅、基础 KV 操作等，兼容 go-redis v9。[详细说明 →](database/redis/README.md)

#### [database/sql](database/sql/)

##### [database/sql/driver](database/sql/driver/)

数据库驱动接口：提供标准的数据库驱动接口定义，支持自定义驱动实现和连接管理。[详细说明 →](database/sql/driver/README.md)

##### [database/sql/mysql](database/sql/mysql/)

MySQL 数据库工具：提供 MySQL 数据库连接池管理、查询构建器和事务处理等功能，支持读写分离和连接池配置。[详细说明 →](database/sql/mysql/README.md)

### [kratos](kratos/)

#### [kratos/config](kratos/config/)

配置解码器：对 Kratos 配置系统的扩展，支持对特定后缀（如 .b64）的配置值进行解码。[详细说明 →](kratos/config/README.md)

#### [kratos/middleware](kratos/middleware/)

中间件集合：提供了验证（validate）和基本认证（basicauth）两个中间件，支持请求验证和 HTTP Basic Authentication。[详细说明 →](kratos/middleware/README.md)

#### [kratos/transport/http](kratos/transport/http/)

HTTP 适配器：提供 Kratos HTTP 服务器到 Gin 引擎的转换功能，支持路由和参数转换。[详细说明 →](kratos/transport/http/README.md)

### [log](log/)

日志抽象接口，提供统一的日志记录标准，支持多种底层实现。[详细说明 →](log/README.md)

### [math](math/)

#### [math/rand](math/rand/)

随机数生成工具：提供范围内的随机数生成和中文字符（汉字、姓氏）随机生成功能，支持自定义随机数生成器。[详细说明 →](math/rand/README.md)

### [net](net/)

#### [net/http](net/http/)

功能丰富的 HTTP 客户端：支持 GET/POST/HEAD/表单/JSON、超时、代理、钩子、慢请求日志、trace、全局方法等。[详细说明 →](net/http/README.md)

#### [net/message](net/message/)

高性能自定义消息协议与连接封装：支持消息类型注册、心跳包、字符串消息、自动分包、并发安全等，适用于分布式服务、长连接、定制协议等场景。[详细说明 →](net/message/README.md)

### [runtime](runtime/)

运行时管理：提供应用程序运行时组件的生命周期管理。[详细说明 →](runtime/README.md)

#### [runtime/goroutine](runtime/goroutine/)

goroutine 管理工具：提供 goroutine ID 获取和高效的协程池实现。支持任务调度、资源管理和性能监控等功能，适用于并发任务处理和性能优化场景。[详细说明 →](runtime/goroutine/README.md)

#### [runtime/retry](runtime/retry/)

重试机制工具：提供通用的重试机制，支持带上下文和指数退避的函数重试，适用于网络请求、数据库操作等易失败场景。[详细说明 →](runtime/retry/README.md)

### [testing](testing/)

测试日志工具：提供带有统一前缀的测试日志输出功能，使测试输出更加清晰易读。[详细说明 →](testing/README.md)

### [time](time/)

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


