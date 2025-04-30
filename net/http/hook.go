// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"net/http"
	"sync"
	"time"
)

var (
	// 断言 HookContext 实现了 context.Context 接口。
	_ context.Context = (*HookContext)(nil)
)

type (
	// HookContext 包含 HTTP 操作的上下文信息。
	//
	// 该结构体用于在 HTTP 请求生命周期中传递操作相关的上下文、请求信息、执行结果、错误信息及自定义钩子数据。
	HookContext struct {
		// 原始上下文对象。
		originContext context.Context
		// HTTP 方法，如 GET、POST 等。
		method string
		// 请求的 URL 地址。
		url string
		// 原始 HTTP 请求对象。
		request *http.Request
		// 操作开始时间。
		startTime time.Time
		// 操作结束时间。
		endTime time.Time
		// 原始操作返回的错误。
		originError error
		// 原始操作的结果。
		originResult any
		// 用于存储钩子相关的键值对数据。
		hookMap sync.Map
	}
)

// NewHookContext 创建一个新的 HookContext 实例。
//
// 参数：
//   - ctx：原始上下文对象，用于控制操作的生命周期。
//   - method：HTTP 方法，如 GET、POST 等。
//   - url：请求的 URL 地址。
//   - request：原始 HTTP 请求对象。
//
// 返回值：
//   - *HookContext：返回一个新创建的 HookContext 实例。
func NewHookContext(ctx context.Context, method, url string, request *http.Request) *HookContext {
	return &HookContext{
		originContext: ctx,
		method:        method,
		url:           url,
		request:       request,
		startTime:     time.Now(),
	}
}

// SetResult 设置操作结果和结束时间。
//
// 参数：
//   - result：操作的结果，可以是任意类型。
//   - err：操作过程中产生的错误，如果没有错误则为 nil。
func (h *HookContext) SetResult(result any, err error) {
	h.originResult = result
	h.originError = err
	h.endTime = time.Now()
}

// StartTime 返回操作开始时间。
//
// 返回值：
//   - time.Time：返回操作的开始时间。
func (h *HookContext) StartTime() time.Time {
	return h.startTime
}

// EndTime 返回操作结束时间。
//
// 返回值：
//   - time.Time：返回操作的结束时间。
func (h *HookContext) EndTime() time.Time {
	return h.endTime
}

// Duration 返回操作持续时间。
//
// 返回值：
//   - time.Duration：返回操作的持续时间。
func (h *HookContext) Duration() time.Duration {
	return h.endTime.Sub(h.startTime)
}

// Method 返回 HTTP 方法。
//
// 返回值：
//   - string：返回 HTTP 方法字符串。
func (h *HookContext) Method() string {
	return h.method
}

// Url 返回请求的 URL。
//
// 返回值：
//   - string：返回请求的 URL 字符串。
func (h *HookContext) Url() string {
	return h.url
}

// Request 返回原始 HTTP 请求对象。
//
// 返回值：
//   - *http.Request：返回原始 HTTP 请求指针。
func (h *HookContext) Request() *http.Request {
	return h.request
}

// OriginError 返回原始操作错误。
//
// 返回值：
//   - error：返回操作过程中产生的原始错误。
func (h *HookContext) OriginError() error {
	return h.originError
}

// OriginResult 返回原始操作结果。
//
// 返回值：
//   - any：返回操作的原始结果。
func (h *HookContext) OriginResult() any {
	return h.originResult
}

// GetHookValue 获取 hook 中的值。
//
// 参数：
//   - key：要获取的键名。
//
// 返回值：
//   - interface{}：返回与键关联的值。
//   - bool：如果键存在返回 true，否则返回 false。
func (h *HookContext) GetHookValue(key string) (interface{}, bool) {
	return h.hookMap.Load(key)
}

// SetHookValue 设置 hook 中的值。
//
// 参数：
//   - key：要设置的键名。
//   - value：要存储的值。
func (h *HookContext) SetHookValue(key string, value interface{}) {
	h.hookMap.Store(key, value)
}

// Deadline 实现 context.Context 接口。
//
// 返回值：
//   - deadline：返回上下文的截止时间。
//   - ok：如果设置了截止时间返回 true，否则返回 false。
func (h *HookContext) Deadline() (deadline time.Time, ok bool) {
	return h.originContext.Deadline()
}

// Done 实现 context.Context 接口。
//
// 返回值：
//   - <-chan struct{}：返回一个 channel，当上下文被取消时会被关闭。
func (h *HookContext) Done() <-chan struct{} {
	return h.originContext.Done()
}

// Err 实现 context.Context 接口。
//
// 返回值：
//   - error：如果上下文被取消，返回取消的原因。
func (h *HookContext) Err() error {
	return h.originContext.Err()
}

// Value 实现 context.Context 接口。
//
// 参数：
//   - key：要获取的值的键。
//
// 返回值：
//   - interface{}：返回与键关联的值。
func (h *HookContext) Value(key interface{}) interface{} {
	return h.originContext.Value(key)
}

type (
	// Hook 定义 HTTP 操作的钩子接口。
	//
	// 该接口用于在 HTTP 操作执行前后插入自定义逻辑。
	Hook interface {
		// Before 在操作执行前调用。
		//
		// 参数：
		//   - ctx：钩子上下文，包含操作的相关信息。
		//
		// 返回值：
		//   - error：如果钩子执行出错，返回相应的错误信息。
		Before(ctx *HookContext) error

		// After 在操作执行后调用。
		//
		// 参数：
		//   - ctx：钩子上下文，包含操作的相关信息和结果。
		//
		// 返回值：
		//   - error：如果钩子执行出错，返回相应的错误信息。
		After(ctx *HookContext) error
	}
)

// HookManager 管理多个 Hook 的执行。
//
// 该结构体用于统一管理和调度多个钩子的执行顺序。
type HookManager struct {
	// hooks 存储注册的所有钩子。
	hooks []Hook
}

// NewHookManager 创建一个新的 HookManager 实例。
//
// 返回值：
//   - *HookManager：返回一个新创建的 HookManager 实例。
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]Hook, 0),
	}
}

// AddHook 添加一个 Hook。
//
// 参数：
//   - hook：要添加的钩子实例。
func (m *HookManager) AddHook(hook Hook) {
	m.hooks = append(m.hooks, hook)
}

// Before 实现 Hook 接口，按顺序执行所有 Hook 的 Before 方法。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息。
//
// 返回值：
//   - error：如果任何钩子执行出错，返回第一个错误信息。
func (m *HookManager) Before(ctx *HookContext) error {
	for _, hook := range m.hooks {
		if err := hook.Before(ctx); err != nil {
			return err
		}
	}
	return nil
}

// After 实现 Hook 接口，按逆序执行所有 Hook 的 After 方法。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息和结果。
//
// 返回值：
//   - error：如果任何钩子执行出错，返回第一个错误信息。
func (m *HookManager) After(ctx *HookContext) error {
	for i := len(m.hooks) - 1; i >= 0; i-- {
		if err := m.hooks[i].After(ctx); err != nil {
			return err
		}
	}
	return nil
}
