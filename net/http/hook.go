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
	// 断言 HookContext 实现 context.Context 接口。
	_ context.Context = (*HookContext)(nil)
	// 断言 HookManager 实现 Hook 接口。
	_ Hook = (*HookManager)(nil)
)

type (
	// HookContext 包含一次 HTTP 请求在 Hook 链中共享的上下文信息。
	//
	// HookContext 同时实现 context.Context，并把 Deadline、Done、Err 和 Value 委托给原始上下文。
	// Hook 可以通过 SetHookValue 和 GetHookValue 共享附加数据；请求结果需在请求完成后由 SetResult 写入。
	HookContext struct {
		// 原始上下文对象，context.Context 接口方法会委托给它。
		originContext context.Context
		// HTTP 方法，如 GET、POST 等。
		method string
		// 请求的 URL 地址。
		url string
		// 当前要发送的 HTTP 请求对象，Hook 可在发送前替换其上下文等属性。
		request *http.Request
		// 操作开始时间，在创建 HookContext 时记录。
		startTime time.Time
		// 操作结束时间，在 SetResult 时记录；请求未完成前为零值。
		endTime time.Time
		// 原始操作返回的错误。
		originError error
		// 原始操作返回的响应。
		originResult *http.Response
		// hookMap 存储 Hook 链内共享的键值对数据，可被多个 goroutine 安全访问。
		hookMap sync.Map
	}
)

// NewHookContext 创建新的 HookContext。
//
// startTime 会在创建时记录；endTime、originResult 和 originError 需要在请求完成后通过 SetResult 写入。
//
// 参数：
//   - ctx: 原始上下文对象，context.Context 接口方法会委托给它。
//   - method: HTTP 方法，如 GET、POST 等。
//   - url: 请求的 URL 地址。
//   - request: 当前要发送的 HTTP 请求对象。
//
// 返回：
//   - *HookContext: 初始化完成的 Hook 上下文。
func NewHookContext(ctx context.Context, method, url string, request *http.Request) *HookContext {
	return &HookContext{
		originContext: ctx,
		method:        method,
		url:           url,
		request:       request,
		startTime:     time.Now(),
	}
}

// SetResult 记录 HTTP 请求的原始响应、错误和结束时间。
//
// 参数：
//   - response: HTTP 返回对象，可为 nil。
//   - err: HTTP 请求返回的原始错误；没有错误时为 nil。
func (h *HookContext) SetResult(response *http.Response, err error) {
	h.originResult = response
	h.originError = err
	h.endTime = time.Now()
}

// StartTime 返回操作开始时间。
//
// 参数：无。
//
// 返回：
//   - time.Time: 创建 HookContext 时记录的开始时间。
func (h *HookContext) StartTime() time.Time {
	return h.startTime
}

// EndTime 返回操作结束时间。
//
// 参数：无。
//
// 返回：
//   - time.Time: SetResult 记录的结束时间；请求未完成前为零值。
func (h *HookContext) EndTime() time.Time {
	return h.endTime
}

// Duration 返回操作持续时间。
//
// 请求未调用 SetResult 时，endTime 为零值，返回值会反映零值时间与 startTime 的差值。
//
// 参数：无。
//
// 返回：
//   - time.Duration: endTime 与 startTime 的时间差。
func (h *HookContext) Duration() time.Duration {
	return h.endTime.Sub(h.startTime)
}

// Method 返回 HTTP 方法。
//
// 参数：无。
//
// 返回：
//   - string: 当前 HookContext 记录的 HTTP 方法。
func (h *HookContext) Method() string {
	return h.method
}

// Url 返回请求 URL 字符串。
//
// 参数：无。
//
// 返回：
//   - string: 当前 HookContext 记录的请求 URL。
func (h *HookContext) Url() string {
	return h.url
}

// Request 返回当前 HTTP 请求对象。
//
// Hook 可在 Before 阶段替换请求上下文，后续 Do 会使用这里返回的请求对象发送。
//
// 参数：无。
//
// 返回：
//   - *http.Request: 当前要发送的 HTTP 请求对象。
func (h *HookContext) Request() *http.Request {
	return h.request
}

// OriginError 返回原始 HTTP 请求错误。
//
// 参数：无。
//
// 返回：
//   - error: SetResult 写入的原始错误；请求成功或尚未写入结果时为 nil。
func (h *HookContext) OriginError() error {
	return h.originError
}

// OriginResult 返回原始 HTTP 响应。
//
// 参数：无。
//
// 返回：
//   - any: SetResult 写入的 *http.Response；尚未写入结果时为 nil。
func (h *HookContext) OriginResult() any {
	return h.originResult
}

// GetHookValue 获取 Hook 链共享数据。
//
// 参数：
//   - key: 要获取的键名。
//
// 返回：
//   - interface{}: 与 key 关联的值。
//   - bool: key 存在时返回 true，否则返回 false。
func (h *HookContext) GetHookValue(key string) (interface{}, bool) {
	return h.hookMap.Load(key)
}

// SetHookValue 写入 Hook 链共享数据。
//
// 参数：
//   - key: 要设置的键名。
//   - value: 要存储的值。
func (h *HookContext) SetHookValue(key string, value interface{}) {
	h.hookMap.Store(key, value)
}

// Deadline 实现 context.Context 接口并返回原始上下文的截止时间。
//
// 参数：无。
//
// 返回：
//   - deadline: 原始上下文的截止时间。
//   - ok: 原始上下文设置了截止时间时返回 true，否则返回 false。
func (h *HookContext) Deadline() (deadline time.Time, ok bool) {
	return h.originContext.Deadline()
}

// Done 实现 context.Context 接口并返回原始上下文的取消通知通道。
//
// 参数：无。
//
// 返回：
//   - <-chan struct{}: 原始上下文被取消或超时时关闭的通道。
func (h *HookContext) Done() <-chan struct{} {
	return h.originContext.Done()
}

// Err 实现 context.Context 接口并返回原始上下文的取消原因。
//
// 参数：无。
//
// 返回：
//   - error: 原始上下文被取消或超时时返回对应错误；未取消时返回 nil。
func (h *HookContext) Err() error {
	return h.originContext.Err()
}

// Value 实现 context.Context 接口并从原始上下文读取值。
//
// HookContext 的共享 Hook 数据不通过 Value 暴露，调用方应使用 GetHookValue 读取。
//
// 参数：
//   - key: 要从原始上下文读取的键。
//
// 返回：
//   - interface{}: 原始上下文中与 key 关联的值；不存在时返回 nil。
func (h *HookContext) Value(key interface{}) interface{} {
	return h.originContext.Value(key)
}

type (
	// Hook 定义 HTTP 请求执行前后的扩展点。
	//
	// Before 在请求发送前执行，可修改 HookContext 中的请求对象或返回错误中止请求；
	// After 在请求完成后执行，适合记录日志、指标和 trace 等观察逻辑。
	Hook interface {
		// Before 在操作执行前调用。
		//
		// 参数：
		//   - ctx: 钩子上下文，包含请求和共享 Hook 数据。
		//
		// 返回：
		//   - error: 返回非 nil 错误时请求会被中止并向调用方返回该错误。
		Before(ctx *HookContext) error

		// After 在操作执行后调用。
		//
		// 参数：
		//   - ctx: 钩子上下文，包含请求、原始响应、原始错误和共享 Hook 数据。
		//
		// 返回：
		//   - error: Hook 执行失败时返回错误；当前 client.Do 会忽略该错误。
		After(ctx *HookContext) error
	}
)

// HookManager 管理多个 Hook 的执行顺序。
//
// Before 按注册顺序执行，After 按逆序执行，便于为一次请求构建成对包裹的 Hook 链。
type HookManager struct {
	// hooks 存储注册的所有 Hook，执行顺序由 AddHook 的调用顺序决定。
	hooks []Hook
}

// NewHookManager 创建新的 HookManager。
//
// 参数：无。
//
// 返回：
//   - *HookManager: 未注册任何 Hook 的 HookManager。
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]Hook, 0),
	}
}

// AddHook 添加一个 Hook 到执行链尾部。
//
// 参数：
//   - hook: 要添加的 Hook 实现；当前实现不会过滤 nil。
func (m *HookManager) AddHook(hook Hook) {
	m.hooks = append(m.hooks, hook)
}

// Before 实现 Hook 接口，按注册顺序执行所有 Hook 的 Before 方法。
//
// 参数：
//   - ctx: 钩子上下文，包含请求和共享 Hook 数据。
//
// 返回：
//   - error: 任一 Hook 的 Before 返回错误时立即停止并返回该错误。
func (m *HookManager) Before(ctx *HookContext) error {
	for _, hook := range m.hooks {
		if err := hook.Before(ctx); err != nil {
			return err
		}
	}
	return nil
}

// After 实现 Hook 接口，按注册逆序执行所有 Hook 的 After 方法。
//
// 参数：
//   - ctx: 钩子上下文，包含请求、原始响应、原始错误和共享 Hook 数据。
//
// 返回：
//   - error: 任一 Hook 的 After 返回错误时立即停止并返回该错误。
func (m *HookManager) After(ctx *HookContext) error {
	for i := len(m.hooks) - 1; i >= 0; i-- {
		if err := m.hooks[i].After(ctx); err != nil {
			return err
		}
	}
	return nil
}
