// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package validate

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
)

type (
	// validator 定义请求对象参与自动校验所需实现的最小接口。
	validator interface {
		// Validate 执行验证逻辑，如果验证失败则返回错误。
		Validate() error
	}

	// ValidateCallback 处理请求对象 Validate 失败后的响应。
	//
	// 参数：
	//   - ctx context.Context：当前请求上下文。
	//   - req interface{}：返回校验错误的原始请求对象。
	//   - errValidate error：请求对象 Validate 返回的原始错误。
	//
	// 返回值：
	//   - interface{}：可选的替代响应；返回非 nil 且 error 为 nil 时，Validator 直接返回该响应。
	//   - error：要返回给调用方的错误；默认实现会把 errValidate 转换为带 cause 的 Kratos BadRequest。
	//
	// 自定义回调可以替换默认错误转换逻辑，也可以在校验失败时返回自定义响应并抑制错误。
	ValidateCallback func(ctx context.Context, req interface{}, errValidate error) (interface{}, error)

	// Option 配置 Validator 返回的请求校验中间件。
	//
	// Option 通常由 WithValidateCallback 返回。Validator 会忽略 nil Option。
	Option func(*Options)

	// Options 存储 Validator 中间件的配置。
	Options struct {
		// callback 是验证失败时将被调用的回调函数。
		callback ValidateCallback
	}
)

// WithValidateCallback 设置校验失败时使用的 ValidateCallback。
//
// 参数：
//   - fn ValidateCallback：自定义的校验失败处理函数。
//
// 返回值：
//   - Option：中间件配置选项。
//
// 若未设置该选项，Validator 会把 Validate 返回的错误转换为 code 为 `VALIDATOR`
// 的 Kratos BadRequest，并使用 WithCause 保留原始校验错误。
func WithValidateCallback(fn ValidateCallback) Option {
	return func(o *Options) {
		o.callback = fn
	}
}

// Validator 创建用于请求对象 Validate 校验的中间件。
//
// 该中间件只处理实现 `Validate() error` 的请求对象；未实现该方法的请求会直接透传给后续处理器。
//
// 参数：
//   - opts ...Option：中间件配置选项。
//
// 返回值：
//   - middleware.Middleware：在调用后续处理器前执行请求校验的中间件。
//
// 当请求对象的 Validate 返回错误时，中间件不会继续执行后续处理器，而是调用
// ValidateCallback 生成返回值。默认回调会返回 code 为 `VALIDATOR` 的 Kratos
// BadRequest，并通过 WithCause 保留原始校验错误。
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
