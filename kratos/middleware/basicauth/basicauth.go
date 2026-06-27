// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package basicauth

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var (
	// ErrInvalidBasicAuth 表示 Basic Authentication 校验失败。
	//
	// 当请求缺少 Authorization 头、头部不是合法的 Basic 凭据，或
	// CredentialValidator 拒绝用户名与密码时，Server 返回该错误。
	// 对于携带服务端 transport 的请求，中间件会在返回前写入
	// WWW-Authenticate 响应头；调用方通常按 401 Unauthorized 处理，
	// 并可通过返回的 Kratos 错误 reason `UNAUTHORIZED` 识别该失败。
	ErrInvalidBasicAuth = errors.Unauthorized("UNAUTHORIZED", "Invalid basic authentication")
)

type (
	// Option 配置 Server 返回的 Basic Authentication 中间件。
	//
	// Option 通常由 WithValidator 或 WithRealm 返回。Server 不会忽略 nil Option，
	// 调用方应只传入有效选项。
	Option func(*options)

	// options 包含中间件配置选项。
	options struct {
		// 认证信息验证器。
		validator CredentialValidator
		// 认证域，显示在浏览器认证对话框中。
		realm string
	}

	// CredentialValidator 校验从 Authorization 头中解析出的用户名和密码。
	//
	// 参数：
	//   - ctx context.Context：当前请求上下文。
	//   - username string：从 Authorization 头中解析出的用户名。
	//   - password string：从 Authorization 头中解析出的密码。
	//
	// 返回值：
	//   - bool：返回 true 表示凭据通过校验；返回 false 表示 Server 应返回 ErrInvalidBasicAuth。
	//
	// Server 仅在从服务端 transport 中成功读取并解析 Basic 凭据后调用该回调。
	// 若未通过 WithValidator 提供自定义实现，默认 validator 始终返回 false。
	CredentialValidator func(ctx context.Context, username, password string) bool
)

// WithValidator 配置自定义的 CredentialValidator。
//
// 参数：
//   - validator CredentialValidator：用于校验 Authorization 头中用户名与密码的回调。
//
// 返回值：
//   - Option：中间件配置选项。
//
// validator 必须为非 nil。若未设置该选项，Server 使用的默认 validator 始终拒绝认证。
func WithValidator(validator CredentialValidator) Option {
	return func(o *options) {
		o.validator = validator
	}
}

// WithRealm 配置认证失败时写入 WWW-Authenticate 头的 realm。
//
// 参数：
//   - realm string：用于构造 `Basic realm="..."` 响应头的认证域名。
//
// 返回值：
//   - Option：中间件配置选项。
//
// 若未设置该选项，Server 使用默认 realm `Restricted`。当请求缺少凭据、凭据格式非法
// 或 CredentialValidator 拒绝认证时，携带服务端 transport 的请求会收到该 realm。
func WithRealm(realm string) Option {
	return func(o *options) {
		o.realm = realm
	}
}

// Server 创建用于服务端请求的 Basic Authentication 中间件。
//
// 参数：
//   - opts ...Option：中间件配置选项。
//
// 返回值：
//   - middleware.Middleware：在进入后续处理器前执行 Basic Authentication 校验的中间件。
//
// 中间件会从 transport.ServerContext 中读取 Authorization 请求头并解析 Basic 凭据。
// 当请求缺少凭据、凭据格式非法或 CredentialValidator 返回 false 时，中间件会设置
// `WWW-Authenticate: Basic realm="..."` 响应头并返回 ErrInvalidBasicAuth。
//
// 若上下文中不存在服务端 transport，中间件不会尝试认证，而是直接调用后续处理器。
// 未显式配置时，默认 validator 始终拒绝认证，默认 realm 为 `Restricted`。
func Server(opts ...Option) middleware.Middleware {
	// 初始化配置选项，设置默认值。
	o := &options{
		validator: func(ctx context.Context, username, password string) bool {
			// 默认认证器，返回 false，需要覆盖此函数。
			return false
		},
		realm: "Restricted",
	}
	// 应用传入的配置选项。
	for _, opt := range opts {
		opt(o)
	}

	// 返回中间件处理函数。
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 从服务端上下文中获取传输层信息。
			if tr, ok := transport.FromServerContext(ctx); ok {
				// 获取请求头中的 Authorization 字段。
				auths := tr.RequestHeader().Get("Authorization")
				if auths == "" {
					// 如果认证头为空，设置 WWW-Authenticate 头，触发浏览器的认证对话框。
					tr.ReplyHeader().Set("WWW-Authenticate", `Basic realm="`+o.realm+`"`)
					return nil, ErrInvalidBasicAuth
				}

				// 解析认证头。
				username, password, err := parseBasicAuth(auths)
				if nil != err {
					// 如果解析失败，设置 WWW-Authenticate 头，触发浏览器的认证对话框。
					tr.ReplyHeader().Set("WWW-Authenticate", `Basic realm="`+o.realm+`"`)
					return nil, ErrInvalidBasicAuth
				}

				// 验证用户名和密码。
				if !o.validator(ctx, username, password) {
					// 如果验证失败，设置 WWW-Authenticate 头，触发浏览器的认证对话框。
					tr.ReplyHeader().Set("WWW-Authenticate", `Basic realm="`+o.realm+`"`)
					return nil, ErrInvalidBasicAuth
				}
			}
			// 验证通过，继续处理请求。
			return handler(ctx, req)
		}
	}
}

// parseBasicAuth 解析 HTTP Basic Auth 头部值，返回用户名和密码。
//
// 参数：
//   - auth string：Authorization 头部值。
//
// 返回值：
//   - username string：用户名。
//   - password string：密码。
//   - error：解析过程中可能发生的错误。
func parseBasicAuth(auth string) (string, string, error) {
	// 检查认证头是否以 "Basic " 开头。
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", ErrInvalidBasicAuth
	}

	// 去除 "Basic " 前缀。
	payload := strings.TrimPrefix(auth, "Basic ")
	// 对 Base64 编码的载荷进行解码。
	decodedBytes, err := base64.StdEncoding.DecodeString(payload)
	if nil != err {
		return "", "", err
	}

	// 将解码后的字节转换为字符串。
	decodedString := string(decodedBytes)
	// 按冒号分割，最多分为两部分（用户名和密码）。
	parts := strings.SplitN(decodedString, ":", 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidBasicAuth
	}

	// 返回用户名和密码。
	return parts[0], parts[1], nil
}
