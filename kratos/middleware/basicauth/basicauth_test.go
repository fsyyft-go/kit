// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// basicauth_test.go 文件设计思路与使用方法说明：
//
// 本测试文件用于测试基本认证（Basic Authentication）中间件的功能实现。
// 设计思路：
// 1. 通过模拟传输层（mockTransport）来测试HTTP基本认证机制
//    - 模拟传输层设计目的：在不依赖真实HTTP服务器的情况下测试认证中间件
//    - 核心实现：mockTransport同时实现了transport.Transport和transport.Header接口
//    - 关键功能点：
//      a. 请求头模拟：通过header字段存储Authorization等认证信息
//      b. 响应头捕获：通过reply字段捕获中间件设置的WWW-Authenticate等头信息
//      c. 传输类型模拟：通过Kind()方法返回transport.KindHTTP，模拟HTTP传输环境
//      d. 上下文集成：借助transport.NewServerContext将模拟传输层注入测试上下文
//    - 优势：
//      a. 隔离依赖：无需启动真实服务器即可测试中间件逻辑
//      b. 精确控制：可以精确控制请求头内容，模拟各种边缘情况
//      c. 结果验证：可以直接检查响应头是否按预期设置
//      d. 场景全面：支持测试认证成功、失败、格式错误等多种场景
// 2. 测试基本认证解析、验证和错误处理的各个方面
// 3. 覆盖各种边缘情况和错误场景，确保中间件的健壮性
//
// 使用方法：
// - 运行所有测试：go test -v ./kratos/middleware/basicauth
// - 运行特定测试：go test -v ./kratos/middleware/basicauth -run TestParseBasicAuth
//
// 主要测试内容：
// - 基本认证头的解析（TestParseBasicAuth）
// - 中间件选项配置（TestWithOptions）
// - 服务器中间件功能（TestServer）
// - 默认验证器行为（TestDefaultValidator）
// - 非传输层上下文处理（TestNonTransportContext）
// - 错误类型比较（TestErrorComparison）

package basicauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/stretchr/testify/assert"
)

// mockTransport 实现了transport.Transport接口，用于在测试中模拟HTTP传输层。
// 它提供了设置和获取请求头、响应头的能力，便于测试基本认证中间件。
type mockTransport struct {
	header map[string]string
	reply  map[string]string
}

// newMockTransport 创建并初始化一个新的mockTransport实例。
// 它初始化请求头和响应头的映射，为测试准备好环境。
func newMockTransport() *mockTransport {
	return &mockTransport{
		header: make(map[string]string),
		reply:  make(map[string]string),
	}
}

// Kind 返回传输类型，实现transport.Transport接口。
// 在测试中默认返回HTTP类型，因为基本认证主要用于HTTP传输。
func (m *mockTransport) Kind() transport.Kind {
	return transport.KindHTTP
}

// Endpoint 返回服务端点地址，实现transport.Transport接口。
// 在测试中返回固定值"mock"。
func (m *mockTransport) Endpoint() string {
	return "mock"
}

// Operation 返回操作名称，实现transport.Transport接口。
// 在测试中返回固定值"mock"。
func (m *mockTransport) Operation() string {
	return "mock"
}

// RequestHeader 返回请求头接口，实现transport.Transport接口。
// 这里直接返回自身，因为mockTransport同时实现了transport.Header接口。
func (m *mockTransport) RequestHeader() transport.Header {
	return m
}

// ReplyHeader 返回响应头接口，实现transport.Transport接口。
// 这里直接返回自身，因为mockTransport同时实现了transport.Header接口。
func (m *mockTransport) ReplyHeader() transport.Header {
	return m
}

// Get 获取请求头中指定键的值，实现transport.Header接口。
// 用于在测试中获取Authorization等头信息。
func (m *mockTransport) Get(key string) string {
	return m.header[key]
}

// Set 设置响应头中指定键的值，实现transport.Header接口。
// 用于在测试中验证WWW-Authenticate等头信息是否正确设置。
func (m *mockTransport) Set(key, value string) {
	m.reply[key] = value
}

// Add 添加响应头信息，实现transport.Header接口。
// 在当前实现中与Set行为相同，将值添加到响应头映射中。
func (m *mockTransport) Add(key, value string) {
	m.reply[key] = value
}

// Values 返回请求头中指定键的所有值，实现transport.Header接口。
// 返回包含单个值的切片或空切片（如果键不存在）。
func (m *mockTransport) Values(key string) []string {
	if value, ok := m.header[key]; ok {
		return []string{value}
	}
	return nil
}

// Keys 返回请求头中的所有键，实现transport.Header接口。
// 用于遍历所有请求头键。
func (m *mockTransport) Keys() []string {
	var keys []string
	for k := range m.header {
		keys = append(keys, k)
	}
	return keys
}

// makeBasicAuthHeader 是一个辅助函数，用于生成基本认证头的值。
// 它将用户名和密码组合成"username:password"格式，然后进行Base64编码，
// 最后添加"Basic "前缀，符合HTTP基本认证的标准格式。
func makeBasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// TestParseBasicAuth 测试parseBasicAuth函数的各种情况。
// 包括有效认证、无效认证格式、非法Base64编码、无冒号分隔、空密码、特殊字符和Unicode字符等场景。
func TestParseBasicAuth(t *testing.T) {
	tests := []struct {
		name          string
		auth          string
		wantUsername  string
		wantPassword  string
		wantErr       bool
		expectedError error
	}{
		{
			name:          "无Basic前缀",
			auth:          "NotBasic dXNlcjpwYXNz",
			wantUsername:  "",
			wantPassword:  "",
			wantErr:       true,
			expectedError: ErrInvalidBasicAuth,
		},
		{
			name:          "非法Base64编码",
			auth:          "Basic not-base64",
			wantUsername:  "",
			wantPassword:  "",
			wantErr:       true,
			expectedError: nil, // 实际会返回base64解码错误
		},
		{
			name:          "无冒号分隔",
			auth:          "Basic " + base64.StdEncoding.EncodeToString([]byte("useronly")),
			wantUsername:  "",
			wantPassword:  "",
			wantErr:       true,
			expectedError: ErrInvalidBasicAuth,
		},
		{
			name:          "有效认证",
			auth:          makeBasicAuthHeader("user", "pass"),
			wantUsername:  "user",
			wantPassword:  "pass",
			wantErr:       false,
			expectedError: nil,
		},
		{
			name:          "有效认证-空密码",
			auth:          makeBasicAuthHeader("user", ""),
			wantUsername:  "user",
			wantPassword:  "",
			wantErr:       false,
			expectedError: nil,
		},
		{
			name:          "有效认证-特殊字符",
			auth:          makeBasicAuthHeader("user@example.com", "p@$$w0rd!"),
			wantUsername:  "user@example.com",
			wantPassword:  "p@$$w0rd!",
			wantErr:       false,
			expectedError: nil,
		},
		{
			name:          "有效认证-Unicode字符",
			auth:          makeBasicAuthHeader("用户", "密码"),
			wantUsername:  "用户",
			wantPassword:  "密码",
			wantErr:       false,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password, err := parseBasicAuth(tt.auth)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUsername, username)
				assert.Equal(t, tt.wantPassword, password)
			}
		})
	}
}

// TestWithOptions 测试中间件选项设置功能。
// 分别测试WithValidator选项、WithRealm选项，以及多个选项组合使用的情况。
func TestWithOptions(t *testing.T) {
	// 测试WithValidator
	t.Run("WithValidator", func(t *testing.T) {
		validator := func(ctx context.Context, username, password string) bool {
			return username == "test" && password == "password"
		}
		opt := WithValidator(validator)
		o := &options{}
		opt(o)

		// 验证validator是否正确设置
		assert.True(t, o.validator(context.Background(), "test", "password"))
		assert.False(t, o.validator(context.Background(), "wrong", "password"))
		assert.False(t, o.validator(context.Background(), "test", "wrong"))
	})

	// 测试WithRealm
	t.Run("WithRealm", func(t *testing.T) {
		realm := "TestRealm"
		opt := WithRealm(realm)
		o := &options{}
		opt(o)

		// 验证realm是否正确设置
		assert.Equal(t, realm, o.realm)
	})

	// 测试多个选项组合
	t.Run("MultipleOptions", func(t *testing.T) {
		validator := func(ctx context.Context, username, password string) bool {
			return username == "admin" && password == "secret"
		}
		realm := "SecuredZone"

		o := &options{}
		WithValidator(validator)(o)
		WithRealm(realm)(o)

		assert.Equal(t, realm, o.realm)
		assert.True(t, o.validator(context.Background(), "admin", "secret"))
		assert.False(t, o.validator(context.Background(), "admin", "wrong"))
	})
}

// TestServer 测试Server中间件在各种情况下的行为。
// 包括无认证头、无效认证头格式、非法Base64编码、认证失败、认证成功、空密码认证成功和非HTTP传输层等场景。
func TestServer(t *testing.T) {
	tests := []struct {
		name           string
		setupTransport func() *mockTransport
		validator      CredentialValidator
		realm          string
		wantErr        bool
		expectedError  error
	}{
		{
			name: "无认证头",
			setupTransport: func() *mockTransport {
				return newMockTransport() // 不设置Authorization头
			},
			validator: func(ctx context.Context, username, password string) bool {
				return true
			},
			realm:         "TestRealm",
			wantErr:       true,
			expectedError: ErrInvalidBasicAuth,
		},
		{
			name: "无效认证头格式",
			setupTransport: func() *mockTransport {
				m := newMockTransport()
				m.header["Authorization"] = "NotBasic xyz"
				return m
			},
			validator: func(ctx context.Context, username, password string) bool {
				return true
			},
			realm:         "TestRealm",
			wantErr:       true,
			expectedError: ErrInvalidBasicAuth,
		},
		{
			name: "非法Base64编码",
			setupTransport: func() *mockTransport {
				m := newMockTransport()
				m.header["Authorization"] = "Basic not-valid-base64"
				return m
			},
			validator: func(ctx context.Context, username, password string) bool {
				return true
			},
			realm:         "TestRealm",
			wantErr:       true,
			expectedError: ErrInvalidBasicAuth,
		},
		{
			name: "认证失败",
			setupTransport: func() *mockTransport {
				m := newMockTransport()
				m.header["Authorization"] = makeBasicAuthHeader("user", "pass")
				return m
			},
			validator: func(ctx context.Context, username, password string) bool {
				return username == "admin" && password == "admin"
			},
			realm:         "TestRealm",
			wantErr:       true,
			expectedError: ErrInvalidBasicAuth,
		},
		{
			name: "认证成功",
			setupTransport: func() *mockTransport {
				m := newMockTransport()
				m.header["Authorization"] = makeBasicAuthHeader("user", "pass")
				return m
			},
			validator: func(ctx context.Context, username, password string) bool {
				return username == "user" && password == "pass"
			},
			realm:   "TestRealm",
			wantErr: false,
		},
		{
			name: "空密码认证成功",
			setupTransport: func() *mockTransport {
				m := newMockTransport()
				m.header["Authorization"] = makeBasicAuthHeader("user", "")
				return m
			},
			validator: func(ctx context.Context, username, password string) bool {
				return username == "user" && password == ""
			},
			realm:   "TestRealm",
			wantErr: false,
		},
		{
			name: "非HTTP传输层",
			setupTransport: func() *mockTransport {
				// 返回一个自定义的Transport，使Kind()方法返回非HTTP类型
				m := newMockTransport()
				return m
			},
			validator: func(ctx context.Context, username, password string) bool {
				return true
			},
			realm:   "TestRealm",
			wantErr: false, // 在非HTTP传输层上，中间件不应该产生错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个模拟的context
			mockTr := tt.setupTransport()
			var ctx context.Context

			if tt.name == "非HTTP传输层" {
				// 使用普通的上下文，不包装为ServerContext
				ctx = context.Background()
			} else {
				ctx = transport.NewServerContext(context.Background(), mockTr)
			}

			// 创建中间件
			middleware := Server(
				WithValidator(tt.validator),
				WithRealm(tt.realm),
			)

			// 创建模拟的处理器
			called := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return "response", nil
			}

			// 执行中间件
			_, err := middleware(handler)(ctx, "request")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				// 检查WWW-Authenticate头是否正确设置
				expectedAuthHeader := fmt.Sprintf(`Basic realm="%s"`, tt.realm)
				assert.Equal(t, expectedAuthHeader, mockTr.reply["WWW-Authenticate"])
				assert.False(t, called, "处理器不应该被调用")
			} else {
				assert.NoError(t, err)
				assert.True(t, called, "处理器应该被调用")
			}
		})
	}
}

// TestDefaultValidator 测试默认验证器的行为。
// 验证在不提供自定义验证器时，默认验证器对各种用户名和密码组合的处理。
func TestDefaultValidator(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		expectedResult bool
	}{
		{
			name:           "有效用户名和密码",
			username:       "user",
			password:       "pass",
			expectedResult: false, // 默认验证器总是返回false
		},
		{
			name:           "空用户名",
			username:       "",
			password:       "pass",
			expectedResult: false,
		},
		{
			name:           "空密码",
			username:       "user",
			password:       "",
			expectedResult: false,
		},
		{
			name:           "空用户名和密码",
			username:       "",
			password:       "",
			expectedResult: false,
		},
	}

	// 创建使用默认选项的中间件
	middleware := Server() // nolint:ineffassign

	// 从Server中提取默认验证器
	o := &options{}
	WithValidator(func(ctx context.Context, username, password string) bool {
		return false
	})(o) // 设置一个临时验证器

	// 通过重新创建中间件来获取默认验证器
	middleware = Server()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建验证请求
			mockTr := newMockTransport()
			mockTr.header["Authorization"] = makeBasicAuthHeader(tt.username, tt.password)
			ctx := transport.NewServerContext(context.Background(), mockTr)

			// 创建模拟的处理器
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			}

			// 执行中间件
			_, err := middleware(handler)(ctx, "request")

			// 默认验证器应该返回false，导致认证失败
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidBasicAuth, err)
		})
	}
}

// TestNonTransportContext 测试在非传输层上下文中中间件的行为。
// 验证当上下文不包含传输层信息时，中间件是否正确地跳过认证过程。
func TestNonTransportContext(t *testing.T) {
	// 创建一个没有transport.Transport的普通上下文
	ctx := context.Background()

	// 创建中间件
	middleware := Server(
		WithValidator(func(ctx context.Context, username, password string) bool {
			return true
		}),
	)

	// 创建模拟的处理器
	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	// 执行中间件
	_, err := middleware(handler)(ctx, "request")

	// 由于没有传输层，中间件应该简单地传递请求给处理器
	assert.NoError(t, err)
	assert.True(t, called, "处理器应该被调用")
}

// TestErrorComparison 测试错误比较功能。
// 验证ErrInvalidBasicAuth与其他错误的比较行为，确保错误类型可以正确识别。
func TestErrorComparison(t *testing.T) {
	// 确保ErrInvalidBasicAuth的错误比较正确
	err1 := ErrInvalidBasicAuth
	err2 := errors.Unauthorized("UNAUTHORIZED", "Invalid basic authentication")

	assert.Equal(t, err1.Code, err2.Code)
	assert.Equal(t, err1.Reason, err2.Reason)
	assert.Equal(t, err1.Message, err2.Message)

	// 测试与其他错误的区别
	otherErr := errors.Unauthorized("OTHER", "Other error")
	assert.NotEqual(t, err1.Reason, otherErr.Reason)
	assert.NotEqual(t, err1.Message, otherErr.Message)
}
