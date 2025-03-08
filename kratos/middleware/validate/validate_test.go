// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// validate_test.go 提供了对验证中间件的测试用例。
//
// 测试设计思路：
// 1. 使用表格驱动测试确保测试用例的可维护性和可扩展性；
// 2. 测试所有可能的场景：请求不需要验证、验证成功、验证失败等；
// 3. 测试默认回调和自定义回调的处理；
// 4. 使用模拟对象和接口实现进行隔离测试；
// 5. 确保测试的代码覆盖率尽可能高。
//
// 使用方法：
// 可以使用以下命令运行测试：
// go test -v github.com/fsyyft-go/kratos/middleware/validate
// 或者使用覆盖率检查：
// go test -cover github.com/fsyyft-go/kratos/middleware/validate
package validate

import (
	"context"
	"errors"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/stretchr/testify/assert"
)

// 测试用的模拟请求结构体，实现了validator接口
type mockValidRequest struct {
	ShouldFail bool
}

// Validate 实现validator接口的验证方法
func (m *mockValidRequest) Validate() error {
	if m.ShouldFail {
		return errors.New("验证失败")
	}
	return nil
}

// 测试用的模拟请求结构体，未实现validator接口
type mockNonValidRequest struct {
	Data string
}

// TestValidatorMiddleware 测试验证器中间件在各种情况下的行为
func TestValidatorMiddleware(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name           string      // 测试用例名称
		request        interface{} // 输入的请求对象
		handlerCalled  bool        // 期望下一个处理器是否被调用
		expectedErr    bool        // 是否期望返回错误
		expectedErrMsg string      // 期望的错误消息，如果为空则不检查
		options        []Option    // 中间件选项
		expectedReply  interface{} // 期望的响应，如果为nil则不检查
	}{
		{
			name:           "非验证器请求",
			request:        &mockNonValidRequest{Data: "测试数据"},
			handlerCalled:  true,
			expectedErr:    false,
			expectedErrMsg: "",
			options:        nil,
			expectedReply:  "处理成功",
		},
		{
			name:           "验证通过",
			request:        &mockValidRequest{ShouldFail: false},
			handlerCalled:  true,
			expectedErr:    false,
			expectedErrMsg: "",
			options:        nil,
			expectedReply:  "处理成功",
		},
		{
			name:           "验证失败-默认回调",
			request:        &mockValidRequest{ShouldFail: true},
			handlerCalled:  false,
			expectedErr:    true,
			expectedErrMsg: "VALIDATOR", // 只检查错误代码部分
			options:        nil,
			expectedReply:  nil,
		},
		{
			name:           "验证失败-自定义回调",
			request:        &mockValidRequest{ShouldFail: true},
			handlerCalled:  false,
			expectedErr:    true,
			expectedErrMsg: "CUSTOM_ERROR", // 只检查错误代码部分
			options: []Option{
				WithValidateCallback(func(ctx context.Context, req interface{}, err error) (interface{}, error) {
					return nil, kerrors.BadRequest("CUSTOM_ERROR", "自定义错误").WithCause(err)
				}),
			},
			expectedReply: nil,
		},
		{
			name:           "验证失败-自定义回调返回响应",
			request:        &mockValidRequest{ShouldFail: true},
			handlerCalled:  false,
			expectedErr:    false,
			expectedErrMsg: "",
			options: []Option{
				WithValidateCallback(func(ctx context.Context, req interface{}, err error) (interface{}, error) {
					return "自定义响应", nil
				}),
			},
			expectedReply: "自定义响应",
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个记录调用状态的测试处理器
			var handlerCalled bool
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "处理成功", nil
			}

			// 创建中间件并应用到处理器
			middleware := Validator(tt.options...)
			wrappedHandler := middleware(handler)

			// 执行处理器
			reply, err := wrappedHandler(context.Background(), tt.request)

			// 验证处理器调用状态
			assert.Equal(t, tt.handlerCalled, handlerCalled, "处理器调用状态不符合预期")

			// 验证错误结果
			if tt.expectedErr {
				assert.Error(t, err, "应该返回错误")
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg, "错误消息不符合预期")
				}
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}

			// 验证响应内容
			if tt.expectedReply != nil {
				assert.Equal(t, tt.expectedReply, reply, "响应内容不符合预期")
			}
		})
	}
}

// TestWithValidateCallback 测试验证回调选项函数
func TestWithValidateCallback(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name            string           // 测试用例名称
		callback        ValidateCallback // 要设置的回调函数
		expectedOutcome func(*Options)   // 预期的选项对象状态检查函数
	}{
		{
			name: "设置自定义回调",
			callback: func(ctx context.Context, req interface{}, err error) (interface{}, error) {
				return "自定义响应", nil
			},
			expectedOutcome: func(o *Options) {
				// 使用一个简单调用来检查回调是否被正确设置
				reply, err := o.callback(context.Background(), nil, nil)
				assert.Equal(t, "自定义响应", reply, "回调响应不符合预期")
				assert.Nil(t, err, "回调错误不符合预期")
			},
		},
		{
			name: "设置错误回调",
			callback: func(ctx context.Context, req interface{}, err error) (interface{}, error) {
				return nil, errors.New("自定义错误")
			},
			expectedOutcome: func(o *Options) {
				// 检查回调是否正确返回错误
				reply, err := o.callback(context.Background(), nil, nil)
				assert.Nil(t, reply, "回调响应不符合预期")
				assert.Error(t, err, "回调错误不符合预期")
				assert.Equal(t, "自定义错误", err.Error(), "错误消息不符合预期")
			},
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建选项对象
			options := &Options{}

			// 应用选项函数
			opt := WithValidateCallback(tt.callback)
			opt(options)

			// 检查选项对象状态
			tt.expectedOutcome(options)
		})
	}
}

// TestValidatorIntegration 集成测试，测试验证器中间件在中间件链中的工作方式
func TestValidatorIntegration(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name          string                  // 测试用例名称
		request       interface{}             // 输入的请求对象
		middlewares   []middleware.Middleware // 中间件链
		expectedErr   bool                    // 是否期望返回错误
		expectedReply interface{}             // 期望的响应
	}{
		{
			name:    "验证通过-单个中间件",
			request: &mockValidRequest{ShouldFail: false},
			middlewares: []middleware.Middleware{
				Validator(),
			},
			expectedErr:   false,
			expectedReply: "处理成功",
		},
		{
			name:    "验证失败-单个中间件",
			request: &mockValidRequest{ShouldFail: true},
			middlewares: []middleware.Middleware{
				Validator(),
			},
			expectedErr:   true,
			expectedReply: nil,
		},
		{
			name:    "验证通过-多个中间件",
			request: &mockValidRequest{ShouldFail: false},
			middlewares: []middleware.Middleware{
				func(handler middleware.Handler) middleware.Handler {
					return func(ctx context.Context, req interface{}) (interface{}, error) {
						// 第一个中间件简单传递请求
						return handler(ctx, req)
					}
				},
				Validator(),
				func(handler middleware.Handler) middleware.Handler {
					return func(ctx context.Context, req interface{}) (interface{}, error) {
						// 第三个中间件简单传递请求
						return handler(ctx, req)
					}
				},
			},
			expectedErr:   false,
			expectedReply: "处理成功",
		},
		{
			name:    "验证失败-多个中间件",
			request: &mockValidRequest{ShouldFail: true},
			middlewares: []middleware.Middleware{
				func(handler middleware.Handler) middleware.Handler {
					return func(ctx context.Context, req interface{}) (interface{}, error) {
						// 第一个中间件简单传递请求
						return handler(ctx, req)
					}
				},
				Validator(),
				func(handler middleware.Handler) middleware.Handler {
					return func(ctx context.Context, req interface{}) (interface{}, error) {
						// 这个中间件不应该被执行
						t.Error("不应该执行到这个中间件")
						return handler(ctx, req)
					}
				},
			},
			expectedErr:   true,
			expectedReply: nil,
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建最终处理器
			finalHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "处理成功", nil
			}

			// 构建中间件链
			handler := finalHandler
			for i := len(tt.middlewares) - 1; i >= 0; i-- {
				handler = tt.middlewares[i](handler)
			}

			// 执行处理器链
			reply, err := handler(context.Background(), tt.request)

			// 验证结果
			if tt.expectedErr {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
			assert.Equal(t, tt.expectedReply, reply, "响应内容不符合预期")
		})
	}
}

// TestValidatorEdgeCases 测试验证器中间件的边缘情况
func TestValidatorEdgeCases(t *testing.T) {
	// 定义一个无效的Option函数（为nil）
	var nilOption Option

	// 测试空Option情况
	t.Run("空Option", func(t *testing.T) {
		middleware := Validator(nilOption)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "处理成功", nil
		}
		wrappedHandler := middleware(handler)

		// 应该正常工作，不会panic
		reply, err := wrappedHandler(context.Background(), &mockNonValidRequest{})
		assert.NoError(t, err, "不应该返回错误")
		assert.Equal(t, "处理成功", reply, "响应内容不符合预期")
	})

	// 测试使用nil请求
	t.Run("nil请求", func(t *testing.T) {
		middleware := Validator()
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "处理成功", nil
		}
		wrappedHandler := middleware(handler)

		// nil请求应该被安全处理（不是validator，直接通过）
		reply, err := wrappedHandler(context.Background(), nil)
		assert.NoError(t, err, "不应该返回错误")
		assert.Equal(t, "处理成功", reply, "响应内容不符合预期")
	})

	// 测试设置回调函数后的行为
	t.Run("设置回调后的行为", func(t *testing.T) {
		// 创建自定义回调
		customCallback := func(ctx context.Context, req interface{}, err error) (interface{}, error) {
			return "自定义回调响应", nil
		}

		// 创建中间件
		middleware := Validator(WithValidateCallback(customCallback))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "不应该到达这里", nil
		}
		wrappedHandler := middleware(handler)

		// 使用会失败的验证器
		req := &mockValidRequest{ShouldFail: true}

		// 执行
		reply, err := wrappedHandler(context.Background(), req)
		assert.NoError(t, err, "使用自定义回调不应该返回错误")
		assert.Equal(t, "自定义回调响应", reply, "响应内容不符合预期")
	})
}
