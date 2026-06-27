// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"crypto/tls"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kitlog "github.com/fsyyft-go/kit/log"
)

type contextKey string

type orderedHook struct {
	name       string
	events     *[]string
	beforeErr  error
	afterErr   error
	beforeSeen *HookContext
	afterSeen  *HookContext
}

// Before 记录 HookManager 前置执行顺序并按配置返回错误。
//
// 该辅助方法用于验证 HookManager 按添加顺序调用 Before，并在遇到错误时停止后续 Hook。
//
// 参数：
//   - ctx: HTTP 请求生命周期中的 Hook 上下文。
//
// 返回：
//   - error: 预设的前置错误；未设置时为 nil。
func (h *orderedHook) Before(ctx *HookContext) error {
	*h.events = append(*h.events, "before:"+h.name)
	h.beforeSeen = ctx
	return h.beforeErr
}

// After 记录 HookManager 后置执行顺序并按配置返回错误。
//
// 该辅助方法用于验证 HookManager 按添加顺序的逆序调用 After，并在遇到错误时停止后续 Hook。
//
// 参数：
//   - ctx: HTTP 请求生命周期中的 Hook 上下文。
//
// 返回：
//   - error: 预设的后置错误；未设置时为 nil。
func (h *orderedHook) After(ctx *HookContext) error {
	*h.events = append(*h.events, "after:"+h.name)
	h.afterSeen = ctx
	return h.afterErr
}

// TestHookContext_ContextAndValues 验证 HookContext 保存请求元数据、结果与原始 context 行为。
//
// 该测试覆盖 HookContext 的方法、URL、请求引用、开始结束时间、持续时间、结果记录、hook value
// 以及 Deadline、Done、Err、Value 对原始 context 的代理语义。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestHookContext_ContextAndValues(t *testing.T) {
	key := contextKey("trace-id")
	deadline := time.Now().Add(time.Hour)
	originCtx, cancel := context.WithDeadline(context.WithValue(context.Background(), key, "abc-123"), deadline)
	t.Cleanup(cancel)
	req := httptest.NewRequest(stdhttp.MethodPost, "http://example.test/path", nil).WithContext(originCtx)

	hookCtx := NewHookContext(originCtx, req.Method, req.URL.String(), req)
	hookCtx.SetHookValue("phase", "before")
	resp := &stdhttp.Response{StatusCode: stdhttp.StatusAccepted}
	errResult := errors.New("transport failed")
	hookCtx.SetResult(resp, errResult)

	gotDeadline, ok := hookCtx.Deadline()
	assert.True(t, ok)
	assert.Equal(t, deadline, gotDeadline)
	assert.Equal(t, originCtx.Done(), hookCtx.Done())
	assert.NoError(t, hookCtx.Err())
	assert.Equal(t, "abc-123", hookCtx.Value(key))
	assert.Equal(t, stdhttp.MethodPost, hookCtx.Method())
	assert.Equal(t, "http://example.test/path", hookCtx.Url())
	assert.Same(t, req, hookCtx.Request())
	assert.False(t, hookCtx.StartTime().IsZero())
	assert.False(t, hookCtx.EndTime().IsZero())
	assert.GreaterOrEqual(t, hookCtx.Duration(), time.Duration(0))
	assert.ErrorIs(t, hookCtx.OriginError(), errResult)
	assert.Same(t, resp, hookCtx.OriginResult())
	gotValue, ok := hookCtx.GetHookValue("phase")
	assert.True(t, ok)
	assert.Equal(t, "before", gotValue)
	_, ok = hookCtx.GetHookValue("missing")
	assert.False(t, ok)
}

// TestHookContext_CanceledOrigin 验证 HookContext 代理原始 context 的取消状态。
//
// 该测试使用已取消的原始 context 构造 HookContext，确保 Err 能反映 context.Canceled。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestHookContext_CanceledOrigin(t *testing.T) {
	originCtx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(stdhttp.MethodGet, "http://example.test", nil).WithContext(originCtx)

	hookCtx := NewHookContext(originCtx, req.Method, req.URL.String(), req)

	assert.ErrorIs(t, hookCtx.Err(), context.Canceled)
}

// TestHookManager_OrderAndErrors 验证 HookManager 的执行顺序与错误短路语义。
//
// 该测试通过表驱动用例覆盖 Before 顺序执行、Before 遇错停止、After 逆序执行以及 After 遇错停止，
// 确保多个 Hook 组合时生命周期行为稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHookManager_OrderAndErrors(t *testing.T) {
	errBefore := errors.New("before failed")
	errAfter := errors.New("after failed")

	tests := []struct {
		name        string
		description string
		setup       func(events *[]string) (*HookManager, error)
		giveCall    func(manager *HookManager, ctx *HookContext) error
		wantErrIs   error
		wantEvents  []string
	}{
		{
			name:        "success/before-order",
			description: "验证 HookManager.Before 按 Hook 添加顺序执行所有前置逻辑。",
			setup: func(events *[]string) (*HookManager, error) {
				manager := NewHookManager()
				manager.AddHook(&orderedHook{name: "one", events: events})
				manager.AddHook(&orderedHook{name: "two", events: events})
				return manager, nil
			},
			giveCall: func(manager *HookManager, ctx *HookContext) error {
				return manager.Before(ctx)
			},
			wantEvents: []string{"before:one", "before:two"},
		},
		{
			name:        "error/before-short-circuit",
			description: "验证 HookManager.Before 在某个 Hook 返回错误后不再执行后续 Hook。",
			setup: func(events *[]string) (*HookManager, error) {
				manager := NewHookManager()
				manager.AddHook(&orderedHook{name: "one", events: events, beforeErr: errBefore})
				manager.AddHook(&orderedHook{name: "two", events: events})
				return manager, nil
			},
			giveCall: func(manager *HookManager, ctx *HookContext) error {
				return manager.Before(ctx)
			},
			wantErrIs:  errBefore,
			wantEvents: []string{"before:one"},
		},
		{
			name:        "success/after-reverse-order",
			description: "验证 HookManager.After 按 Hook 添加顺序的逆序执行所有后置逻辑。",
			setup: func(events *[]string) (*HookManager, error) {
				manager := NewHookManager()
				manager.AddHook(&orderedHook{name: "one", events: events})
				manager.AddHook(&orderedHook{name: "two", events: events})
				return manager, nil
			},
			giveCall: func(manager *HookManager, ctx *HookContext) error {
				return manager.After(ctx)
			},
			wantEvents: []string{"after:two", "after:one"},
		},
		{
			name:        "error/after-short-circuit",
			description: "验证 HookManager.After 在逆序执行中遇到错误后不再执行剩余 Hook。",
			setup: func(events *[]string) (*HookManager, error) {
				manager := NewHookManager()
				manager.AddHook(&orderedHook{name: "one", events: events})
				manager.AddHook(&orderedHook{name: "two", events: events, afterErr: errAfter})
				return manager, nil
			},
			giveCall: func(manager *HookManager, ctx *HookContext) error {
				return manager.After(ctx)
			},
			wantErrIs:  errAfter,
			wantEvents: []string{"after:two"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			events := make([]string, 0)
			manager, err := tt.setup(&events)
			require.NoError(t, err)
			req := httptest.NewRequest(stdhttp.MethodGet, "http://example.test", nil)
			hookCtx := NewHookContext(t.Context(), req.Method, req.URL.String(), req)

			err = tt.giveCall(manager, hookCtx)

			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantEvents, events)
		})
	}
}

// TestTraceInfo_Durations 验证 traceInfo 的耗时计算语义。
//
// 该测试通过表驱动用例覆盖 DNS、连接和 TLS 耗时在时间点完整与缺失时的返回值。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestTraceInfo_Durations(t *testing.T) {
	base := time.Now()
	tests := []struct {
		name        string
		description string
		giveInfo    traceInfo
		assert      func(t *testing.T, info traceInfo)
	}{
		{
			name:        "success/all-durations",
			description: "验证 DNS、连接和 TLS 时间点完整时返回对应耗时。",
			giveInfo: traceInfo{
				TimeDNSStart:          base,
				TimeDNSDone:           base.Add(10 * time.Millisecond),
				TimeGetConn:           base.Add(20 * time.Millisecond),
				TimeGotConn:           base.Add(45 * time.Millisecond),
				TimeTLSHandshakeStart: base.Add(50 * time.Millisecond),
				TimeTLSHandshakeDone:  base.Add(80 * time.Millisecond),
			},
			assert: func(t *testing.T, info traceInfo) {
				assert.Equal(t, 10*time.Millisecond, info.DNSUseTime())
				assert.Equal(t, 25*time.Millisecond, info.ConnectUseTime())
				assert.Equal(t, 30*time.Millisecond, info.TLSUseTime())
			},
		},
		{
			name:        "boundary/missing-start-or-end",
			description: "验证任一关键时间点缺失时对应耗时返回零值。",
			giveInfo: traceInfo{
				TimeDNSDone:           base.Add(10 * time.Millisecond),
				TimeGetConn:           base.Add(20 * time.Millisecond),
				TimeTLSHandshakeStart: base.Add(50 * time.Millisecond),
			},
			assert: func(t *testing.T, info traceInfo) {
				assert.Zero(t, info.DNSUseTime())
				assert.Zero(t, info.ConnectUseTime())
				assert.Zero(t, info.TLSUseTime())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			tt.assert(t, tt.giveInfo)
		})
	}
}

// TestTraceHook_BeforeAndAfter 验证 traceHook 注入 httptrace 并记录 trace 信息。
//
// 该测试直接调用 trace hook，确认 Before 会在 HookContext 中存入 traceInfo 并替换请求 context，
// 同时通过触发 httptrace 回调验证关键字段记录；After 在已有 traceInfo 时应稳定返回 nil。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestTraceHook_BeforeAndAfter(t *testing.T) {
	logger, err := kitlog.NewStdLogger("")
	require.NoError(t, err)
	hook := NewTraceHook(logger)
	req := httptest.NewRequest(stdhttp.MethodGet, "https://example.test", nil)
	hookCtx := NewHookContext(t.Context(), req.Method, req.URL.String(), req)

	err = hook.Before(hookCtx)

	require.NoError(t, err)
	stored, ok := hookCtx.GetHookValue("traceInfo")
	require.True(t, ok)
	info := stored.(*traceInfo)
	assert.Empty(t, info.TimeConnectStart)
	clientTrace := httptrace.ContextClientTrace(hookCtx.Request().Context())
	require.NotNil(t, clientTrace)

	clientTrace.GetConn("example.test:443")
	clientTrace.GotConn(httptrace.GotConnInfo{Reused: true})
	clientTrace.PutIdleConn(nil)
	clientTrace.GotFirstResponseByte()
	clientTrace.Got100Continue()
	require.NoError(t, clientTrace.Got1xxResponse(103, textproto.MIMEHeader{"Link": {"</style.css>"}}))
	clientTrace.DNSStart(httptrace.DNSStartInfo{Host: "example.test"})
	clientTrace.DNSDone(httptrace.DNSDoneInfo{})
	clientTrace.ConnectStart("tcp", "127.0.0.1:443")
	clientTrace.ConnectDone("tcp", "127.0.0.1:443", nil)
	clientTrace.TLSHandshakeStart()
	clientTrace.TLSHandshakeDone(tls.ConnectionState{}, nil)
	clientTrace.WroteHeaderField("X-Test", []string{"ok"})
	clientTrace.WroteHeaders()
	clientTrace.Wait100Continue()
	clientTrace.WroteRequest(httptrace.WroteRequestInfo{})

	assert.Equal(t, "example.test:443", info.HostPortGetConn)
	assert.NotNil(t, info.GotConnInfo)
	assert.False(t, info.TimePutIdleConn.IsZero())
	assert.False(t, info.TimeGotFirstResponseByte.IsZero())
	assert.False(t, info.TimeGot100Continue.IsZero())
	assert.Equal(t, 103, info.CodeGot1xxResponse)
	assert.Equal(t, "example.test", info.DNSStartInfo.Host)
	assert.Equal(t, []string{"tcp"}, info.NetworkConnectStart)
	assert.Equal(t, []string{"127.0.0.1:443"}, info.AddrConnectStart)
	assert.Equal(t, []string{"tcp"}, info.NetworkConnectDone)
	assert.Equal(t, []string{"127.0.0.1:443"}, info.AddrConnectDone)
	assert.Len(t, info.ErrorConnectDone, 1)
	assert.NoError(t, info.ErrorConnectDone[0])
	assert.False(t, info.TimeTLSHandshakeStart.IsZero())
	assert.False(t, info.TimeTLSHandshakeDone.IsZero())
	assert.False(t, info.TimeWroteHeaders.IsZero())
	assert.False(t, info.TimeWait100Continue.IsZero())
	assert.Len(t, info.TimeWroteRequest, 1)
	assert.NoError(t, hook.After(hookCtx))
}

// TestBuiltInHooks_BasicBehavior 验证内置慢请求与错误日志 Hook 的基础行为。
//
// 该测试不依赖日志输出内容，仅验证构造函数保存阈值、Before 无副作用、After 在慢请求或错误请求场景下稳定返回 nil。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBuiltInHooks_BasicBehavior(t *testing.T) {
	logger, err := kitlog.NewStdLogger("")
	require.NoError(t, err)
	req := httptest.NewRequest(stdhttp.MethodGet, "http://example.test", nil)

	tests := []struct {
		name        string
		description string
		giveHook    Hook
		setup       func(ctx *HookContext)
		assert      func(t *testing.T, hook Hook)
	}{
		{
			name:        "success/slow-hook",
			description: "验证 slowHook 保存慢请求阈值且对超过阈值的请求稳定返回 nil。",
			giveHook:    NewSlowHook(logger, time.Nanosecond),
			setup: func(ctx *HookContext) {
				ctx.startTime = time.Now().Add(-time.Millisecond)
				ctx.endTime = time.Now()
			},
			assert: func(t *testing.T, hook Hook) {
				slow := hook.(*slowHook)
				assert.Equal(t, time.Nanosecond, slow.threshold)
			},
		},
		{
			name:        "success/log-error-hook",
			description: "验证 logErrorHook 对包含原始错误的上下文稳定返回 nil。",
			giveHook:    NewLogErrorHook(logger),
			setup: func(ctx *HookContext) {
				ctx.SetResult(nil, errors.New("request failed"))
			},
			assert: func(t *testing.T, hook Hook) {
				assert.IsType(t, &logErrorHook{}, hook)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			hookCtx := NewHookContext(t.Context(), req.Method, req.URL.String(), req)
			tt.setup(hookCtx)

			assert.NoError(t, tt.giveHook.Before(hookCtx))
			assert.NoError(t, tt.giveHook.After(hookCtx))
			tt.assert(t, tt.giveHook)
		})
	}
}
