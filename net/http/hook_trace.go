// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"crypto/tls"
	"net/http/httptrace"
	"net/textproto"
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
)

type (
	traceHook struct {
		logger kitlog.Logger
	}
	traceInfo struct {
		TimeGetConn     time.Time // TimeGetConn 准备获取连接的时间，在 ClientTrace.GetConn 事件中产生。
		HostPortGetConn string    // HostPortGetConn 目标服务器的信息，格式为 host:port，在 ClientTrace.GetConn 事件中产生。

		TimeGotConn time.Time              // TimeGotConn 获取到连接的时间，在 ClientTrace.GotConn 事件中产生。
		GotConnInfo *httptrace.GotConnInfo //GotConnInfo 获取到的连接信息，在 ClientTrace.GotConn 事件中产生。

		TimePutIdleConn  time.Time // TimePutIdleConn 将连接放回连接池的时间，在 ClientTrace.PutIdleConn 事件中产生。
		ErrorPutIdleConn error     // ErrorPutIdleConn 连接放回连接池的错误信息，在 ClientTrace.PutIdleConn 事件中产生。

		TimeGotFirstResponseByte time.Time // TimeGotFirstResponseByte 获取到第一个返回字节的时间，在 ClientTrace.GotFirstResponseByte 事件中产生。

		TimeGot100Continue time.Time // TimeGot100Continue 获取到 100 状态码的时间，在 ClientTrace.Got100Continue 事件中产生。

		TimeGot1xxResponse   time.Time            // TimeGot1xxResponse 获取到 1xx 状态码的时间，在 ClientTrace.Got1xxResponse 事件中产生。
		CodeGot1xxResponse   int                  // CodeGot1xxResponse 获取到 1xx 状态码对应的具体状态码，在 ClientTrace.Got1xxResponse 事件中产生。
		HeaderGot1xxResponse textproto.MIMEHeader // HeaderGot1xxResponse 获取到 1xx 状态码对应的头部信息，在 ClientTrace.Got1xxResponse 事件中产生。

		TimeDNSStart time.Time              // TimeDNSStart 准备进行 DNS 解析的时间，在 ClientTrace.DNSStart 事件中产生。
		DNSStartInfo httptrace.DNSStartInfo // DNSStartInfo DNS 请求信息，在 ClientTrace.DNSStart 事件中产生。

		TimeDNSDone time.Time             // TimeDNSDone 完成 DNS 解析的时间，在 ClientTrace.DNSDone 事件中产生。
		DNSDoneInfo httptrace.DNSDoneInfo // DNSDoneInfo DNS 的结果信息，在 ClientTrace.DNSDone 事件中产生。

		TimeConnectStart    []time.Time // TimeConnectStart 新的连接开始之前的时间，在 ClientTrace.ConnectStart 事件中产生。
		NetworkConnectStart []string    // NetworkConnectStart 新的连接开始之前的网络类型信息，在 ClientTrace.ConnectStart 事件中产生。
		AddrConnectStart    []string    // AddrConnectStart 新的连接开始之前的地址信息，在 ClientTrace.ConnectStart 事件中产生。

		TimeConnectDone    []time.Time // TimeConnectDone 新的连接完成的时间，在 ClientTrace.ConnectDone 事件中产生。
		NetworkConnectDone []string    // NetworkConnectDone 新的连接完成的网络类型信息，在 ClientTrace.ConnectDone 事件中产生。
		AddrConnectDone    []string    // AddrConnectDone 新的连接完成的地址信息，在 ClientTrace.ConnectDone 事件中产生。
		ErrorConnectDone   []error     // ErrorConnectDone 新的连接完成的错误信息，在 ClientTrace.ConnectDone 事件中产生。

		TimeTLSHandshakeStart time.Time // TimeTLSHandshakeStart 开始进行 TLS 握手的时间，在 ClientTrace.TLSHandshakeStart 事件中产生。
		TimeTLSHandshakeDone  time.Time // TimeTLSHandshakeDone 完成 TLS 握手的时间，在 ClientTrace.TLSHandshakeDone 事件中产生。

		TimeWroteHeaders    time.Time   // TimeWroteHeaders 在传输已写入所有请求标头之后的时间，在 ClientTrace.WroteHeaders 事件中产生。
		TimeWait100Continue time.Time   // TimeWait100Continue 在传输已写入所有请求标头之后的时间，在 ClientTrace.Wait100Continue 事件中产生。
		TimeWroteRequest    []time.Time // TimeWroteRequest 在传输写入请求信息的时间，如果有重试，可能出现多次，在 ClientTrace.WroteRequest 事件中产生。
	}
)

func (i *traceInfo) DNSUseTime() time.Duration {
	if i.TimeDNSDone.IsZero() || i.TimeDNSStart.IsZero() {
		return 0
	}
	return i.TimeDNSDone.Sub(i.TimeDNSStart)
}

func (i *traceInfo) ConnectUseTime() time.Duration {
	if i.TimeGetConn.IsZero() || i.TimeGotConn.IsZero() {
		return 0
	}
	return i.TimeGotConn.Sub(i.TimeGetConn)
}

func (i *traceInfo) TLSUseTime() time.Duration {
	if i.TimeTLSHandshakeStart.IsZero() || i.TimeTLSHandshakeDone.IsZero() {
		return 0
	}
	return i.TimeTLSHandshakeDone.Sub(i.TimeTLSHandshakeStart)
}
func (h *traceHook) Before(ctx *HookContext) error {
	traceInfo := &traceInfo{
		TimeConnectStart:    make([]time.Time, 0),
		NetworkConnectStart: make([]string, 0),
		AddrConnectStart:    make([]string, 0),
		TimeConnectDone:     make([]time.Time, 0),
		NetworkConnectDone:  make([]string, 0),
		AddrConnectDone:     make([]string, 0),
		ErrorConnectDone:    make([]error, 0),
		TimeWroteRequest:    make([]time.Time, 0),
	}
	ctx.SetHookValue("traceInfo", traceInfo)

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			// 在创建连接或从空闲池中检索连接之前，将调用 GetConn。 hostPort 是目标或代理的 "host:port" 。
			// 即使已经有空闲的缓存连接可用，也会调用 GetConn。
			traceInfo.TimeGetConn = time.Now()
			traceInfo.HostPortGetConn = hostPort
		},
		GotConn: func(info httptrace.GotConnInfo) {
			// 获得成功的连接后，将调用 GotConn。
			// 没有连接失败的钩子。可以使用 Transport.RoundTrip 中的错误代替。
			traceInfo.TimeGotConn = time.Now()
			traceInfo.GotConnInfo = &info
		},
		PutIdleConn: func(err error) {
			// 当连接返回到空闲池时，将调用 PutIdleConn。
			// 如果 err 为 nil，则连接已成功返回到空闲池。
			// 如果 err 不为 nil，则说明原因。
			// 如果通过 Transport.DisableKeepAlives 禁用了连接重用，则不会调用 PutIdleConn。
			// 在调用者的 Response.Body.Close 调用返回之前，将调用 PutIdleConn。
			// 对于 HTTP2，当前未使用此钩子。
			traceInfo.TimePutIdleConn = time.Now()
			traceInfo.ErrorPutIdleConn = err
		},
		GotFirstResponseByte: func() {
			// 当响应头的第一个字节可用时，将调用 GotFirstResponseByte。
			traceInfo.TimeGotFirstResponseByte = time.Now()
		},
		Got100Continue: func() {
			// 如果服务器回复“100 Continue”响应，则会调用 Got100Continue。
			traceInfo.TimeGot100Continue = time.Now()
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			// 在最终的非 1xx 响应之前，为每个返回的 1xx 信息响应头调用 Got1xxResponse。
			// 即使还定义了 Got100Continue，也会对“100 Continue”响应调用 Got1xxResponse。
			// 如果返回错误，则客户端请求将使用该错误值中止。
			traceInfo.TimeGot1xxResponse = time.Now()
			traceInfo.CodeGot1xxResponse = code
			traceInfo.HeaderGot1xxResponse = header
			return nil
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			// DNS 查找开始时，将调用 DNSStart。
			traceInfo.TimeDNSStart = time.Now()
			traceInfo.DNSStartInfo = info
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			// DNS 查找结束时，将调用 DNSDone。
			traceInfo.TimeDNSDone = time.Now()
			traceInfo.DNSDoneInfo = info
		},
		ConnectStart: func(network, addr string) {
			// 新连接的拨号开始时，将调用 ConnectStart。如果启用了 net.Dialer.DualStack（IPv6 的“Happy Eyeballs”）支持，则可能需要多次调用。
			traceInfo.TimeConnectStart = append(traceInfo.TimeConnectStart, time.Now())
			traceInfo.NetworkConnectStart = append(traceInfo.NetworkConnectStart, network)
			traceInfo.AddrConnectStart = append(traceInfo.AddrConnectStart, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			// 新连接的拨号完成后，将调用 ConnectDone。
			// 提供的错误指示连接是否成功完成。如果启用了 net.Dialer.DualStack（IPv6 的“Happy Eyeballs”）支持，则可能需要多次调用。
			traceInfo.TimeConnectDone = append(traceInfo.TimeConnectDone, time.Now())
			traceInfo.NetworkConnectDone = append(traceInfo.NetworkConnectDone, network)
			traceInfo.AddrConnectDone = append(traceInfo.AddrConnectDone, addr)
			traceInfo.ErrorConnectDone = append(traceInfo.ErrorConnectDone, err)
		},
		TLSHandshakeStart: func() {
			// TLS 握手开始时，将调用 TLSHandshakeStart。
			// 如果通过 HTTP 代理连接到 HTTPS 站点时，握手在代理处理 CONNECT 请求之后发生。
			traceInfo.TimeTLSHandshakeStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			// TLS 握手后，将以成功握手的连接状态或握手失败时显示非 nil 错误的方式调用 TLSHandshakeDone。
			traceInfo.TimeTLSHandshakeDone = time.Now()
		},
		WroteHeaderField: func(key string, value []string) {
			// 在传输已写入每个请求标头之后，将调用 WroteHeaderField。
			// 在进行此调用时，这些值可能已被缓冲并且尚未写入网络。
		},
		WroteHeaders: func() {
			// 在传输已写入所有请求标头之后，将调用 WroteHeaders。
			traceInfo.TimeWroteHeaders = time.Now()
		},
		Wait100Continue: func() {
			// 如果请求指定为“期望：100-continue”，并且传输已写入请求标头，但在写入请求正文之前，服务器正在等待服务器的“100 Continue”，则调用 Wait100Continue。
			traceInfo.TimeWait100Continue = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			// 写入请求和任何正文的结果将调用 WroteRequest。
			// 在重试请求的情况下，可以多次调用它。
			traceInfo.TimeWroteRequest = append(traceInfo.TimeWroteRequest, time.Now())
		},
	}
	traceContext := httptrace.WithClientTrace(ctx.request.Context(), trace)
	// 注意：这里需要修改 Request，这个 Hook 尽量放在 Hook 列表的前面。
	ctx.request = ctx.request.WithContext(traceContext)

	return nil
}

func (h *traceHook) After(ctx *HookContext) error {
	if ti, ok := ctx.GetHookValue("traceInfo"); ok {
		l := h.logger
		if i, ok := ti.(*traceInfo); ok {
			l = l.WithField("dnsUseTime", i.DNSUseTime())
			l = l.WithField("connectUseTime", i.ConnectUseTime())
			l = l.WithField("tlsUseTime", i.TLSUseTime())
			if nil != i.DNSDoneInfo.Addrs {
				l = l.WithField("remoteAddr", i.DNSDoneInfo.Addrs[0].String())
			}
		}
		l.Debug("")

	}
	return nil
}

func NewTraceHook(logger kitlog.Logger) *traceHook {
	h := &traceHook{
		logger: logger,
	}
	return h
}
