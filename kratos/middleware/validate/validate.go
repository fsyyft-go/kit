// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// package validate 提供了请求验证的中间件功能，用于在处理请求前验证请求的合法性。
package validate

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

type (
	// validator 接口定义了需要验证的请求结构体必须实现的方法。
	// 任何实现此接口的结构体都可以被验证器中间件自动验证。
	validator interface {
		// Validate 执行验证逻辑，如果验证失败则返回错误。
		Validate() error
	}

	// ValidateCallback 是一个函数类型，用于处理验证失败时的回调。
	//
	// 参数：
	//   - ctx context.Context：上下文。
	//   - req interface{}：原始请求。
	//   - errValidate error：验证错误。
	//
	// 返回值：
	//   - interface{}：可选的替代响应。
	//   - error：处理后的错误信息。
	ValidateCallback func(ctx context.Context, req interface{}, errValidate error) (interface{}, error)

	// Option 是验证器中间件的配置选项函数。
	Option func(*Options)

	// Options 存储验证器中间件的所有配置。
	Options struct {
		// callback 是验证失败时将被调用的回调函数。
		callback ValidateCallback
	}
)

// WithValidateCallback 设置验证失败时的回调函数。
//
// 参数：
//   - fn ValidateCallback：自定义的验证失败处理函数。
//
// 返回值：
//   - Option：中间件配置选项。
func WithValidateCallback(fn ValidateCallback) Option {
	return func(o *Options) {
		o.callback = fn
	}
}

// Validator 创建一个验证器中间件。
//
// 该中间件会检查请求是否实现了 validator 接口，如果实现了则调用其 Validate 方法。
//
// 参数：
//   - opts ...Option：中间件配置选项。
//
// 返回值：
//   - middleware.Middleware：验证器中间件。
func Validator(opts ...Option) middleware.Middleware {
	// 初始化选项对象，设置默认的回调函数。
	// 默认回调函数将验证错误转换为 BadRequest 类型的错误响应。
	options := &Options{
		callback: func(ctx context.Context, req interface{}, errValidate error) (interface{}, error) {
			return nil, errors.BadRequest("VALIDATOR", errValidate.Error()).WithCause(errValidate)
		},
	}
	// 应用所有提供的选项配置。
	for _, o := range opts {
		if nil == o {
			continue
		}
		o(options)
	}

	// 返回中间件函数，该函数接收下一个处理器作为参数。
	return func(handler middleware.Handler) middleware.Handler {
		// 返回实际的请求处理函数。
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 检查请求是否实现了 validator 接口。
			if v, ok := req.(validator); ok {
				// 调用请求的 Validate 方法进行验证。
				if err := v.Validate(); nil != err {
					// 如果验证失败，则调用配置的回调函数处理错误。
					return options.callback(ctx, req, err)
				}
			}
			// 验证成功或请求不需要验证，继续处理请求。
			return handler(ctx, req)
		}
	}
}
