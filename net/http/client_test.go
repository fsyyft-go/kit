// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"net/http"
	"testing"

	kitlog "github.com/fsyyft-go/kit/log"
	kitnet "github.com/fsyyft-go/kit/net"
)

func TestSimpleRequest(t *testing.T) {
	if !kitnet.TestNetwork() {
		t.Skipf("环境变量缺少 %s，跳过测试", kitnet.EnvTestNetwork)
	}
	logger, _ := kitlog.NewStdLogger("")
	logger.SetLevel(kitlog.DebugLevel)
	client := NewClient(
		WithName("test-simple-request"),
		WithProxy(nil),
		WithTraceEnable(true),
		WithLogger(logger))

	cases := []struct {
		url    string
		method string
	}{
		{"http://baidu.com", http.MethodHead},
		{"http://vv.video.qq.com/checktime?otype=json", http.MethodGet},
	}

	for _, c := range cases {
		var err error
		var resp *http.Response

		switch c.method {
		case http.MethodHead:
			resp, err = client.Head(t.Context(), c.url)
		case http.MethodGet:
			resp, err = client.Get(t.Context(), c.url)
		default:
			t.Fatalf("不支持的请求方法: %s", c.method)
		}

		if nil != err {
			t.Error(err)
		}
		t.Log(resp)
	}
}
